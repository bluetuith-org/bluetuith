package gonetworkmanager

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/godbus/dbus/v5"
)

const (
	dbusMethodAddMatch = "org.freedesktop.DBus.AddMatch"
)

type dbusBase struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

func (d *dbusBase) init(iface string, objectPath dbus.ObjectPath) error {
	var err error

	d.conn, err = dbus.SystemBus()
	if err != nil {
		return err
	}

	d.obj = d.conn.Object(iface, objectPath)

	return nil
}

func (d *dbusBase) call(method string, args ...interface{}) error {
	return d.obj.Call(method, 0, args...).Err
}

func (d *dbusBase) callWithReturn(ret interface{}, method string, args ...interface{}) error {
	return d.obj.Call(method, 0, args...).Store(ret)
}

func (d *dbusBase) callWithReturn2(ret1 interface{}, ret2 interface{}, method string, args ...interface{}) error {
	return d.obj.Call(method, 0, args...).Store(ret1, ret2)
}

func (d *dbusBase) subscribe(iface, member string) {
	rule := fmt.Sprintf("type='signal',interface='%s',path='%s',member='%s'",
		iface, d.obj.Path(), NetworkManagerInterface)
	d.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
}

func (d *dbusBase) subscribeNamespace(namespace string) {
	rule := fmt.Sprintf("type='signal',path_namespace='%s'", namespace)
	d.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
}

func (d *dbusBase) getProperty(iface string) (interface{}, error) {
	variant, err := d.obj.GetProperty(iface)
	return variant.Value(), err
}

func (d *dbusBase) setProperty(iface string, value interface{}) error {
	err := d.obj.SetProperty(iface, dbus.MakeVariant(value))
	return err
}

func (d *dbusBase) getObjectProperty(iface string) (value dbus.ObjectPath, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(dbus.ObjectPath)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceObjectProperty(iface string) (value []dbus.ObjectPath, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]dbus.ObjectPath)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getBoolProperty(iface string) (value bool, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(bool)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getStringProperty(iface string) (value string, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(string)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceStringProperty(iface string) (value []string, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]string)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceSliceByteProperty(iface string) (value [][]byte, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([][]byte)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getMapStringVariantProperty(iface string) (value map[string]dbus.Variant, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(map[string]dbus.Variant)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getUint8Property(iface string) (value uint8, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(uint8)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getUint32Property(iface string) (value uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getInt32Property(iface string) (value int32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(int32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getInt64Property(iface string) (value int64, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(int64)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getUint64Property(iface string) (value uint64, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(uint64)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceUint32Property(iface string) (value []uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceSliceUint32Property(iface string) (value [][]uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([][]uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceMapStringVariantProperty(iface string) (value []map[string]dbus.Variant, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]map[string]dbus.Variant)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceByteProperty(iface string) (value []byte, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]byte)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func makeErrVariantType(iface string) error {
	return fmt.Errorf("unexpected variant type for '%s'", iface)
}

func ip4ToString(ip uint32) string {
	bs := []byte{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(bs, ip)
	return net.IP(bs).String()
}
