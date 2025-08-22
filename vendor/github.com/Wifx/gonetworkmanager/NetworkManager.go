package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	NetworkManagerInterface  = "org.freedesktop.NetworkManager"
	NetworkManagerObjectPath = "/org/freedesktop/NetworkManager"

	/* Methods */
	NetworkManagerReload                          = NetworkManagerInterface + ".Reload"
	NetworkManagerGetDevices                      = NetworkManagerInterface + ".GetDevices"
	NetworkManagerGetAllDevices                   = NetworkManagerInterface + ".GetAllDevices"
	NetworkManagerGetDeviceByIpIface              = NetworkManagerInterface + ".GetDeviceByIpIface"
	NetworkManagerActivateConnection              = NetworkManagerInterface + ".ActivateConnection"
	NetworkManagerAddAndActivateConnection        = NetworkManagerInterface + ".AddAndActivateConnection"
	NetworkManagerAddAndActivateConnection2       = NetworkManagerInterface + ".AddAndActivateConnection2"
	NetworkManagerDeactivateConnection            = NetworkManagerInterface + ".DeactivateConnection"
	NetworkManagerSleep                           = NetworkManagerInterface + ".Sleep"
	NetworkManagerEnable                          = NetworkManagerInterface + ".Enable"
	NetworkManagerGetPermissions                  = NetworkManagerInterface + ".GetPermissions"
	NetworkManagerSetLogging                      = NetworkManagerInterface + ".SetLogging"
	NetworkManagerGetLogging                      = NetworkManagerInterface + ".GetLogging"
	NetworkManagerCheckConnectivity               = NetworkManagerInterface + ".CheckConnectivity"
	NetworkManagerState                           = NetworkManagerInterface + ".state"
	NetworkManagerCheckpointCreate                = NetworkManagerInterface + ".CheckpointCreate"
	NetworkManagerCheckpointDestroy               = NetworkManagerInterface + ".CheckpointDestroy"
	NetworkManagerCheckpointRollback              = NetworkManagerInterface + ".CheckpointRollback"
	NetworkManagerCheckpointAdjustRollbackTimeout = NetworkManagerInterface + ".CheckpointAdjustRollbackTimeout"

	/* Property */
	NetworkManagerPropertyDevices                    = NetworkManagerInterface + ".Devices"                    // readable   ao
	NetworkManagerPropertyAllDevices                 = NetworkManagerInterface + ".AllDevices"                 // readable   ao
	NetworkManagerPropertyCheckpoints                = NetworkManagerInterface + ".Checkpoints"                // readable   ao
	NetworkManagerPropertyNetworkingEnabled          = NetworkManagerInterface + ".NetworkingEnabled"          // readable   b
	NetworkManagerPropertyWirelessEnabled            = NetworkManagerInterface + ".WirelessEnabled"            // readwrite  b
	NetworkManagerPropertyWirelessHardwareEnabled    = NetworkManagerInterface + ".WirelessHardwareEnabled"    // readable   b
	NetworkManagerPropertyWwanEnabled                = NetworkManagerInterface + ".WwanEnabled"                // readwrite  b
	NetworkManagerPropertyWwanHardwareEnabled        = NetworkManagerInterface + ".WwanHardwareEnabled"        // readable   b
	NetworkManagerPropertyWimaxEnabled               = NetworkManagerInterface + ".WimaxEnabled"               // readwrite  b
	NetworkManagerPropertyWimaxHardwareEnabled       = NetworkManagerInterface + ".WimaxHardwareEnabled"       // readable   b
	NetworkManagerPropertyActiveConnections          = NetworkManagerInterface + ".ActiveConnections"          // readable   ao
	NetworkManagerPropertyPrimaryConnection          = NetworkManagerInterface + ".PrimaryConnection"          // readable   o
	NetworkManagerPropertyPrimaryConnectionType      = NetworkManagerInterface + ".PrimaryConnectionType"      // readable   s
	NetworkManagerPropertyMetered                    = NetworkManagerInterface + ".Metered"                    // readable   u
	NetworkManagerPropertyActivatingConnection       = NetworkManagerInterface + ".ActivatingConnection"       // readable   o
	NetworkManagerPropertyStartup                    = NetworkManagerInterface + ".Startup"                    // readable   b
	NetworkManagerPropertyVersion                    = NetworkManagerInterface + ".Version"                    // readable   s
	NetworkManagerPropertyCapabilities               = NetworkManagerInterface + ".Capabilities"               // readable   au
	NetworkManagerPropertyState                      = NetworkManagerInterface + ".State"                      // readable   u
	NetworkManagerPropertyConnectivity               = NetworkManagerInterface + ".Connectivity"               // readable   u
	NetworkManagerPropertyConnectivityCheckAvailable = NetworkManagerInterface + ".ConnectivityCheckAvailable" // readable   b
	NetworkManagerPropertyConnectivityCheckEnabled   = NetworkManagerInterface + ".ConnectivityCheckEnabled"   // readwrite  b
	NetworkManagerPropertyGlobalDnsConfiguration     = NetworkManagerInterface + ".GlobalDnsConfiguration"     // readwrite  a{sv}
)

