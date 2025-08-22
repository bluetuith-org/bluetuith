package gonetworkmanager

//go:generate stringer -type=NmConnectivity
type NmConnectivity uint32

const (
	NmConnectivityUnknown NmConnectivity = 0 // Network connectivity is unknown. This means the connectivity checks are disabled (e.g. on server installations) or has not run yet. The graphical shell should assume the Internet connection might be available and not present a captive portal window.
	NmConnectivityNone    NmConnectivity = 1 // The host is not connected to any network. There's no active connection that contains a default route to the internet and thus it makes no sense to even attempt a connectivity check. The graphical shell should use this state to indicate the network connection is unavailable.
	NmConnectivityPortal  NmConnectivity = 2 // The Internet connection is hijacked by a captive portal gateway. The graphical shell may open a sandboxed web browser window (because the captive portals typically attempt a man-in-the-middle attacks against the https connections) for the purpose of authenticating to a gateway and retrigger the connectivity check with CheckConnectivity() when the browser window is dismissed.
	NmConnectivityLimited NmConnectivity = 3 // The host is connected to a network, does not appear to be able to reach the full Internet, but a captive portal has not been detected.
	NmConnectivityFull    NmConnectivity = 4 // The host is connected to a network, and appears to be able to reach the full Internet.
)

//go:generate stringer -type=NmState
type NmState uint32

const (
	NmStateUnknown         NmState = 0  // Networking state is unknown. This indicates a daemon error that makes it unable to reasonably assess the state. In such event the applications are expected to assume Internet connectivity might be present and not disable controls that require network access. The graphical shells may hide the network accessibility indicator altogether since no meaningful status indication can be provided.
	NmStateAsleep          NmState = 10 // Networking is not enabled, the system is being suspended or resumed from suspend.
	NmStateDisconnected    NmState = 20 // There is no active network connection. The graphical shell should indicate no network connectivity and the applications should not attempt to access the network.
	NmStateDisconnecting   NmState = 30 // Network connections are being cleaned up. The applications should tear down their network sessions.
	NmStateConnecting      NmState = 40 // A network connection is being started The graphical shell should indicate the network is being connected while the applications should still make no attempts to connect the network.
	NmStateConnectedLocal  NmState = 50 // There is only local IPv4 and/or IPv6 connectivity, but no default route to access the Internet. The graphical shell should indicate no network connectivity.
	NmStateConnectedSite   NmState = 60 // There is only site-wide IPv4 and/or IPv6 connectivity. This means a default route is available, but the Internet connectivity check (see "Connectivity" property) did not succeed. The graphical shell should indicate limited network connectivity.
	NmStateConnectedGlobal NmState = 70 // There is global IPv4 and/or IPv6 Internet connectivity This means the Internet connectivity check succeeded, the graphical shell should indicate full network connectivity.
)

//go:generate stringer -type=NmCheckpointCreateFlags
type NmCheckpointCreateFlags uint32

const (
	NmCheckpointCreateFlagsNone                 NmCheckpointCreateFlags = 0    // no flags
	NmCheckpointCreateFlagsDestroyAll           NmCheckpointCreateFlags = 0x01 // when creating a new checkpoint, destroy all existing ones.
	NmCheckpointCreateFlagsDeleteNewConnections NmCheckpointCreateFlags = 0x02 // upon rollback, delete any new connection added after the checkpoint (Since: 1.6)
	NmCheckpointCreateFlagsDisconnectNewDevices NmCheckpointCreateFlags = 0x04 // upon rollback, disconnect any new device appeared after the checkpoint (Since: 1.6)
	NmCheckpointCreateFlagsAllowOverlapping     NmCheckpointCreateFlags = 0x08 // by default, creating a checkpoint fails if there are already existing checkoints that reference the same devices. With this flag, creation of such checkpoints is allowed, however, if an older checkpoint that references overlapping devices gets rolled back, it will automatically destroy this checkpoint during rollback. This allows to create several overlapping checkpoints in parallel, and rollback to them at will. With the special case that rolling back to an older checkpoint will invalidate all overlapping younger checkpoints. This opts-in that the checkpoint can be automatically destroyed by the rollback of an older checkpoint. (Since: 1.12)
)

//go:generate stringer -type=NmCapability
type NmCapability uint32

const (
	NmCapabilityTeam NmCapability = 1 // Teams can be managed
)

//go:generate stringer -type=NmMetered
type NmMetered uint32

const (
	NmMeteredUnknown  NmMetered = 0 // The metered status is unknown
	NmMeteredYes      NmMetered = 1 // Metered, the value was statically set
	NmMeteredNo       NmMetered = 2 // Not metered, the value was statically set
	NmMeteredGuessYes NmMetered = 3 // Metered, the value was guessed
	NmMeteredGuessNo  NmMetered = 4 // Not metered, the value was guessed
)

