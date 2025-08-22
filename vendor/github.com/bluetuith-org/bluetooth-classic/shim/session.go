//go:build !linux

package shim

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"path"
	"sync"
	"sync/atomic"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/puzpuzpuz/xsync/v3"

	ac "github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/api/config"
	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	sstore "github.com/bluetuith-org/bluetooth-classic/api/helpers/sessionstore"
	"github.com/bluetuith-org/bluetooth-classic/api/platforminfo"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/commands"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/events"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/serde"
)

// ShimSession describes a connected session with a running shim RPC server.
//
//revive:disable
type ShimSession struct {
	features   *ac.FeatureSet
	authorizer bluetooth.SessionAuthorizer

	conn net.Conn

	listenerEvents chan []byte
	sessionClosed  atomic.Bool

	cancel context.CancelFunc

	id         *xsync.Counter
	requestMap *xsync.MapOf[int64, chan commands.CommandResponse]

	store sstore.SessionStore

	sync.Mutex
}

//revive:enable

const socketName = "hd.sock"

// Start attempts to initialize a session with the system's Bluetooth daemon or service.
// Upon complete initialization, it returns the session descriptor, and capabilities of
// the application.
func (s *ShimSession) Start(authHandler bluetooth.SessionAuthorizer, cfg config.Configuration) (*ac.FeatureSet, platforminfo.PlatformInfo, error) {
	var ce ac.Errors

	platform := platforminfo.NewPlatformInfo("")

	var initialized bool
	defer func() {
		if !initialized {
			s.Stop()
		}
	}()

	if authHandler == nil {
		authHandler = bluetooth.DefaultAuthorizer{}
	}
	s.authorizer = authHandler

	if cfg.SocketPath == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return nil, platform,
				fault.Wrap(err,
					fctx.With(context.Background(), "error_at", "socket-dir"),
					ftag.With(ftag.Internal),
					fmsg.With("Cannot find socket directory"),
				)
		}

		cfg.SocketPath = path.Join(dir, "haraltd", socketName)
	}

	ctx := s.reset(false)

	if err := s.startListener(ctx, cfg.SocketPath); err != nil {
		return nil, platform,
			fault.Wrap(errors.New(err.Error()),
				fctx.With(context.Background(), "error_at", "listener-shim"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot start listener on provided socket"),
			)
	}

	features, err := commands.GetFeatureFlags().ExecuteWith(s.executor)
	if err != nil {
		return nil, platform,
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "shim-features"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot get advertised features from shim"),
			)
	}

	platformInfo, err := commands.GetPlatformInfo().ExecuteWith(s.executor)
	if err != nil {
		return nil, platform,
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "shim-features"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot get advertised features from shim"),
			)
	}

	if err := s.refreshStore(); err != nil {
		return nil, platform,
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "shim-features"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot initialize the new session store"),
			)
	}

	initialized = true

	for _, absentFeatures := range features.AbsentFeatures() {
		ce.Append(ac.NewError(absentFeatures, errorkinds.ErrNotSupported))
	}

	s.features = ac.NewFeatureSet(features, ce)

	if s.features.Has(ac.FeatureSendFile, ac.FeatureReceiveFile) {
		if _, err := commands.RegisterAgent(commands.ObexAgent).ExecuteWith(s.executor); err != nil {
			ce.Append(ac.NewError(ac.FeatureSendFile, err))
		}
	}

	return s.features, platformInfo, nil
}

// Stop attempts to stop a session with the system's Bluetooth daemon or service.
func (s *ShimSession) Stop() error {
	if s.sessionClosed.Load() {
		return errorkinds.ErrSessionNotExist
	}

	s.reset(true)

	return nil
}

// Adapters returns a list of known adapters.
func (s *ShimSession) Adapters() ([]bluetooth.AdapterData, error) {
	return s.store.Adapters()
}

// Adapter returns a function call interface to invoke adapter related functions.
func (s *ShimSession) Adapter(adapterAddress bluetooth.MacAddress) bluetooth.Adapter {
	return &adapter{s, adapterAddress}
}

// Device returns a function call interface to invoke device related functions.
func (s *ShimSession) Device(deviceAddress bluetooth.MacAddress) bluetooth.Device {
	return &device{s, deviceAddress}
}

// Obex returns a function call interface to invoke obex related functions.
func (s *ShimSession) Obex(deviceAddress bluetooth.MacAddress) bluetooth.Obex {
	return &obex{s, deviceAddress}
}

// Network returns a function call interface to invoke network related functions.
func (s *ShimSession) Network(bluetooth.MacAddress) bluetooth.Network {
	return &network{}
}

// MediaPlayer returns a function call interface to invoke media player/control
// related functions on a device.
func (s *ShimSession) MediaPlayer(bluetooth.MacAddress) bluetooth.MediaPlayer {
	return &mediaPlayer{}
}