type NetworkManager interface {
	/* METHODS */

	// Reload NetworkManager's configuration and perform certain updates, like flushing a cache or rewriting external state to disk. This is similar to sending SIGHUP to NetworkManager but it allows for more fine-grained control over what to reload (see flags). It also allows non-root access via PolicyKit and contrary to signals it is synchronous.
	// No flags (0x00) means to reload everything that is supported which is identical to sending a SIGHUP.
	// (0x01) means to reload the NetworkManager.conf configuration from disk. Note that this does not include connections, which can be reloaded via Setting's ReloadConnections.
	// (0x02) means to update DNS configuration, which usually involves writing /etc/resolv.conf anew.
	// (0x04) means to restart the DNS plugin. This is for example useful when using dnsmasq plugin, which uses additional configuration in /etc/NetworkManager/dnsmasq.d. If you edit those files, you can restart the DNS plugin. This action shortly interrupts name resolution. Note that flags may affect each other. For example, restarting the DNS plugin (0x04) implicitly updates DNS too (0x02). Or when reloading the configuration (0x01), changes to DNS setting also cause a DNS update (0x02). However, (0x01) does not involve restarting the DNS plugin (0x04) or update resolv.conf (0x02), unless the DNS related configuration changes in NetworkManager.conf.
	Reload(flags uint32) error

	// GetDevices Get the list of realized network devices.
	GetDevices() ([]Device, error)

	// GetAllDevices Get the list of all network devices.
	GetAllDevices() ([]Device, error)

	// GetDeviceByIpIface Return the object path of the network device referenced by its IP interface name. Note that some devices (usually modems) only have an IP interface name when they are connected.
	GetDeviceByIpIface(interfaceId string) (Device, error)

	// ActivateConnection Activate a connection using the supplied device.
	ActivateConnection(connection Connection, device Device, specificObject *dbus.Object) (ActiveConnection, error)

	// AddAndActivateConnection Adds a new connection using the given details (if any) as a template (automatically filling in missing settings with the capabilities of the given device), then activate the new connection. Cannot be used for VPN connections at this time.
	AddAndActivateConnection(connection map[string]map[string]interface{}, device Device) (ActiveConnection, error)

	// ActivateWirelessConnection requests activating access point to network device
	ActivateWirelessConnection(connection Connection, device Device, accessPoint AccessPoint) (ActiveConnection, error)

	// AddAndActivateWirelessConnection adds a new connection profile to the network device it has been
	// passed. It then activates the connection to the passed access point. The first parameter contains
	// additional information for the connection (most probably the credentials).
	// Example contents for connection are:
	// connection := make(map[string]map[string]interface{})
	// connection["802-11-wireless"] = make(map[string]interface{})
	// connection["802-11-wireless"]["security"] = "802-11-wireless-security"
	// connection["802-11-wireless-security"] = make(map[string]interface{})
	// connection["802-11-wireless-security"]["key-mgmt"] = "wpa-psk"
	// connection["802-11-wireless-security"]["psk"] = password
	AddAndActivateWirelessConnection(connection map[string]map[string]interface{}, device Device, accessPoint AccessPoint) (ActiveConnection, error)

	// DeactivateConnection Deactivate an active connection.
	DeactivateConnection(connection ActiveConnection) error

	// Sleep Control the NetworkManager daemon's sleep state. When asleep, all interfaces that it manages are deactivated. When awake, devices are available to be activated. This command should not be called directly by users or clients; it is intended for system suspend/resume tracking.
	// sleepnWake: Indicates whether the NetworkManager daemon should sleep or wake.
	Sleep(sleepNWake bool) error

	// Enable Control whether overall networking is enabled or disabled. When disabled, all interfaces that NM manages are deactivated. When enabled, all managed interfaces are re-enabled and available to be activated. This command should be used by clients that provide to users the ability to enable/disable all networking.
	// enableNDisable: If FALSE, indicates that all networking should be disabled. If TRUE, indicates that NetworkManager should begin managing network devices.
	Enable(enableNDisable bool) error

	// CheckConnectivity Re-check the network connectivity state.
	CheckConnectivity() error

	// State The overall networking state as determined by the NetworkManager daemon, based on the state of network devices under its management.
	State() (NmState, error)

	// CheckpointCreate Create a checkpoint of the current networking configuration for given interfaces. If rollback_timeout is not zero, a rollback is automatically performed after the given timeout.
	// devices: A list of device paths for which a checkpoint should be created. An empty list means all devices.
	// rollbackTimeout: The time in seconds until NetworkManager will automatically rollback to the checkpoint. Set to zero for infinite.
	// flags: Flags for the creation.
	// returns: On success, the new checkpoint.
	CheckpointCreate(devices []Device, rollbackTimeout uint32, flags uint32) (Checkpoint, error)

	// CheckpointDestroy Destroy a previously created checkpoint.
	// checkpoint: The checkpoint to be destroyed. Set to empty to cancel all pending checkpoints.
	CheckpointDestroy(checkpoint Checkpoint) error

	// CheckpointRollback Rollback a checkpoint before the timeout is reached.
	// checkpoint: The checkpoint to be rolled back.
	// result: On return, a dictionary of devices and results. Devices are represented by their original D-Bus path; each result is a RollbackResult.
	CheckpointRollback(checkpoint Checkpoint) (result map[dbus.ObjectPath]NmRollbackResult, err error)

	// CheckpointAdjustRollbackTimeout Reset the timeout for rollback for the checkpoint.
	// Since: 1.12
	// addTimeout: number of seconds from ~now~ in which the timeout will expire. Set to 0 to disable the timeout. Note that the added seconds start counting from now, not "Created" timestamp or the previous expiration time. Note that the "Created" property of the checkpoint will stay unchanged by this call. However, the "RollbackTimeout" will be recalculated to give the approximate new expiration time. The new "RollbackTimeout" property will be approximate up to one second precision, which is the accuracy of the property.
	CheckpointAdjustRollbackTimeout(checkpoint Checkpoint, addTimeout uint32) error

	/* PROPERTIES */

	// GetPropertyDevices The list of realized network devices. Realized devices are those which have backing resources (eg from the kernel or a management daemon like ModemManager, teamd, etc).
	GetPropertyDevices() ([]Device, error)

	// GetPropertyAllDevices The list of both realized and un-realized network devices. Un-realized devices are software devices which do not yet have backing resources, but for which backing resources can be created if the device is activated.
	GetPropertyAllDevices() ([]Device, error)

	// GetPropertyCheckpoints The list of active checkpoints.
	GetPropertyCheckpoints() ([]Checkpoint, error)

	// GetPropertyNetworkingEnabled Indicates if overall networking is currently enabled or not. See the Enable() method.
	GetPropertyNetworkingEnabled() (bool, error)

	// GetPropertyWirelessEnabled Indicates if wireless is currently enabled or not.
	GetPropertyWirelessEnabled() (bool, error)
	SetPropertyWirelessEnabled(bool) error

	// GetPropertyWirelessHardwareEnabled Indicates if the wireless hardware is currently enabled, i.e. the state of the RF kill switch.
	GetPropertyWirelessHardwareEnabled() (bool, error)

	// GetPropertyWwanEnabled Indicates if mobile broadband devices are currently enabled or not.
	GetPropertyWwanEnabled() (bool, error)

	// GetPropertyWwanHardwareEnabled Indicates if the mobile broadband hardware is currently enabled, i.e. the state of the RF kill switch.
	GetPropertyWwanHardwareEnabled() (bool, error)

	// GetPropertyWimaxEnabled Indicates if WiMAX devices are currently enabled or not.
	GetPropertyWimaxEnabled() (bool, error)

	// GetPropertyWimaxHardwareEnabled Indicates if the WiMAX hardware is currently enabled, i.e. the state of the RF kill switch.
	GetPropertyWimaxHardwareEnabled() (bool, error)

	// GetPropertyActiveConnections List of active connection object paths.
	GetPropertyActiveConnections() ([]ActiveConnection, error)

	// GetPropertyPrimaryConnection The object path of the "primary" active connection being used to access the network. In particular, if there is no VPN active, or the VPN does not have the default route, then this indicates the connection that has the default route. If there is a VPN active with the default route, then this indicates the connection that contains the route to the VPN endpoint.
	GetPropertyPrimaryConnection() (ActiveConnection, error)

	// GetPropertyPrimaryConnectionType The connection type of the "primary" active connection being used to access the network. This is the same as the Type property on the object indicated by PrimaryConnection.
	GetPropertyPrimaryConnectionType() (string, error)

	// GetPropertyMetered Indicates whether the connectivity is metered. This is equivalent to the metered property of the device associated with the primary connection.
	GetPropertyMetered() (NmMetered, error)

	// GetPropertyActivatingConnection The object path of an active connection that is currently being activated and which is expected to become the new PrimaryConnection when it finishes activating.
	GetPropertyActivatingConnection() (ActiveConnection, error)

	// GetPropertyStartup Indicates whether NM is still starting up; this becomes FALSE when NM has finished attempting to activate every connection that it might be able to activate at startup.
	GetPropertyStartup() (bool, error)

	// GetPropertyVersion NetworkManager version.
	GetPropertyVersion() (string, error)

	// GetPropertyCapabilities The current set of capabilities. See NMCapability for currently defined capability numbers. The array is guaranteed to be sorted in ascending order without duplicates.
	GetPropertyCapabilities() ([]NmCapability, error)

	// GetPropertyState The overall state of the NetworkManager daemon.
	// This takes state of all active connections and the connectivity state into account to produce a single indicator of the network accessibility status.
	// The graphical shells may use this property to provide network connection status indication and applications may use this to check if Internet connection is accessible. Shell that is able to cope with captive portals should use the "Connectivity" property to decide whether to present a captive portal authentication dialog.
	GetPropertyState() (NmState, error)

	// GetPropertyConnectivity The result of the last connectivity check. The connectivity check is triggered automatically when a default connection becomes available, periodically and by calling a CheckConnectivity() method.
	// This property is in general useful for the graphical shell to determine whether the Internet access is being hijacked by an authentication gateway (a "captive portal"). In such case it would typically present a web browser window to give the user a chance to authenticate and call CheckConnectivity() when the user submits a form or dismisses the window.
	// To determine the whether the user is able to access the Internet without dealing with captive portals (e.g. to provide a network connection indicator or disable controls that require Internet access), the "State" property is more suitable.
	GetPropertyConnectivity() (NmConnectivity, error)

	// GetPropertyConnectivityCheckAvailable Indicates whether connectivity checking service has been configured. This may return true even if the service is not currently enabled.
	// This is primarily intended for use in a privacy control panel, as a way to determine whether to show an option to enable/disable the feature.
	GetPropertyConnectivityCheckAvailable() (bool, error)

	// GetPropertyConnectivityCheckEnabled Indicates whether connectivity checking is enabled. This property can also be written to to disable connectivity checking (as a privacy control panel might want to do).
	GetPropertyConnectivityCheckEnabled() (bool, error)

	// Dictionary of global DNS settings where the key is one of "searches", "options" and "domains". The values for the "searches" and "options" keys are string arrays describing the list of search domains and resolver options, respectively. The value of the "domains" key is a second-level dictionary, where each key is a domain name, and each key's value is a third-level dictionary with the keys "servers" and "options". "servers" is a string array of DNS servers, "options" is a string array of domain-specific options.
	//GetPropertyGlobalDnsConfiguration() []interface{}

	Subscribe() <-chan *dbus.Signal
	Unsubscribe()

	MarshalJSON() ([]byte, error)
}