//go:generate stringer -type=NmDeviceState
type NmDeviceState uint32

const (
	NmDeviceStateUnknown      NmDeviceState = 0   // the device's state is unknown
	NmDeviceStateUnmanaged    NmDeviceState = 10  // the device is recognized, but not managed by NetworkManager
	NmDeviceStateUnavailable  NmDeviceState = 20  // the device is managed by NetworkManager, but is not available for use. Reasons may include the wireless switched off, missing firmware, no ethernet carrier, missing supplicant or modem manager, etc.
	NmDeviceStateDisconnected NmDeviceState = 30  // the device can be activated, but is currently idle and not connected to a network.
	NmDeviceStatePrepare      NmDeviceState = 40  // the device is preparing the connection to the network. This may include operations like changing the MAC address, setting physical link properties, and anything else required to connect to the requested network.
	NmDeviceStateConfig       NmDeviceState = 50  // the device is connecting to the requested network. This may include operations like associating with the Wi-Fi AP, dialing the modem, connecting to the remote Bluetooth device, etc.
	NmDeviceStateNeedAuth     NmDeviceState = 60  // the device requires more information to continue connecting to the requested network. This includes secrets like WiFi passphrases, login passwords, PIN codes, etc.
	NmDeviceStateIpConfig     NmDeviceState = 70  // the device is requesting IPv4 and/or IPv6 addresses and routing information from the network.
	NmDeviceStateIpCheck      NmDeviceState = 80  // the device is checking whether further action is required for the requested network connection. This may include checking whether only local network access is available, whether a captive portal is blocking access to the Internet, etc.
	NmDeviceStateSecondaries  NmDeviceState = 90  // the device is waiting for a secondary connection (like a VPN) which must activated before the device can be activated
	NmDeviceStateActivated    NmDeviceState = 100 // the device has a network connection, either local or global.
	NmDeviceStateDeactivating NmDeviceState = 110 // a disconnection from the current network connection was requested, and the device is cleaning up resources used for that connection. The network connection may still be valid.
	NmDeviceStateFailed       NmDeviceState = 120 // the device failed to connect to the requested network and is cleaning up the connection request
)

//go:generate stringer -type=NmActiveConnectionState
type NmActiveConnectionState uint32

const (
	NmActiveConnectionStateUnknown      NmActiveConnectionState = 0 // The state of the connection is unknown
	NmActiveConnectionStateActivating   NmActiveConnectionState = 1 // A network connection is being prepared
	NmActiveConnectionStateActivated    NmActiveConnectionState = 2 // There is a connection to the network
	NmActiveConnectionStateDeactivating NmActiveConnectionState = 3 // The network connection is being torn down and cleaned up
	NmActiveConnectionStateDeactivated  NmActiveConnectionState = 4 // The network connection is disconnected and will be removed
)

//go:generate stringer -type=NmActivationStateFlag
type NmActivationStateFlag uint32

const (
	NmActivationStateFlagNone                             NmActivationStateFlag = 0x00 // an alias for numeric zero, no flags set.
	NmActivationStateFlagIsMaster                         NmActivationStateFlag = 0x01 // the device is a master.
	NmActivationStateFlagIsSlave                          NmActivationStateFlag = 0x02 // the device is a slave.
	NmActivationStateFlagLayer2Ready                      NmActivationStateFlag = 0x04 // layer2 is activated and ready.
	NmActivationStateFlagIp4Ready                         NmActivationStateFlag = 0x08 // IPv4 setting is completed.
	NmActivationStateFlagIp6Ready                         NmActivationStateFlag = 0x10 // IPv6 setting is completed.
	NmActivationStateFlagMasterHasSlaves                  NmActivationStateFlag = 0x20 // The master has any slave devices attached. This only makes sense if the device is a master.
	NmActivationStateFlagLifetimeBoundToProfileVisibility NmActivationStateFlag = 0x40 // the lifetime of the activation is bound to the visilibity of the connection profile, which in turn depends on "connection.permissions" and whether a session for the user exists. Since: 1.16
)

//go:generate stringer -type=NmDeviceType
type NmDeviceType uint32

