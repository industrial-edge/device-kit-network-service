/*
 * Copyright © Siemens 2024 - 2025. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gonetworkmanager

import (
	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/mock"
)

// Mocking nm.DeviceWired interface
type MockDeviceWired struct {
	mock.Mock
}

func (m *MockDeviceWired) SubscribeState(receiver chan nm.DeviceStateChange, exit chan struct{}) error {
	args := m.Called(receiver, exit)
	return args.Error(0)
}

func (m *MockDeviceWired) SetPropertyAutoConnect(b bool) error {
	args := m.Called(b)
	return args.Error(1)
}

func (m *MockDeviceWired) GetPath() dbus.ObjectPath {
	args := m.Called()
	return args.Get(0).(dbus.ObjectPath)
}

func (m *MockDeviceWired) Reapply(connection nm.Connection, versionId uint64, flags uint32) error {
	args := m.Called(connection, versionId, flags)
	return args.Error(0)
}

func (m *MockDeviceWired) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDeviceWired) Delete() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDeviceWired) GetPropertyUdi() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyIpInterface() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyDriver() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyDriverVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyFirmwareVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyState() (nm.NmDeviceState, error) {
	args := m.Called()
	return args.Get(0).(nm.NmDeviceState), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyIP4Config() (nm.IP4Config, error) {
	args := m.Called()
	return args.Get(0).(nm.IP4Config), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyDHCP4Config() (nm.DHCP4Config, error) {
	args := m.Called()
	return args.Get(0).(nm.DHCP4Config), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyIP6Config() (nm.IP6Config, error) {
	args := m.Called()
	return args.Get(0).(nm.IP6Config), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyDHCP6Config() (nm.DHCP6Config, error) {
	args := m.Called()
	return args.Get(0).(nm.DHCP6Config), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyManaged() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWired) SetPropertyManaged(b bool) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockDeviceWired) GetPropertyAutoConnect() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyFirmwareMissing() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyNmPluginMissing() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyDeviceType() (nm.NmDeviceType, error) {
	args := m.Called()
	return args.Get(0).(nm.NmDeviceType), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyAvailableConnections() ([]nm.Connection, error) {
	args := m.Called()
	return args.Get(0).([]nm.Connection), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyPhysicalPortId() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyMtu() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyReal() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyIp4Connectivity() (nm.NmConnectivity, error) {
	args := m.Called()
	return args.Get(0).(nm.NmConnectivity), args.Error(1)
}

func (m *MockDeviceWired) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyPermHwAddress() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertySpeed() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyS390Subchannels() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyCarrier() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyActiveConnection() (nm.ActiveConnection, error) {
	args := m.Called()
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyHwAddress() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWired) GetPropertyInterface() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