func NewNetworkManager() (NetworkManager, error) {
	var nm networkManager
	return &nm, nm.init(NetworkManagerInterface, NetworkManagerObjectPath)
}

type networkManager struct {
	dbusBase

	sigChan chan *dbus.Signal
}

func (nm *networkManager) Reload(flags uint32) error {
	return nm.call(NetworkManagerReload, flags)
}

func (nm *networkManager) GetDevices() (devices []Device, err error) {
	var devicePaths []dbus.ObjectPath
	err = nm.callWithReturn(&devicePaths, NetworkManagerGetDevices)
	if err != nil {
		return
	}

	devices = make([]Device, len(devicePaths))

	for i, path := range devicePaths {
		devices[i], err = DeviceFactory(path)
		if err != nil {
			return
		}
	}

	return
}

func (nm *networkManager) GetAllDevices() (devices []Device, err error) {
	var devicePaths []dbus.ObjectPath

	err = nm.callWithReturn(&devicePaths, NetworkManagerGetAllDevices)
	if err != nil {
		return
	}

	devices = make([]Device, len(devicePaths))

	for i, path := range devicePaths {
		devices[i], err = DeviceFactory(path)
		if err != nil {
			return
		}
	}

	return
}

func (nm *networkManager) GetDeviceByIpIface(interfaceId string) (device Device, err error) {
	var devicePath dbus.ObjectPath

	err = nm.callWithReturn(&devicePath, NetworkManagerGetDeviceByIpIface, interfaceId)
	if err != nil {
		return
	}

	device, err = DeviceFactory(devicePath)
	if err != nil {
		return
	}

	return
}

