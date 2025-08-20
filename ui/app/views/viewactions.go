package views

import (
	"context"
	"errors"

	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/keybindings"
)

// viewActions holds an instance of a view actions manager,
// which maps different actions to their respective view action contexts and actions.
type viewActions struct {
	rv *Views

	fnmap map[viewActionContext]map[keybindings.Key]func(set ...string) bool
}

// viewActionContext describes the context in which the
// action is supposed to be executed in.
type viewActionContext int

// The different context types for actions.
const (
	actionInvoke viewActionContext = iota
	actionInitializer
	actionVisibility
)

// newViewActions returns a new view actions manager.
func newViewActions(rv *Views) *viewActions {
	v := &viewActions{rv: rv}

	return v.initViewActions()
}

// initViewActions initializes and stores the different view actions based on their view action contexts.
func (v *viewActions) initViewActions() *viewActions {
	v.fnmap = map[viewActionContext]map[keybindings.Key]func(set ...string) bool{
		actionInvoke: {
			keybindings.KeyAdapterTogglePower:        v.power,
			keybindings.KeyAdapterToggleDiscoverable: v.discoverable,
			keybindings.KeyAdapterTogglePairable:     v.pairable,
			keybindings.KeyAdapterToggleScan:         v.scan,
			keybindings.KeyAdapterChange:             v.changeAdapter,
			keybindings.KeyDeviceConnect:             v.connect,
			keybindings.KeyDevicePair:                v.pair,
			keybindings.KeyDeviceTrust:               v.trust,
			keybindings.KeyDeviceBlock:               v.block,
			keybindings.KeyDeviceSendFiles:           v.send,
			keybindings.KeyDeviceNetwork:             v.networkAP,
			keybindings.KeyDeviceAudioProfiles:       v.profiles,
			keybindings.KeyPlayerShow:                v.showplayer,
			keybindings.KeyDeviceInfo:                v.info,
			keybindings.KeyDeviceRemove:              v.remove,
			keybindings.KeyProgressView:              v.progress,
			keybindings.KeyPlayerHide:                v.hideplayer,
			keybindings.KeyQuit:                      v.quit,
		},
		actionInitializer: {
			keybindings.KeyAdapterTogglePower:        v.initPower,
			keybindings.KeyAdapterToggleDiscoverable: v.initDiscoverable,
			keybindings.KeyAdapterTogglePairable:     v.initPairable,
			keybindings.KeyDeviceConnect:             v.initConnect,
			keybindings.KeyDeviceTrust:               v.initTrust,
			keybindings.KeyDeviceBlock:               v.initBlock,
		},
		actionVisibility: {
			keybindings.KeyDeviceSendFiles:     v.visibleSend,
			keybindings.KeyDeviceNetwork:       v.visibleNetwork,
			keybindings.KeyDeviceAudioProfiles: v.visibleProfile,
			keybindings.KeyPlayerShow:          v.visiblePlayer,
		},
	}

	return v
}

// handler executes the handler assigned to the key type based on
// the action context.
func (v *viewActions) handler(key keybindings.Key, actionContext viewActionContext) func() bool {
	handler := v.fnmap[actionContext][key]

	if actionContext == actionInvoke {
		return func() bool {
			go handler()
			return false
		}
	}

	return func() bool {
		return handler()
	}
}

// power checks and toggles the adapter's powered state.
func (v *viewActions) power(set ...string) bool {
	var poweredText string

	props, err := v.rv.adapter.currentSession().Properties()
	if err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	powered := props.Powered

	if set != nil {
		state := set[0] == "yes"
		if state == powered {
			return false
		}

		powered = !state
	}

	if err := v.rv.adapter.currentSession().SetPoweredState(!powered); err != nil {
		v.rv.status.ErrorMessage(errors.New("cannot set adapter power state"))
		return false
	}

	if powered {
		poweredText = "off"
	} else {
		poweredText = "on"
	}

	v.rv.status.InfoMessage(props.UniqueName+" is powered "+poweredText, false)

	v.rv.menu.toggleItemByKey(keybindings.KeyAdapterTogglePower, !powered)

	return true
}