const (
	NmDeviceTypeUnknown      NmDeviceType = 0  // unknown device
	NmDeviceTypeGeneric      NmDeviceType = 14 // generic support for unrecognized device types
	NmDeviceTypeEthernet     NmDeviceType = 1  // a wired ethernet device
	NmDeviceTypeWifi         NmDeviceType = 2  // an 802.11 Wi-Fi device
	NmDeviceTypeUnused1      NmDeviceType = 3  // not used
	NmDeviceTypeUnused2      NmDeviceType = 4  // not used
	NmDeviceTypeBt           NmDeviceType = 5  // a Bluetooth device supporting PAN or DUN access protocols
	NmDeviceTypeOlpcMesh     NmDeviceType = 6  // an OLPC XO mesh networking device
	NmDeviceTypeWimax        NmDeviceType = 7  // an 802.16e Mobile WiMAX broadband device
	NmDeviceTypeModem        NmDeviceType = 8  // a modem supporting analog telephone, CDMA/EVDO, GSM/UMTS, or LTE network access protocols
	NmDeviceTypeInfiniband   NmDeviceType = 9  // an IP-over-InfiniBand device
	NmDeviceTypeBond         NmDeviceType = 10 // a bond master interface
	NmDeviceTypeVlan         NmDeviceType = 11 // an 802.1Q VLAN interface
	NmDeviceTypeAdsl         NmDeviceType = 12 // ADSL modem
	NmDeviceTypeBridge       NmDeviceType = 13 // a bridge master interface
	NmDeviceTypeTeam         NmDeviceType = 15 // a team master interface
	NmDeviceTypeTun          NmDeviceType = 16 // a TUN or TAP interface
	NmDeviceTypeIpTunnel     NmDeviceType = 17 // a IP tunnel interface
	NmDeviceTypeMacvlan      NmDeviceType = 18 // a MACVLAN interface
	NmDeviceTypeVxlan        NmDeviceType = 19 // a VXLAN interface
	NmDeviceTypeVeth         NmDeviceType = 20 // a VETH interface
	NmDeviceTypeMacsec       NmDeviceType = 21 // a MACsec interface
	NmDeviceTypeDummy        NmDeviceType = 22 // a dummy interface
	NmDeviceTypePpp          NmDeviceType = 23 // a PPP interface
	NmDeviceTypeOvsInterface NmDeviceType = 24 // a Open vSwitch interface
	NmDeviceTypeOvsPort      NmDeviceType = 25 // a Open vSwitch port
	NmDeviceTypeOvsBridge    NmDeviceType = 26 // a Open vSwitch bridge
	NmDeviceTypeWpan         NmDeviceType = 27 // a IEEE 802.15.4 (WPAN) MAC Layer Device
	NmDeviceType6lowpan      NmDeviceType = 28 // 6LoWPAN interface
	NmDeviceTypeWireguard    NmDeviceType = 29 // a WireGuard interface
	NmDeviceTypeWifiP2p      NmDeviceType = 30 // an 802.11 Wi-Fi P2P device (Since: 1.16)
)

//go:generate stringer -type=Nm80211APFlags
type Nm80211APFlags uint32

const (
	Nm80211APFlagsNone    Nm80211APFlags = 0x0
	Nm80211APFlagsPrivacy Nm80211APFlags = 0x1
)

//go:generate stringer -type=Nm80211APSec
type Nm80211APSec uint32

const (
	Nm80211APSecNone         Nm80211APSec = 0x0
	Nm80211APSecPairWEP40    Nm80211APSec = 0x1
	Nm80211APSecPairWEP104   Nm80211APSec = 0x2
	Nm80211APSecPairTKIP     Nm80211APSec = 0x4
	Nm80211APSecPairCCMP     Nm80211APSec = 0x8
	Nm80211APSecGroupWEP40   Nm80211APSec = 0x10
	Nm80211APSecGroupWEP104  Nm80211APSec = 0x20
	Nm80211APSecGroupTKIP    Nm80211APSec = 0x40
	Nm80211APSecGroupCCMP    Nm80211APSec = 0x80
	Nm80211APSecKeyMgmtPSK   Nm80211APSec = 0x100
	Nm80211APSecKeyMgmt8021X Nm80211APSec = 0x200
	Nm80211APSecKeyMgmtSAE   Nm80211APSec = 0x400
	Nm80211APSecKeyMgmtOWE   Nm80211APSec = 0x800
	Nm80211APSecKeyMgmtOWETM Nm80211APSec = 0x1000
)

//go:generate stringer -type=Nm80211Mode
type Nm80211Mode uint32

const (
	Nm80211ModeUnknown Nm80211Mode = 0
	Nm80211ModeAdhoc   Nm80211Mode = 1
	Nm80211ModeInfra   Nm80211Mode = 2
	Nm80211ModeAp      Nm80211Mode = 3
)

//go:generate stringer -type=NmActiveConnectionState
type NmVpnConnectionState uint32