func (nm *networkManager) ActivateConnection(connection Connection, device Device, specificObject *dbus.Object) (ac ActiveConnection, err error) {
	var connectionPath dbus.ObjectPath

	var devicePath dbus.ObjectPath
	if device != nil {
		devicePath = device.GetPath()
	} else {
		devicePath = "/"
	}

	var specificObjectPath dbus.ObjectPath
	if specificObject != nil {
		specificObjectPath = specificObject.Path()
	} else {
		specificObjectPath = "/"
	}

	err = nm.callWithReturn(&connectionPath, NetworkManagerActivateConnection, connection.GetPath(), devicePath, specificObjectPath)
	if err != nil {
		return
	}

	ac, err = NewActiveConnection(connectionPath)
	if err != nil {
		return
	}

	return
}

func (nm *networkManager) AddAndActivateConnection(connection map[string]map[string]interface{}, d Device) (ac ActiveConnection, err error) {
	var opath1 dbus.ObjectPath
	var opath2 dbus.ObjectPath

	var devicePath dbus.ObjectPath
	if d != nil {
		devicePath = d.GetPath()
	}

	err = nm.callWithReturn2(&opath1, &opath2, NetworkManagerAddAndActivateConnection, connection, devicePath, dbus.ObjectPath("/"))
	if err != nil {
		return
	}

	ac, err = NewActiveConnection(opath2)
	if err != nil {
		return
	}

	return
}

