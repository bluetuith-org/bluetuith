//go:build linux

package networkmanager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	nm "github.com/Wifx/gonetworkmanager"
	ac "github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"
)

// Network holds the network manager and active connections.
type Network struct {
	Address      bluetooth.MacAddress
	DeviceExists func() error

	*NetManager
	bluetooth.NetworkDunSettings
}

// NetManager holds the network manager session.
type NetManager struct {
	ActiveConnection xsync.MapOf[bluetooth.MacAddress, nm.ActiveConnection]

	nm.NetworkManager
}

// connectionSettings holds a device's network connection settings.
type connectionSettings struct {
	Name    string
	Address bluetooth.MacAddress

	ConnectionType bluetooth.NetworkType
	ConnectionUUID uuid.UUID

	bluetooth.NetworkDunSettings
}

// Initialize initializes and returns a new NetManager.
func Initialize() (*NetManager, ac.Features, *ac.Error) {
	manager, err := nm.NewNetworkManager()
	if err != nil {
		return nil, ac.FeatureNone,
			ac.NewError(ac.FeatureNetwork, err)
	}

	network := &NetManager{
		NetworkManager:   manager,
		ActiveConnection: xsync.MapOf[bluetooth.MacAddress, nm.ActiveConnection]{},
	}

	return network, ac.FeatureNetwork, nil
}

// Connect connects to the device's network interface.
func (n *Network) Connect(name string, nt bluetooth.NetworkType) error {
	if err := n.check(); err != nil {
		return err
	}

	active, err := n.isConnectionActive(nt)
	if err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "network-connect-active",
				"address", n.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot verify connection status"),
		)
	}

	if active {
		return errorkinds.ErrNetworkAlreadyActive
	}

	activated, err := n.activateExistingConnection(nt)
	if err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "network-connect-activated",
				"address", n.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot activate existing connection"),
		)
	}

	if activated {
		return nil
	}

	if err := n.createConnection(name, nt); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "network-connect-create",
				"address", n.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot create connection"),
		)
	}

	return nil
}

// Disconnect deactivates the connection.
func (n *Network) Disconnect() error {
	if err := n.check(); err != nil {
		return err
	}

	activeConn, ok := n.ActiveConnection.Load(n.Address)
	if ok {
		n.ActiveConnection.Delete(n.Address)
	}

	if activeConn == nil {
		return nil
	}

	if err := n.DeactivateConnection(activeConn); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "network-disconnect-deactivated",
				"address", n.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot deactivate connection"),
		)
	}

	return nil
}

// isConnectionActive checks if the device's connection is active.
func (n *Network) isConnectionActive(nt bluetooth.NetworkType) (bool, error) {
	activeConnections, err := n.GetPropertyActiveConnections()
	if err != nil {
		return false, err
	}

	for _, activeConn := range activeConnections {
		ctype, err := activeConn.GetPropertyType()
		if err != nil {
			return false, err
		}

		if ctype != "bluetooth" {
			continue
		}

		conn, err := activeConn.GetPropertyConnection()
		if err != nil {
			return false, err
		}

		exist, err := n.addrExist(conn, nt)
		if err != nil {
			return false, err
		}

		if exist {
			return true, nil
		}
	}

	return false, nil
}

// activateExistingConnection activates an existing device connection profile.
func (n *Network) activateExistingConnection(nt bluetooth.NetworkType) (bool, error) {
	devices, err := n.GetPropertyDevices()
	if err != nil {
		return false, err
	}

	for _, device := range devices {
		dtype, err := device.GetPropertyDeviceType()
		if err != nil {
			return false, err
		}

		if dtype != nm.NmDeviceTypeBt {
			continue
		}

		conns, err := device.GetPropertyAvailableConnections()
		if err != nil {
			return false, err
		}

		for _, conn := range conns {
			exist, err := n.addrExist(conn, nt)
			if err != nil {
				return false, err
			}

			if exist {
				err := n.applySettings(conn, nt)
				if err != nil {
					return false, err
				}

				return true, n.activateConnection(conn, device)
			}
		}
	}

	return false, nil
}

