package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceStatisticsInterface = DeviceInterface + ".Statistics"

	// Properties
	DeviceStatisticsPropertyRefreshRateMs = DeviceStatisticsInterface + ".RefreshRateMs" // readwrite  u
	DeviceStatisticsPropertyTxBytes       = DeviceStatisticsInterface + ".TxBytes"       // readable   t
	DeviceStatisticsPropertyRxBytes       = DeviceStatisticsInterface + ".RxBytes"       // readable   t
)

type DeviceStatistics interface {
	GetPath() dbus.ObjectPath

	// GetPropertyRefreshRateMs Refresh rate of the rest of properties of this interface. The properties are guaranteed to be refreshed each RefreshRateMs milliseconds in case the underlying counter has changed too. If zero, there is no guaranteed refresh rate of the properties.
	GetPropertyRefreshRateMs() (uint32, error)

	SetPropertyRefreshRateMs(uint32) error

	// GetPropertyTxBytes Number of transmitted bytes
	GetPropertyTxBytes() (uint64, error)

	// GetPropertyRxBytes Number of received bytes
	GetPropertyRxBytes() (uint64, error)
}

func NewDeviceStatistics(objectPath dbus.ObjectPath) (DeviceStatistics, error) {
	var d deviceStatistics
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type deviceStatistics struct {
	dbusBase
}

func (d *deviceStatistics) GetPath() dbus.ObjectPath {
	return d.obj.Path()
}

func (d *deviceStatistics) GetPropertyRefreshRateMs() (uint32, error) {
	return d.getUint32Property(DeviceStatisticsPropertyRefreshRateMs)
}

func (d *deviceStatistics) SetPropertyRefreshRateMs(rate uint32) error {
	return d.setProperty(DeviceStatisticsPropertyRefreshRateMs, rate)
}

func (d *deviceStatistics) GetPropertyTxBytes() (uint64, error) {
	return d.getUint64Property(DeviceStatisticsPropertyTxBytes)
}

func (d *deviceStatistics) GetPropertyRxBytes() (uint64, error) {
	return d.getUint64Property(DeviceStatisticsPropertyRxBytes)
}

func (d *deviceStatistics) marshalMap() map[string]interface{} {
	return map[string]interface{}{}
}

func (d *deviceStatistics) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	m["RefreshRateMs"], _ = d.GetPropertyRefreshRateMs()
	m["TxBytes"], _ = d.GetPropertyTxBytes()
	m["RxBytes"], _ = d.GetPropertyRxBytes()

	return json.Marshal(m)
}