func (nm *networkManager) ActivateWirelessConnection(c Connection, d Device, ap AccessPoint) (ac ActiveConnection, err error) {
	var opath dbus.ObjectPath
	err = nm.callWithReturn(&opath, NetworkManagerActivateConnection, c.GetPath(), d.GetPath(), ap.GetPath())
	if err != nil {
		return nil, err
	}

	ac, err = NewActiveConnection(opath)
	if err != nil {
		return nil, err
	}

	return
}

func (nm *networkManager) AddAndActivateWirelessConnection(connection map[string]map[string]interface{}, d Device, ap AccessPoint) (ac ActiveConnection, err error) {
	var opath1 dbus.ObjectPath
	var opath2 dbus.ObjectPath

	err = nm.callWithReturn2(&opath1, &opath2, NetworkManagerAddAndActivateConnection, connection, d.GetPath(), ap.GetPath())
	if err != nil {
		return
	}

	ac, err = NewActiveConnection(opath2)
	if err != nil {
		return
	}
	return
}

func (nm *networkManager) DeactivateConnection(c ActiveConnection) error {
	return nm.call(NetworkManagerDeactivateConnection, c.GetPath())
}

func (nm *networkManager) Sleep(sleepNWake bool) error {
	return nm.call(NetworkManagerSleep, sleepNWake)
}

