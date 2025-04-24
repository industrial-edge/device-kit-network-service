package gonetworkmanager

import (
	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/mock"
)

type MockActiveConnection struct {
	mock.Mock
}

func (m *MockActiveConnection) GetPath() dbus.ObjectPath {
	args := m.Called()
	return args.Get(0).(dbus.ObjectPath)
}

func (m *MockActiveConnection) GetPropertyConnection() (nm.Connection, error) {
	args := m.Called()
	return args.Get(0).(nm.Connection), args.Error(1)
}

func (m *MockActiveConnection) GetPropertySpecificObject() (nm.AccessPoint, error) {
	args := m.Called()
	return args.Get(0).(nm.AccessPoint), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyUUID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyType() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyDevices() ([]nm.Device, error) {
	args := m.Called()
	return args.Get(0).([]nm.Device), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyState() (nm.NmActiveConnectionState, error) {
	args := m.Called()
	return args.Get(0).(nm.NmActiveConnectionState), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyStateFlags() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyDefault() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyIP4Config() (nm.IP4Config, error) {
	args := m.Called()
	return args.Get(0).(nm.IP4Config), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyDHCP4Config() (nm.DHCP4Config, error) {
	args := m.Called()
	return args.Get(0).(nm.DHCP4Config), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyDefault6() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyIP6Config() (nm.IP6Config, error) {
	args := m.Called()
	return args.Get(0).(nm.IP6Config), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyDHCP6Config() (nm.DHCP6Config, error) {
	args := m.Called()
	return args.Get(0).(nm.DHCP6Config), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyVPN() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockActiveConnection) GetPropertyMaster() (nm.Device, error) {
	args := m.Called()
	return args.Get(0).(nm.Device), args.Error(1)
}

func (m *MockActiveConnection) SubscribeState(receiver chan nm.StateChange, exit chan struct{}) error {
	args := m.Called(receiver, exit)
	return args.Error(0)
}