const (
	NmVpnConnectionUnknown      = 0 //The state of the VPN connection is unknown.
	NmVpnConnectionPrepare      = 1 //The VPN connection is preparing to connect.
	NmVpnConnectionNeedAuth     = 2 //The VPN connection needs authorization credentials.
	NmVpnConnectionConnect      = 3 //The VPN connection is being established.
	NmVpnConnectionIpConfigGet  = 4 //The VPN connection is getting an IP address.
	NmVpnConnectionActivated    = 5 //The VPN connection is active.
	NmVpnConnectionFailed       = 6 //The VPN connection failed.
	NmVpnConnectionDisconnected = 7 //The VPN connection is disconnected.
)

//go:generate stringer -type=NmDeviceStateReason
type NmDeviceStateReason uint32

const (
	NmDeviceStateReasonNone                        = 0  // No reason given
	NmDeviceStateReasonUnknown                     = 1  // Unknown error
	NmDeviceStateReasonNowManaged                  = 2  // Device is now managed
	NmDeviceStateReasonNowUnmanaged                = 3  // Device is now unmanaged
	NmDeviceStateReasonConfigFailed                = 4  // The device could not be readied for configuration
	NmDeviceStateReasonIpConfigUnavailable         = 5  // IP configuration could not be reserved (no available address, timeout, etc)
	NmDeviceStateReasonIpConfigExpired             = 6  // The IP config is no longer valid
	NmDeviceStateReasonNoSecrets                   = 7  // Secrets were required, but not provided
	NmDeviceStateReasonSupplicantDisconnect        = 8  // 802.1x supplicant disconnected
	NmDeviceStateReasonSupplicantConfigFailed      = 9  // 802.1x supplicant configuration failed
	NmDeviceStateReasonSupplicantFailed            = 10 // 802.1x supplicant failed
	NmDeviceStateReasonSupplicantTimeout           = 11 // 802.1x supplicant took too long to authenticate
	NmDeviceStateReasonPppStartFailed              = 12 // PPP service failed to start
	NmDeviceStateReasonPppDisconnect               = 13 // PPP service disconnected
	NmDeviceStateReasonPppFailed                   = 14 // PPP failed
	NmDeviceStateReasonDhcpStartFailed             = 15 // DHCP client failed to start
	NmDeviceStateReasonDhcpError                   = 16 // DHCP client error
	NmDeviceStateReasonDhcpFailed                  = 17 // DHCP client failed
	NmDeviceStateReasonSharedStartFailed           = 18 // Shared connection service failed to start
	NmDeviceStateReasonSharedFailed                = 19 // Shared connection service failed
	NmDeviceStateReasonAutoipStartFailed           = 20 // AutoIP service failed to start
	NmDeviceStateReasonAutoipError                 = 21 // AutoIP service error
	NmDeviceStateReasonAutoipFailed                = 22 // AutoIP service failed
	NmDeviceStateReasonModemBusy                   = 23 // The line is busy
	NmDeviceStateReasonModemNoDialTone             = 24 // No dial tone
	NmDeviceStateReasonModemNoCarrier              = 25 // No carrier could be established
	NmDeviceStateReasonModemDialTimeout            = 26 // The dialing request timed out
	NmDeviceStateReasonModemDialFailed             = 27 // The dialing attempt failed
	NmDeviceStateReasonModemInitFailed             = 28 // Modem initialization failed
	NmDeviceStateReasonGsmApnFailed                = 29 // Failed to select the specified APN
	NmDeviceStateReasonGsmRegistrationNotSearching = 30 // Not searching for networks
	NmDeviceStateReasonGsmRegistrationDenied       = 31 // Network registration denied
	NmDeviceStateReasonGsmRegistrationTimeout      = 32 // Network registration timed out
	NmDeviceStateReasonGsmRegistrationFailed       = 33 // Failed to register with the requested network
	NmDeviceStateReasonGsmPinCheckFailed           = 34 // PIN check failed
	NmDeviceStateReasonFirmwareMissing             = 35 // Necessary firmware for the device may be missing
	NmDeviceStateReasonRemoved                     = 36 // The device was removed
	NmDeviceStateReasonSleeping                    = 37 // NetworkManager went to sleep
	NmDeviceStateReasonConnectionRemoved           = 38 // The device's active connection disappeared
	NmDeviceStateReasonUserRequested               = 39 // Device disconnected by user or client
	NmDeviceStateReasonCarrier                     = 40 // Carrier/link changed
	NmDeviceStateReasonConnectionAssumed           = 41 // The device's existing connection was assumed
	NmDeviceStateReasonSupplicantAvailable         = 42 // The supplicant is now available
	NmDeviceStateReasonModemNotFound               = 43 // The modem could not be found
	NmDeviceStateReasonBtFailed                    = 44 // The Bluetooth connection failed or timed out
	NmDeviceStateReasonGsmSimNotInserted           = 45 // GSM Modem's SIM Card not inserted
	NmDeviceStateReasonGsmSimPinRequired           = 46 // GSM Modem's SIM Pin required
	NmDeviceStateReasonGsmSimPukRequired           = 47 // GSM Modem's SIM Puk required
	NmDeviceStateReasonGsmSimWrong                 = 48 // GSM Modem's SIM wrong
	NmDeviceStateReasonInfinibandMode              = 49 // InfiniBand device does not support connected mode
	NmDeviceStateReasonDependencyFailed            = 50 // A dependency of the connection failed
	NmDeviceStateReasonBr2684Failed                = 51 // Problem with the RFC 2684 Ethernet over ADSL bridge
	NmDeviceStateReasonModemManagerUnavailable     = 52 // ModemManager not running
	NmDeviceStateReasonSsidNotFound                = 53 // The Wi-Fi network could not be found
	NmDeviceStateReasonSecondaryConnectionFailed   = 54 // A secondary connection of the base connection failed
	NmDeviceStateReasonDcbFcoeFailed               = 55 // DCB or FCoE setup failed
	NmDeviceStateReasonTeamdControlFailed          = 56 // teamd control failed
	NmDeviceStateReasonModemFailed                 = 57 // Modem failed or no longer available
	NmDeviceStateReasonModemAvailable              = 58 // Modem now ready and available
	NmDeviceStateReasonSimPinIncorrect             = 59 // SIM PIN was incorrect
	NmDeviceStateReasonNewActivation               = 60 // New connection activation was enqueued
	NmDeviceStateReasonParentChanged               = 61 // the device's parent changed
	NmDeviceStateReasonParentManagedChanged        = 62 // the device parent's management changed
	NmDeviceStateReasonOvsdbFailed                 = 63 // problem communicating with Open vSwitch database
	NmDeviceStateReasonIpAddressDuplicate          = 64 // a duplicate IP address was detected
	NmDeviceStateReasonIpMethodUnsupported         = 65 // The selected IP method is not supported
	NmDeviceStateReasonSriovConfigurationFailed    = 66 // configuration of SR-IOV parameters failed
	NmDeviceStateReasonPeerNotFound                = 67 // The Wi-Fi P2P peer could not be found
)