func (nm *networkManager) Enable(enableNDisable bool) error {
	return nm.call(NetworkManagerEnable, enableNDisable)
}

func (nm *networkManager) CheckConnectivity() error {
	return nm.call(NetworkManagerCheckConnectivity)
}

func (nm *networkManager) State() (state NmState, err error) {
	err = nm.callWithReturn(&state, NetworkManagerState)
	return
}

func (nm *networkManager) CheckpointCreate(devices []Device, rollbackTimeout uint32, flags uint32) (cp Checkpoint, err error) {

	var devicePaths []dbus.ObjectPath
	if len(devices) > 0 {
		for _, device := range devices {
			devicePaths = append(devicePaths, device.GetPath())
		}
	}

	var checkpointPath dbus.ObjectPath
	err = nm.callWithReturn(&checkpointPath, NetworkManagerCheckpointCreate, devicePaths, rollbackTimeout, flags)
	if err != nil {
		return
	}

	cp, err = NewCheckpoint(checkpointPath)
	return
}

func (nm *networkManager) CheckpointDestroy(checkpoint Checkpoint) error {
	if checkpoint == nil {
		return nm.call(NetworkManagerCheckpointDestroy)
	} else {
		return nm.call(NetworkManagerCheckpointDestroy, checkpoint.GetPath())
	}
}

func (nm *networkManager) CheckpointRollback(checkpoint Checkpoint) (results map[dbus.ObjectPath]NmRollbackResult, err error) {

	var ret map[dbus.ObjectPath]NmRollbackResult

	err = nm.callWithReturn(&ret, NetworkManagerCheckpointRollback, checkpoint.GetPath())
	if err != nil {
		return
	}

	results = map[dbus.ObjectPath]NmRollbackResult{}
	for devicePath, result := range ret {
		results[devicePath] = result
	}

	return
}

func (nm *networkManager) CheckpointAdjustRollbackTimeout(checkpoint Checkpoint, addTimeout uint32) error {
	return nm.call(NetworkManagerCheckpointAdjustRollbackTimeout, checkpoint, addTimeout)
}

/* PROPERTIES */

func (nm *networkManager) GetPropertyDevices() ([]Device, error) {
	devicesPaths, err := nm.getSliceObjectProperty(NetworkManagerPropertyDevices)
	if err != nil {
		return nil, err
	}

	devices := make([]Device, len(devicesPaths))
	for i, path := range devicesPaths {
		devices[i], err = NewDevice(path)
		if err != nil {
			return devices, err
		}
	}

	return devices, nil
}

func (nm *networkManager) GetPropertyAllDevices() ([]Device, error) {
	devicesPaths, err := nm.getSliceObjectProperty(NetworkManagerPropertyAllDevices)
	if err != nil {
		return nil, err
	}

	devices := make([]Device, len(devicesPaths))
	for i, path := range devicesPaths {
		devices[i], err = NewDevice(path)
		if err != nil {
			return devices, err
		}
	}

	return devices, nil
}

