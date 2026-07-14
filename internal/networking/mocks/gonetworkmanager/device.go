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

type MockDevice struct {
	mock.Mock
}

func (m MockDevice) Delete() error {
	panic("unimplemented")
}

func (m MockDevice) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m MockDevice) GetPath() dbus.ObjectPath {
	args := m.Called()
	return args.Get(0).(dbus.ObjectPath)
}

func (m MockDevice) GetPropertyActiveConnection() (nm.ActiveConnection, error) {
	args := m.Called()
	activeConn := args.Get(0)
	if activeConn == nil {
		return nil, args.Error(1)
	}
	return activeConn.(nm.ActiveConnection), args.Error(1)
}

func (m MockDevice) GetPropertyAutoConnect() (bool, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyAvailableConnections() ([]nm.Connection, error) {
	args := m.Called()
	return args.Get(0).([]nm.Connection), args.Error(1)
}

func (m MockDevice) GetPropertyDHCP4Config() (nm.DHCP4Config, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyDHCP6Config() (nm.DHCP6Config, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyDeviceType() (nm.NmDeviceType, error) {
	args := m.Called()
	return args.Get(0).(nm.NmDeviceType), args.Error(1)
}

func (m MockDevice) GetPropertyDriver() (string, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyDriverVersion() (string, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyFirmwareMissing() (bool, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyFirmwareVersion() (string, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyIP4Config() (nm.IP4Config, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyIP6Config() (nm.IP6Config, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyInterface() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m MockDevice) GetPropertyIp4Connectivity() (nm.NmConnectivity, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyIpInterface() (string, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyManaged() (bool, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyMtu() (uint32, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyNmPluginMissing() (bool, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyPhysicalPortId() (string, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyReal() (bool, error) {
	panic("unimplemented")
}

func (m MockDevice) GetPropertyState() (nm.NmDeviceState, error) {
	args := m.Called()
	return args.Get(0).(nm.NmDeviceState), args.Error(1)
}

func (m MockDevice) GetPropertyUdi() (string, error) {
	panic("unimplemented")
}

func (m MockDevice) MarshalJSON() ([]byte, error) {
	panic("unimplemented")
}

func (m MockDevice) Reapply(connection nm.Connection, versionId uint64, flags uint32) error {
	panic("unimplemented")
}

func (m MockDevice) SetPropertyAutoConnect(bool) error {
	panic("unimplemented")
}

func (m MockDevice) SetPropertyManaged(bool) error {
	panic("unimplemented")
}

func (m MockDevice) SubscribeState(receiver chan nm.DeviceStateChange, exit chan struct{}) (err error) {
	panic("unimplemented")
}