// adapter returns an adapter-related function call interface for internal use.
// This is used primarily to initialize adapter objects.
func (s *ShimSession) adapter() *adapter {
	return &adapter{}
}

// device returns an device-related function call interface for internal use.
// This is used primarily to initialize device objects.
func (s *ShimSession) device() *device {
	return &device{}
}

// refreshStore refreshes the global session store with adapter and device objects
// that are retrieved from the shim.
func (s *ShimSession) refreshStore() error {
	adapters, err := commands.GetAdapters().ExecuteWith(s.executor)
	if err != nil {
		return err
	}

	for _, adapter := range adapters {
		newAdapter, err := s.adapter().appendProperties(adapter)
		if err != nil {
			return err
		}
		s.store.AddAdapter(newAdapter)

		devices, err := commands.GetPairedDevices(adapter.Address).ExecuteWith(s.executor)
		if err != nil {
			return err
		}
		for _, device := range devices {
			newDevice, err := s.device().appendProperties(device, adapter)
			if err != nil {
				return err
			}

			s.store.AddDevice(newDevice)
		}
	}

	return nil
}

// startListener starts the socket and the listener.
func (s *ShimSession) startListener(ctx context.Context, socketpath string) error {
	socket, err := net.Dial("unix", socketpath)
	if err != nil {
		return err
	}

	s.conn = socket
	go s.listen(ctx)

	return nil
}

// listen listens to the socket for any incoming messages and events.
func (s *ShimSession) listen(ctx context.Context) {
	sendData := func(c chan commands.CommandResponse, m commands.CommandResponse) {
		select {
		case <-ctx.Done():
			close(c)
		case c <- m:
			close(c)
		default:
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		default:
		}

		if s.sessionClosed.Load() {
			return
		}

		scanner := bufio.NewScanner(s.conn)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			var response struct {
				commands.CommandResponse
				events.ServerEvent
			}

			if err := scanner.Err(); err != nil {
				s.handleListenerError(err, true)
				return
			}

			if err := serde.UnmarshalJson(scanner.Bytes(), &response); err != nil {
				s.handleListenerError(err, false)
			}

			if response.EventId > 0 {
				go s.handleListenerEvent(response.ServerEvent)
				continue
			}

			replyChan, ok := s.requestMap.LoadAndDelete(int64(response.RequestId))
			if ok {
				sendData(replyChan, response.CommandResponse)
			}
		}

		s.handleListenerError(scanner.Err(), true)
	}
}