// createConnection creates a new connection.
func (n *Network) createConnection(name string, nt bluetooth.NetworkType) error {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	connectionSettings := connectionSettings{
		Address:            n.Address,
		Name:               name,
		ConnectionType:     nt,
		ConnectionUUID:     newUUID,
		NetworkDunSettings: n.NetworkDunSettings,
	}.toMap()

	settings, err := nm.NewSettings()
	if err != nil {
		return err
	}

	conn, err := settings.AddConnection(connectionSettings)
	if err != nil {
		return err
	}

	device, err := n.GetDeviceByIpIface(n.Address.String())
	if err != nil {
		return err
	}

	return n.activateConnection(conn, device)
}

// activateConnection activates the connection.
func (n *Network) activateConnection(conn nm.Connection, device nm.Device) error {
	var state nm.StateChange

	activeConn, err := n.ActivateConnection(conn, device, nil)
	if err != nil {
		return err
	}

	exit := make(chan struct{})
	activeState := make(chan nm.StateChange)

	err = activeConn.SubscribeState(activeState, exit)
	if err != nil {
		return err
	}

	n.ActiveConnection.Store(n.Address, activeConn)

	for state := range activeState {
		if state.State == nm.NmActiveConnectionStateActivating {
			continue
		}

		exit <- struct{}{}

		break
	}

	if state.State != nm.NmActiveConnectionStateActivated {
		return errorkinds.ErrNetworkEstablishError
	}

	return nil
}

// addrExist checks if the device's address is present in the connection's settings.
func (n *Network) addrExist(conn nm.Connection, nt bluetooth.NetworkType) (bool, error) {
	settings, err := conn.GetSettings()
	if err != nil {
		return false, err
	}

	addr, ok := settings["bluetooth"]["bdaddr"].([]byte)
	if ok {
		bdtype, ok := settings["bluetooth"]["type"].(string)

		if ok &&
			bluetooth.MacAddress(addr) == n.Address &&
			bdtype == nt.String() {
			return true, nil
		}
	}

	return false, nil
}

// applySettings checks and modifies the device connection's settings.
func (n *Network) applySettings(conn nm.Connection, nt bluetooth.NetworkType) error {
	if nt != bluetooth.NetworkDun {
		return nil
	}

	settings, err := conn.GetSettings()
	if err != nil {
		return err
	}

	gsmSettings, ok := settings["gsm"]
	if !ok {
		return errors.New("GSM setting not found in connection details")
	}

	if n.APN != gsmSettings["apn"] {
		gsmSettings["apn"] = n.APN
	}

	if n.Number != gsmSettings["number"] {
		gsmSettings["number"] = n.Number
	}

	delete(settings, "ipv6")

	return conn.Update(settings)
}

// check checks whether the network manager was initialized.
func (n *Network) check() error {
	if n.NetManager == nil {
		return fault.Wrap(
			errorkinds.ErrNetworkInitSession,
			fctx.With(context.Background(),
				"error_at", "network-check-manager",
				"address", n.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot call network manager method"),
		)
	}

	_, ok := dbh.PathConverter.DbusPath(dbh.DbusPathDevice, n.Address)
	if !ok {
		return fault.Wrap(errorkinds.ErrDeviceNotFound,
			fctx.With(context.Background(),
				"error_at", "network-check-device",
				"address", n.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Device does not exist"),
		)
	}

	return nil
}

// toMap returns a connection setting as a map.
func (n connectionSettings) toMap() map[string]map[string]any {
	connType := n.ConnectionType.String()
	name := fmt.Sprintf("%s Access Point (%s)",
		n.Name, strings.ToUpper(connType),
	)

	settings := map[string]map[string]any{
		"connection": {
			"id":          name,
			"type":        "bluetooth",
			"uuid":        n.ConnectionUUID.String(),
			"autoconnect": false,
		},
		"bluetooth": {
			"bdaddr": n.Address[:],
			"type":   connType,
		},
	}

	if n.ConnectionType == bluetooth.NetworkDun {
		settings["gsm"] = map[string]any{
			"apn":    n.APN,
			"number": n.Number,
		}
	}

	return settings
}