// discoverable checks and toggles the adapter's discoverable state.
func (v *viewActions) discoverable(set ...string) bool {
	var discoverableText string

	props, err := v.rv.adapter.currentSession().Properties()
	if err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	discoverable := props.Discoverable

	if set != nil {
		state := set[0] == "yes"
		if state == discoverable {
			return false
		}

		discoverable = !state
	}

	if err := v.rv.adapter.currentSession().SetDiscoverableState(!discoverable); err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	if !discoverable {
		discoverableText = "discoverable"
	} else {
		discoverableText = "not discoverable"
	}

	v.rv.status.InfoMessage(props.UniqueName+" is "+discoverableText, false)

	v.rv.menu.toggleItemByKey(keybindings.KeyAdapterToggleDiscoverable, !discoverable)

	return true
}

// pairable checks and toggles the adapter's pairable state.
func (v *viewActions) pairable(set ...string) bool {
	var pairableText string

	props, err := v.rv.adapter.currentSession().Properties()
	if err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	pairable := props.Pairable

	if set != nil {
		state := set[0] == "yes"
		if state == pairable {
			return false
		}

		pairable = !state
	}

	if err := v.rv.adapter.currentSession().SetPairableState(!pairable); err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	if !pairable {
		pairableText = "pairable"
	} else {
		pairableText = "not pairable"
	}

	v.rv.status.InfoMessage(props.UniqueName+" is "+pairableText, false)

	v.rv.menu.toggleItemByKey(keybindings.KeyAdapterTogglePairable, !pairable)

	return true
}

// scan checks the current adapter's state and starts/stops discovery.
func (v *viewActions) scan(set ...string) bool {
	props, err := v.rv.adapter.currentSession().Properties()
	if err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	discover := props.Discovering

	if set != nil {
		state := set[0] == "yes"
		if state == discover {
			return false
		}

		discover = !state
	}

	if !discover {
		if err := v.rv.app.Session().Adapter(props.Address).StartDiscovery(); err != nil {
			v.rv.status.ErrorMessage(err)
			return false
		}
		v.rv.status.InfoMessage("Scanning for devices...", true)
	} else {
		if err := v.rv.app.Session().Adapter(props.Address).StopDiscovery(); err != nil {
			v.rv.status.ErrorMessage(err)
			return false
		}
		v.rv.status.InfoMessage("Scanning stopped", false)
	}

	v.rv.menu.toggleItemByKey(keybindings.KeyAdapterToggleScan, !discover)

	return true
}

// changeAdapter launches a popup with the adapters list.
func (v *viewActions) changeAdapter(_ ...string) bool {
	v.rv.app.QueueDraw(func() {
		v.rv.adapter.change()
	})

	return true
}

// progress displays the progress view.
func (v *viewActions) progress(_ ...string) bool {
	v.rv.app.QueueDraw(func() {
		v.rv.progress.show()
	})

	return true
}

// quit stops discovery mode for all existing adapters, closes the bluetooth connection
// and exits the application.
func (v *viewActions) quit(_ ...string) bool {
	if v.rv.cfg.Values.ConfirmOnQuit && v.rv.status.SetInput("Quit (y/n)?") != "y" {
		return false
	}

	if adapters, err := v.rv.app.Session().Adapters(); err != nil {
		for _, adapter := range adapters {
			v.rv.app.Session().Adapter(adapter.Address).StopDiscovery()
		}
	}

	v.rv.app.Close()

	return true
}

// initPower _s the oncreate handler for the power submenu option.
func (v *viewActions) initPower(_ ...string) bool {
	props, err := v.rv.adapter.currentSession().Properties()

	return err == nil && props.Powered
}

// initDiscoverable _s the oncreate handler for the discoverable submenu option
func (v *viewActions) initDiscoverable(_ ...string) bool {
	props, err := v.rv.adapter.currentSession().Properties()

	return err == nil && props.Discoverable
}

// initPairable _s the oncreate handler for the pairable submenu option.
func (v *viewActions) initPairable(_ ...string) bool {
	props, err := v.rv.adapter.currentSession().Properties()

	return err == nil && props.Pairable
}

// initConnect _s the oncreate handler for the connect submenu option.
func (v *viewActions) initConnect(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return device.Connected
}