//go:generate stringer -type=NmActiveConnectionStateReason
type NmActiveConnectionStateReason uint32

const (
	NmActiveConnectionStateReasonUnknown             = 0  // The reason for the active connection state change is unknown.
	NmActiveConnectionStateReasonNone                = 1  // No reason was given for the active connection state change.
	NmActiveConnectionStateReasonUserDisconnected    = 2  // The active connection changed state because the user disconnected it.
	NmActiveConnectionStateReasonDeviceDisconnected  = 3  // The active connection changed state because the device it was using was disconnected.
	NmActiveConnectionStateReasonServiceStopped      = 4  // The service providing the VPN connection was stopped.
	NmActiveConnectionStateReasonIpConfigInvalid     = 5  // The IP config of the active connection was invalid.
	NmActiveConnectionStateReasonConnectTimeout      = 6  // The connection attempt to the VPN service timed out.
	NmActiveConnectionStateReasonServiceStartTimeout = 7  // A timeout occurred while starting the service providing the VPN connection.
	NmActiveConnectionStateReasonServiceStartFailed  = 8  // Starting the service providing the VPN connection failed.
	NmActiveConnectionStateReasonNoSecrets           = 9  // Necessary secrets for the connection were not provided.
	NmActiveConnectionStateReasonLoginFailed         = 10 // Authentication to the server failed.
	NmActiveConnectionStateReasonConnectionRemoved   = 11 // The connection was deleted from settings.
	NmActiveConnectionStateReasonDependencyFailed    = 12 // Master connection of this connection failed to activate.
	NmActiveConnectionStateReasonDeviceRealizeFailed = 13 // Could not create the software device link.
	NmActiveConnectionStateReasonDeviceRemoved       = 14 // The device this connection depended on disappeared.
)

//go:generate stringer -type=NmRollbackResult
type NmRollbackResult uint32

const (
	NmRollbackResultOk                 = 0 // the rollback succeeded.
	NmRollbackResultErrNoDevice        = 1 // the device no longer exists.
	NmRollbackResultErrDeviceUnmanaged = 2 // the device is now unmanaged.
	NmRollbackResultErrFailed          = 3 // other errors during rollback.
)
