/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
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
type MockDeviceWireless struct {
	mock.Mock
}

func (m *MockDeviceWireless) SubscribeState(receiver chan nm.DeviceStateChange, exit chan struct{}) error {
	args := m.Called(receiver, exit)
	return args.Error(0)
}

func (m *MockDeviceWireless) SetPropertyAutoConnect(b bool) error {
	args := m.Called(b)
	return args.Error(1)
}

func (m *MockDeviceWireless) GetPath() dbus.ObjectPath {
	args := m.Called()
	return args.Get(0).(dbus.ObjectPath)
}

func (m *MockDeviceWireless) Reapply(connection nm.Connection, versionId uint64, flags uint32) error {
	args := m.Called(connection, versionId, flags)
	return args.Error(0)
}

func (m *MockDeviceWireless) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDeviceWireless) Delete() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDeviceWireless) GetPropertyUdi() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyIpInterface() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyDriver() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyDriverVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyFirmwareVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyState() (nm.NmDeviceState, error) {
	args := m.Called()
	return args.Get(0).(nm.NmDeviceState), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyIP4Config() (nm.IP4Config, error) {
	args := m.Called()
	return args.Get(0).(nm.IP4Config), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyDHCP4Config() (nm.DHCP4Config, error) {
	args := m.Called()
	return args.Get(0).(nm.DHCP4Config), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyIP6Config() (nm.IP6Config, error) {
	args := m.Called()
	return args.Get(0).(nm.IP6Config), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyDHCP6Config() (nm.DHCP6Config, error) {
	args := m.Called()
	return args.Get(0).(nm.DHCP6Config), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyManaged() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWireless) SetPropertyManaged(b bool) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockDeviceWireless) GetPropertyAutoConnect() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyFirmwareMissing() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyNmPluginMissing() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyDeviceType() (nm.NmDeviceType, error) {
	args := m.Called()
	return args.Get(0).(nm.NmDeviceType), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyAvailableConnections() ([]nm.Connection, error) {
	args := m.Called()
	return args.Get(0).([]nm.Connection), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyPhysicalPortId() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyMtu() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyReal() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyIp4Connectivity() (nm.NmConnectivity, error) {
	args := m.Called()
	return args.Get(0).(nm.NmConnectivity), args.Error(1)
}

func (m *MockDeviceWireless) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyPermHwAddress() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertySpeed() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyS390Subchannels() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyCarrier() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyActiveConnection() (nm.ActiveConnection, error) {
	args := m.Called()
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyHwAddress() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyInterface() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
func (m *MockDeviceWireless) GetPropertyMode() (nm.Nm80211Mode, error) {
	args := m.Called()
	return args.Get(0).(nm.Nm80211Mode), args.Error(1)
}

func (m *MockDeviceWireless) GetPropertyBitrate() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}
func (m *MockDeviceWireless) GetPropertyAccessPoints() ([]nm.AccessPoint, error) {
	args := m.Called()
	return args.Get(0).([]nm.AccessPoint), args.Error(1)
}
func (m *MockDeviceWireless) GetPropertyActiveAccessPoint() (nm.AccessPoint, error) {
	args := m.Called()
	return args.Get(0).(nm.AccessPoint), args.Error(1)
}
func (m *MockDeviceWireless) GetPropertyWirelessCapabilities() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}
func (m *MockDeviceWireless) GetPropertyLastScan() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDeviceWireless) GetAccessPoints() ([]nm.AccessPoint, error) {
	args := m.Called()
	return args.Get(0).([]nm.AccessPoint), args.Error(1)
}

func (m *MockDeviceWireless) GetAllAccessPoints() ([]nm.AccessPoint, error) {
	args := m.Called()
	return args.Get(0).([]nm.AccessPoint), args.Error(1)
}

func (m *MockDeviceWireless) RequestScan() error {
	args := m.Called()
	return args.Error(0)
}