// initTrust _s the oncreate handler for the trust submenu option.
func (v *viewActions) initTrust(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return device.Trusted
}

// initBlock _s the oncreate handler for the block submenu option.
func (v *viewActions) initBlock(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return device.Blocked
}

// visibleSend _s the visible handler for the send submenu option.
func (v *viewActions) visibleSend(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return v.rv.app.Features().Has(appfeatures.FeatureSendFile, appfeatures.FeatureReceiveFile) &&
		device.HaveService(bluetooth.ObexObjpushServiceClass)
}

// visibleNetwork _s the visible handler for the network submenu option.
func (v *viewActions) visibleNetwork(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return v.rv.app.Features().Has(appfeatures.FeatureNetwork) &&
		device.HaveService(bluetooth.NapServiceClass) &&
		(device.HaveService(bluetooth.PanuServiceClass) ||
			device.HaveService(bluetooth.DialupNetServiceClass))
}

// visibleProfile _s the visible handler for the audio profiles submenu option.
func (v *viewActions) visibleProfile(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return device.HaveService(bluetooth.AudioSourceServiceClass) ||
		device.HaveService(bluetooth.AudioSinkServiceClass)
}

// visiblePlayer _s the visible handler for the media player submenu option.
func (v *viewActions) visiblePlayer(_ ...string) bool {
	device := v.rv.device.getSelection(false)
	if device.Address.IsNil() {
		return false
	}

	return device.HaveService(bluetooth.AudioSourceServiceClass) &&
		device.HaveService(bluetooth.AvRemoteServiceClass) &&
		device.HaveService(bluetooth.AvRemoteTargetServiceClass)
}

// connect retrieves the selected device, and toggles its connection state.
func (v *viewActions) connect(set ...string) bool {
	var device bluetooth.DeviceData

	if set != nil {
		devices, err := v.rv.adapter.currentSession().Devices()
		if err != nil {
			v.rv.status.ErrorMessage(err)
			return false
		}

		for _, d := range devices {
			if d.Address.String() == set[0] {
				device = d
				break
			}
		}
	} else {
		device = v.rv.device.getSelection(true)
		if device.Address.IsNil() {
			return false
		}
	}

	disconnectFunc := func() {
		if err := v.rv.app.Session().Device(device.Address).Disconnect(); err != nil {
			v.rv.status.ErrorMessage(err)
			return
		}

		v.rv.player.closeForDevice(device.Address)
	}

	connectFunc := func() {
		v.rv.status.InfoMessage("Connecting to "+device.Name, true)
		if err := v.rv.app.Session().Device(device.Address).Connect(); err != nil {
			v.rv.status.ErrorMessage(err)
			return
		}
		v.rv.status.InfoMessage("Connected to "+device.Name, false)
	}

	if !device.Connected {
		v.rv.op.startOperation(
			connectFunc,
			func() {
				disconnectFunc()
				v.rv.status.InfoMessage("Cancelled connection to "+device.Name, false)
			},
		)
	} else {
		v.rv.status.InfoMessage("Disconnecting from "+device.Name, true)
		disconnectFunc()
		v.rv.status.InfoMessage("Disconnected from "+device.Name, false)
	}

	v.rv.menu.toggleItemByKey(keybindings.KeyDeviceConnect, !device.Connected)

	return true
}

// pair retrieves the selected device, and attempts to pair with it.
func (v *viewActions) pair(_ ...string) bool {
	device := v.rv.device.getSelection(true)
	if device.Address.IsNil() {
		return false
	}
	if device.Paired {
		v.rv.status.InfoMessage(device.Name+" is already paired", false)
		return false
	}

	v.rv.op.startOperation(
		func() {
			v.rv.status.InfoMessage("Pairing with "+device.Name, true)
			if err := v.rv.app.Session().Device(device.Address).Pair(); err != nil {
				v.rv.status.ErrorMessage(err)
				return
			}
			v.rv.status.InfoMessage("Paired with "+device.Name, false)
		},
		func() {
			if err := v.rv.app.Session().Device(device.Address).CancelPairing(); err != nil {
				v.rv.status.ErrorMessage(err)
				return
			}
			v.rv.status.InfoMessage("Cancelled pairing with "+device.Name, false)
		},
	)

	return true
}