func (nm *networkManager) GetPropertyCheckpoints() ([]Checkpoint, error) {
	checkpointsPaths, err := nm.getSliceObjectProperty(NetworkManagerPropertyCheckpoints)
	if err != nil {
		return nil, err
	}

	checkpoints := make([]Checkpoint, len(checkpointsPaths))
	for i, path := range checkpointsPaths {
		checkpoints[i], err = NewCheckpoint(path)
		if err != nil {
			return checkpoints, err
		}
	}

	return checkpoints, nil
}

func (nm *networkManager) GetPropertyNetworkingEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyNetworkingEnabled)
}

func (nm *networkManager) GetPropertyWirelessEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyWirelessEnabled)
}

func (nm *networkManager) SetPropertyWirelessEnabled(enabled bool) error {
	return nm.setProperty(NetworkManagerPropertyWirelessEnabled, enabled)
}

func (nm *networkManager) GetPropertyWirelessHardwareEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyWirelessHardwareEnabled)
}

func (nm *networkManager) GetPropertyWwanEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyWwanEnabled)
}

func (nm *networkManager) GetPropertyWwanHardwareEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyWwanHardwareEnabled)
}

func (nm *networkManager) GetPropertyWimaxEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyWimaxEnabled)
}

func (nm *networkManager) GetPropertyWimaxHardwareEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyWimaxHardwareEnabled)
}

func (nm *networkManager) GetPropertyActiveConnections() ([]ActiveConnection, error) {
	acPaths, err := nm.getSliceObjectProperty(NetworkManagerPropertyActiveConnections)
	if err != nil {
		return nil, err
	}

	ac := make([]ActiveConnection, len(acPaths))
	for i, path := range acPaths {
		ac[i], err = NewActiveConnection(path)
		if err != nil {
			return ac, err
		}
	}

	return ac, nil
}

func (nm *networkManager) GetPropertyPrimaryConnection() (ActiveConnection, error) {
	activeConnectionPath, err := nm.getObjectProperty(NetworkManagerPropertyPrimaryConnection)

	if err != nil {
		return nil, err
	}

	return NewActiveConnection(activeConnectionPath)
}

func (nm *networkManager) GetPropertyPrimaryConnectionType() (string, error) {
	return nm.getStringProperty(NetworkManagerPropertyPrimaryConnectionType)
}

func (nm *networkManager) GetPropertyMetered() (NmMetered, error) {
	v, err := nm.getUint32Property(NetworkManagerPropertyMetered)
	return NmMetered(v), err
}

func (nm *networkManager) GetPropertyActivatingConnection() (ActiveConnection, error) {
	panic("implement me")
}

func (nm *networkManager) GetPropertyStartup() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyStartup)
}

func (nm *networkManager) GetPropertyVersion() (string, error) {
	return nm.getStringProperty(NetworkManagerPropertyVersion)
}

func (nm *networkManager) GetPropertyCapabilities() ([]NmCapability, error) {
	panic("implement me")
}

func (nm *networkManager) GetPropertyState() (NmState, error) {
	v, err := nm.getUint32Property(NetworkManagerPropertyState)
	return NmState(v), err
}

func (nm *networkManager) GetPropertyConnectivity() (NmConnectivity, error) {
	v, err := nm.getUint32Property(NetworkManagerPropertyConnectivity)
	return NmConnectivity(v), err
}

func (nm *networkManager) GetPropertyConnectivityCheckAvailable() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyConnectivityCheckAvailable)
}

func (nm *networkManager) GetPropertyConnectivityCheckEnabled() (bool, error) {
	return nm.getBoolProperty(NetworkManagerPropertyConnectivityCheckEnabled)
}

func (nm *networkManager) Subscribe() <-chan *dbus.Signal {
	if nm.sigChan != nil {
		return nm.sigChan
	}

	nm.subscribeNamespace(NetworkManagerObjectPath)
	nm.sigChan = make(chan *dbus.Signal, 10)
	nm.conn.Signal(nm.sigChan)

	return nm.sigChan
}