// handleListenerEvent handles an event that was received from the socket (i.e listener).
func (s *ShimSession) handleListenerEvent(ev events.ServerEvent) {
	switch ev.EventId {
	case bluetooth.EventError:
		var genError error

		errorEvent, err := events.Unmarshal[errorkinds.GenericError](ev)
		if err != nil {
			genError = err
		} else {
			genError = errorEvent
		}

		bluetooth.ErrorEvents().PublishAdded(wrapError(genError))

	case bluetooth.EventAuthentication:
		authEvent, err := events.UnmarshalAuthEvent(ev)
		if err != nil {
			bluetooth.ErrorEvents().PublishAdded(wrapError(err))
			return
		}

		err = authEvent.CallAuthorizer(s.authorizer, func(authEvent events.AuthEventData, reply events.AuthReply, err error) {
			var response string

			if err == nil {
				response = reply.Reply
			}

			_, err = commands.AuthenticationReply(authEvent.AuthID, response).ExecuteWith(s.executor, (authEvent.TimeoutMs/1000)+2)
			if err != nil {
				bluetooth.ErrorEvents().PublishAdded(wrapError(err))
			}
		})
		if err != nil {
			bluetooth.ErrorEvents().PublishAdded(wrapError(err))
		}

	case bluetooth.EventAdapter:
		adapter, err := events.Unmarshal[bluetooth.AdapterData](ev)
		if err != nil {
			bluetooth.ErrorEvents().PublishAdded(wrapError(err))
			return
		}

		switch ev.EventAction {
		case bluetooth.EventActionAdded:
			bluetooth.AdapterEvents().PublishUpdated(adapter.AdapterEventData)

			adapter, err := commands.AdapterProperties(adapter.Address).ExecuteWith(s.executor)
			if err != nil {
				bluetooth.ErrorEvents().PublishAdded(wrapError(err))
				return
			}

			s.store.AddAdapter(adapter)

		case bluetooth.EventActionUpdated:
			updated, err := s.store.UpdateAdapter(adapter.Address, func(dd *bluetooth.AdapterData) error {
				return events.UnmarshalRawEvent(ev, &dd.AdapterEventData)
			})
			if err != nil {
				bluetooth.ErrorEvents().PublishAdded(wrapError(err))
				return
			}

			bluetooth.AdapterEvents().PublishUpdated(updated)

		case bluetooth.EventActionRemoved:
			bluetooth.AdapterEvents().PublishRemoved(adapter.AdapterEventData)
			s.store.RemoveAdapter(adapter.Address)
		}

	case bluetooth.EventDevice:
		device, err := events.Unmarshal[bluetooth.DeviceData](ev)
		if err != nil {
			bluetooth.ErrorEvents().PublishAdded(wrapError(err))
			return
		}

		switch ev.EventAction {
		case bluetooth.EventActionAdded:
			device.Type = bluetooth.DeviceTypeFromClass(device.Class)

			bluetooth.DeviceEvents().PublishAdded(device)
			s.store.AddDevice(device)

		case bluetooth.EventActionUpdated:
			updated, err := s.store.UpdateDevice(device.Address, func(dd *bluetooth.DeviceData) error {
				return events.UnmarshalRawEvent(ev, &dd.DeviceEventData)
			})
			if err != nil {
				bluetooth.ErrorEvents().PublishAdded(wrapError(err))
				return
			}

			bluetooth.DeviceEvents().PublishUpdated(updated)

		case bluetooth.EventActionRemoved:
			bluetooth.DeviceEvents().PublishRemoved(device.DeviceEventData)
			s.store.RemoveDevice(device.Address)
		}

	case bluetooth.EventObjectPush:
		filetransfer, err := events.Unmarshal[bluetooth.ObjectPushData](ev)
		if err != nil {
			bluetooth.ErrorEvents().PublishAdded(wrapError(err))
			return
		}

		switch ev.EventAction {
		case bluetooth.EventActionAdded:
			bluetooth.ObjectPushEvents().PublishAdded(filetransfer)

		case bluetooth.EventActionUpdated:
			bluetooth.ObjectPushEvents().PublishUpdated(filetransfer.ObjectPushEventData)

		case bluetooth.EventActionRemoved:
			bluetooth.ObjectPushEvents().PublishRemoved(filetransfer.ObjectPushEventData)
		}

	case bluetooth.EventMediaPlayer:
		mediaplayer, err := events.Unmarshal[bluetooth.MediaData](ev)
		if err != nil {
			bluetooth.ErrorEvents().PublishAdded(wrapError(err))
			return
		}

		switch ev.EventAction {
		case bluetooth.EventActionAdded:
			bluetooth.MediaEvents().PublishAdded(mediaplayer)

		case bluetooth.EventActionUpdated:
			bluetooth.MediaEvents().PublishUpdated(mediaplayer)

		case bluetooth.EventActionRemoved:
			bluetooth.MediaEvents().PublishRemoved(mediaplayer)
		}
	}
}

// handleListenerError handles any errors that occurred during listening from the socket.
// If the 'stop' parameter is specified, it means that an unrecoverable error occurred
// and the application must exit.
func (s *ShimSession) handleListenerError(err error, stop bool) {
	if err != nil {
		bluetooth.ErrorEvents().PublishAdded(wrapError(err))
	}

	if stop {
		s.Stop()
	}
}

// executor forms a request using the provided parameters, generates a unique request ID,
// and sends the request to the server. The request is tracked, and any responses to the
// request will be handled by the listener.
//
// This function is mainly used by the 'commands' package.
func (s *ShimSession) executor(params []string) (chan commands.CommandResponse, error) {
	if s.sessionClosed.Load() {
		return nil, errorkinds.ErrSessionNotExist
	}

	s.Lock()
	defer s.Unlock()

	s.id.Inc()
	replyChan := make(chan commands.CommandResponse, 1)
	s.requestMap.Store(s.id.Value(), replyChan)

	command := map[string]any{
		"command":    params,
		"request_id": s.id.Value(),
	}

	commandBytes, err := serde.MarshalJson(command)
	if err != nil {
		return nil, err
	}

	if _, err = s.conn.Write(commandBytes); err != nil {
		return nil, err
	}
	if _, err = s.conn.Write([]byte("\n")); err != nil {
		return nil, err
	}

	return replyChan, nil
}

// reset resets the state of the session. If 'isClosed' is true (i.e the session is stopped),
// it will close the socket connection. If 'isClosed is false (i.e the session is started),
// all session internals are initialized.
func (s *ShimSession) reset(isClosed bool) context.Context {
	s.Lock()
	defer s.Unlock()

	s.features = nil

	s.sessionClosed.Store(isClosed)
	if isClosed {
		s.cleanup()

		return context.Background()
	}

	s.id = xsync.NewCounter()
	s.requestMap = xsync.NewMapOf[int64, chan commands.CommandResponse]()

	s.listenerEvents = make(chan []byte, 1)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.store = sstore.NewSessionStore()

	return ctx
}

// cleanup is called by 'reset()' to close all listeners and connections when
// the session is stopped.
func (s *ShimSession) cleanup() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.conn != nil {
		s.conn.Close()
	}
}

func wrapError(err error) errorkinds.GenericError {
	return errorkinds.GenericError{Errors: err}
}
