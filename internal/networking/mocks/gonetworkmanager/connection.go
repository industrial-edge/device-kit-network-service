package gonetworkmanager

import (
	. "github.com/Wifx/gonetworkmanager/v2"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/mock"
)

type MockConnection struct {
	mock.Mock
}

func (m *MockConnection) GetPropertyConnection() (Connection, error) {
	args := m.Called()
	return args.Get(0).(Connection), args.Error(1)
}

func (m *MockConnection) GetPropertySpecificObject() (AccessPoint, error) {
	args := m.Called()
	return args.Get(0).(AccessPoint), args.Error(1)
}

func (m *MockConnection) GetPropertyID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConnection) GetPropertyUUID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConnection) GetPropertyType() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConnection) GetPropertyDevices() ([]Device, error) {
	args := m.Called()
	return args.Get(0).([]Device), args.Error(1)
}

func (m *MockConnection) GetPropertyState() (NmActiveConnectionState, error) {
	args := m.Called()
	return args.Get(0).(NmActiveConnectionState), args.Error(1)
}

func (m *MockConnection) GetPropertyStateFlags() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockConnection) GetPropertyDefault() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockConnection) GetPropertyIP4Config() (IP4Config, error) {
	args := m.Called()
	return args.Get(0).(IP4Config), args.Error(1)
}

func (m *MockConnection) GetPropertyDHCP4Config() (DHCP4Config, error) {
	args := m.Called()
	return args.Get(0).(DHCP4Config), args.Error(1)
}

func (m *MockConnection) GetPropertyDefault6() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockConnection) GetPropertyIP6Config() (IP6Config, error) {
	args := m.Called()
	return args.Get(0).(IP6Config), args.Error(1)
}

func (m *MockConnection) GetPropertyDHCP6Config() (DHCP6Config, error) {
	args := m.Called()
	return args.Get(0).(DHCP6Config), args.Error(1)
}

func (m *MockConnection) GetPropertyVPN() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockConnection) GetPropertyMaster() (Device, error) {
	args := m.Called()
	return args.Get(0).(Device), args.Error(1)
}

func (m *MockConnection) SubscribeState(receiver chan StateChange, exit chan struct{}) (err error) {
	args := m.Called(receiver, exit)
	return args.Error(0)
}

func (m *MockConnection) GetPath() dbus.ObjectPath {
	args := m.Called()
	return args.Get(0).(dbus.ObjectPath)
}

func (m *MockConnection) Update(settings ConnectionSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockConnection) UpdateUnsaved(settings ConnectionSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockConnection) Delete() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) GetSettings() (ConnectionSettings, error) {
	args := m.Called()
	return args.Get(0).(ConnectionSettings), args.Error(1)
}

func (m *MockConnection) GetSecrets(settingName string) (ConnectionSettings, error) {
	args := m.Called(settingName)
	return args.Get(0).(ConnectionSettings), args.Error(1)
}

func (m *MockConnection) ClearSecrets() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) Save() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) GetPropertyUnsaved() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockConnection) GetPropertyFlags() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockConnection) GetPropertyFilename() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConnection) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