func (nm *networkManager) Unsubscribe() {
	nm.conn.RemoveSignal(nm.sigChan)
	nm.sigChan = nil
}

func (nm *networkManager) MarshalJSON() ([]byte, error) {

	Devices, err := nm.GetPropertyDevices()
	if err != nil {
		return nil, err
	}

	AllDevices, err := nm.GetPropertyAllDevices()
	if err != nil {
		return nil, err
	}

	Checkpoints, err := nm.GetPropertyCheckpoints()
	if err != nil {
		return nil, err
	}

	NetworkingEnabled, err := nm.GetPropertyNetworkingEnabled()
	if err != nil {
		return nil, err
	}

	WirelessEnabled, err := nm.GetPropertyWirelessEnabled()
	if err != nil {
		return nil, err
	}

	WirelessHardwareEnabled, err := nm.GetPropertyWirelessHardwareEnabled()
	if err != nil {
		return nil, err
	}

	WwanEnabled, err := nm.GetPropertyWwanEnabled()
	if err != nil {
		return nil, err
	}

	WwanHardwareEnabled, err := nm.GetPropertyWwanHardwareEnabled()
	if err != nil {
		return nil, err
	}

	WimaxEnabled, err := nm.GetPropertyWimaxEnabled()
	if err != nil {
		return nil, err
	}

	WimaxHardwareEnabled, err := nm.GetPropertyWimaxHardwareEnabled()
	if err != nil {
		return nil, err
	}

	ActiveConnections, err := nm.GetPropertyActiveConnections()
	if err != nil {
		return nil, err
	}

	PrimaryConnection, err := nm.GetPropertyPrimaryConnection()
	if err != nil {
		return nil, err
	}

	PrimaryConnectionType, err := nm.GetPropertyPrimaryConnectionType()
	if err != nil {
		return nil, err
	}

	Metered, err := nm.GetPropertyMetered()
	if err != nil {
		return nil, err
	}

	ActivatingConnection, err := nm.GetPropertyActivatingConnection()
	if err != nil {
		return nil, err
	}

	Startup, err := nm.GetPropertyStartup()
	if err != nil {
		return nil, err
	}

	Version, err := nm.GetPropertyVersion()
	if err != nil {
		return nil, err
	}

	Capabilities, err := nm.GetPropertyCapabilities()
	if err != nil {
		return nil, err
	}

	State, err := nm.GetPropertyState()
	if err != nil {
		return nil, err
	}

	Connectivity, err := nm.GetPropertyConnectivity()
	if err != nil {
		return nil, err
	}

	ConnectivityCheckAvailable, err := nm.GetPropertyConnectivityCheckAvailable()
	if err != nil {
		return nil, err
	}

	ConnectivityCheckEnabled, err := nm.GetPropertyConnectivityCheckEnabled()
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"Devices":                    Devices,
		"AllDevices":                 AllDevices,
		"Checkpoints":                Checkpoints,
		"NetworkingEnabled":          NetworkingEnabled,
		"WirelessEnabled":            WirelessEnabled,
		"WirelessHardwareEnabled":    WirelessHardwareEnabled,
		"WwanEnabled":                WwanEnabled,
		"WwanHardwareEnabled":        WwanHardwareEnabled,
		"WimaxEnabled":               WimaxEnabled,
		"WimaxHardwareEnabled":       WimaxHardwareEnabled,
		"ActiveConnections":          ActiveConnections,
		"PrimaryConnection":          PrimaryConnection,
		"PrimaryConnectionType":      PrimaryConnectionType,
		"Metered":                    Metered,
		"ActivatingConnection":       ActivatingConnection,
		"Startup":                    Startup,
		"Version":                    Version,
		"Capabilities":               Capabilities,
		"State":                      State,
		"Connectivity":               Connectivity,
		"ConnectivityCheckAvailable": ConnectivityCheckAvailable,
		"ConnectivityCheckEnabled":   ConnectivityCheckEnabled,
	})
}