// trust retrieves the selected device, and toggles its trust property.
func (v *viewActions) trust(_ ...string) bool {
	device := v.rv.device.getSelection(true)
	if device.Address.IsNil() {
		return false
	}

	if err := v.rv.app.Session().Device(device.Address).SetTrusted(!device.Trusted); err != nil {
		v.rv.status.ErrorMessage(errors.New("cannot set trusted property for " + device.Name))
		return false
	}

	v.rv.menu.toggleItemByKey(keybindings.KeyDeviceTrust, !device.Trusted)

	return true
}

// block retrieves the selected device, and toggles its block property.
func (v *viewActions) block(_ ...string) bool {
	device := v.rv.device.getSelection(true)
	if device.Address.IsNil() {
		return false
	}

	if err := v.rv.app.Session().Device(device.Address).SetBlocked(!device.Blocked); err != nil {
		v.rv.status.ErrorMessage(errors.New("cannot set blocked property for " + device.Name))
		return false
	}

	v.rv.menu.toggleItemByKey(keybindings.KeyDeviceBlock, !device.Blocked)

	return true
}

// send gets a file list from the file picker and sends all selected files
// to the target device.
func (v *viewActions) send(_ ...string) bool {
	device := v.rv.device.getSelection(true)
	if !device.Paired || !device.Connected {
		v.rv.status.ErrorMessage(errors.New(device.Name + " is not paired and/or connected"))
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())

	v.rv.op.startOperation(
		func() {
			v.rv.status.InfoMessage("Initializing OBEX Session()..", true)
			oppSession := v.rv.app.Session().Obex(device.Address).ObjectPush()

			err := oppSession.CreateSession(ctx)
			if err != nil {
				v.rv.status.ErrorMessage(err)
				return
			}

			v.rv.op.cancelOperation(false)

			v.rv.status.InfoMessage("Created OBEX session", false)

			fileList, err := v.rv.filepicker.Show()
			if err != nil {
				v.rv.status.ErrorMessage(err)
				oppSession.RemoveSession()
				return
			}
			if len(fileList) == 0 {
				oppSession.RemoveSession()
				return
			}

			proplist := make([]bluetooth.ObjectPushData, 0, len(fileList))
			for _, file := range fileList {
				props, err := oppSession.SendFile(file)
				if err != nil || props.Status == bluetooth.TransferError {
					oppSession.RemoveSession()
					v.rv.status.ErrorMessage(err)
					return
				}

				proplist = append(proplist, props)
			}

			v.rv.progress.startTransfer(device.Address, oppSession, proplist)
		},
		func() {
			cancel()
			v.rv.status.InfoMessage("Cancelled OBEX session creation", false)
		},
	)

	return true
}

// networkAP launches a popup with the available networks.
func (v *viewActions) networkAP(_ ...string) bool {
	v.rv.app.QueueDraw(func() {
		v.rv.network.networkSelect()
	})

	return true
}

// profiles launches a popup with the available audio profiles.
func (v *viewActions) profiles(_ ...string) bool {
	v.rv.app.QueueDraw(func() {
		v.rv.audioProfiles.audioProfiles()
	})

	return true
}

// showplayer starts the media player.
func (v *viewActions) showplayer(_ ...string) bool {
	v.rv.player.show()

	return true
}

// hideplayer hides the media player.
func (v *viewActions) hideplayer(_ ...string) bool {
	v.rv.player.close()

	return true
}

// info retrieves the selected device, and shows the device information.
func (v *viewActions) info(_ ...string) bool {
	v.rv.app.QueueDraw(func() {
		v.rv.device.showDetailedInfo()
	})

	return true
}

// remove retrieves the selected device, and removes it from the adapter.
func (v *viewActions) remove(_ ...string) bool {
	device := v.rv.device.getSelection(true)
	if device.Address.IsNil() {
		return false
	}

	if txt := v.rv.status.SetInput("Remove " + device.Name + " (y/n)?"); txt != "y" {
		return false
	}

	if err := v.rv.app.Session().Device(device.Address).Remove(); err != nil {
		v.rv.status.ErrorMessage(err)
		return false
	}

	v.rv.status.InfoMessage("Removed "+device.Name, false)

	return true
}
