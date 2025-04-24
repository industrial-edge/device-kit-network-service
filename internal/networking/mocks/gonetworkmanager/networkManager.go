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

type MockNetworkManager struct {
	mock.Mock
}

func (m MockNetworkManager) Reload(flags uint32) error {
	args := m.Called(flags)
	return args.Error(0)
}

func (m MockNetworkManager) GetDevices() ([]nm.Device, error) {
	args := m.Called()
	return args.Get(0).([]nm.Device), args.Error(1)
}

func (m MockNetworkManager) GetAllDevices() ([]nm.Device, error) {
	args := m.Called()
	return args.Get(0).([]nm.Device), args.Error(1)
}

func (m MockNetworkManager) GetDeviceByIpIface(interfaceId string) (nm.Device, error) {
	args := m.Called(interfaceId)
	return args.Get(0).(nm.Device), args.Error(1)
}

func (m MockNetworkManager) ActivateConnection(connection nm.Connection, device nm.Device, specificObject *dbus.Object) (nm.ActiveConnection, error) {
	args := m.Called(connection, device, specificObject)
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) AddAndActivateConnection(connection map[string]map[string]interface{}, device nm.Device) (nm.ActiveConnection, error) {
	args := m.Called(connection, device)
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) ActivateWirelessConnection(connection nm.Connection, device nm.Device, accessPoint nm.AccessPoint) (nm.ActiveConnection, error) {
	args := m.Called(connection, device, accessPoint)
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) AddAndActivateWirelessConnection(connection map[string]map[string]interface{}, device nm.Device, accessPoint nm.AccessPoint) (nm.ActiveConnection, error) {
	args := m.Called(connection, device, accessPoint)
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) DeactivateConnection(connection nm.ActiveConnection) error {
	args := m.Called(connection)
	return args.Error(0)
}

func (m MockNetworkManager) Sleep(sleepNWake bool) error {
	args := m.Called(sleepNWake)
	return args.Error(0)
}

func (m MockNetworkManager) Enable(enableNDisable bool) error {
	args := m.Called(enableNDisable)
	return args.Error(0)
}

func (m MockNetworkManager) CheckConnectivity() error {
	args := m.Called()
	return args.Error(0)
}

func (m MockNetworkManager) State() (nm.NmState, error) {
	args := m.Called()
	return args.Get(0).(nm.NmState), args.Error(1)
}

func (m MockNetworkManager) CheckpointCreate(devices []nm.Device, rollbackTimeout uint32, flags uint32) (nm.Checkpoint, error) {
	args := m.Called(devices, rollbackTimeout, flags)
	return args.Get(0).(nm.Checkpoint), args.Error(1)
}

func (m MockNetworkManager) CheckpointDestroy(checkpoint nm.Checkpoint) error {
	args := m.Called(checkpoint)
	return args.Error(0)
}

func (m MockNetworkManager) CheckpointRollback(checkpoint nm.Checkpoint) (result map[dbus.ObjectPath]nm.NmRollbackResult, err error) {
	args := m.Called(checkpoint)
	return args.Get(0).(map[dbus.ObjectPath]nm.NmRollbackResult), args.Error(1)
}

func (m MockNetworkManager) CheckpointAdjustRollbackTimeout(checkpoint nm.Checkpoint, addTimeout uint32) error {
	args := m.Called(checkpoint, addTimeout)
	return args.Error(0)
}

func (m MockNetworkManager) GetPropertyDevices() ([]nm.Device, error) {
	args := m.Called()
	return args.Get(0).([]nm.Device), args.Error(1)
}

func (m MockNetworkManager) GetPropertyAllDevices() ([]nm.Device, error) {
	args := m.Called()
	return args.Get(0).([]nm.Device), args.Error(1)
}

func (m MockNetworkManager) GetPropertyCheckpoints() ([]nm.Checkpoint, error) {
	args := m.Called()
	return args.Get(0).([]nm.Checkpoint), args.Error(1)
}

func (m MockNetworkManager) GetPropertyNetworkingEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyWirelessEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) SetPropertyWirelessEnabled(b bool) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m MockNetworkManager) GetPropertyWirelessHardwareEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyWwanEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyWwanHardwareEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyWimaxEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyWimaxHardwareEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyActiveConnections() ([]nm.ActiveConnection, error) {
	args := m.Called()
	return args.Get(0).([]nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) GetPropertyPrimaryConnection() (nm.ActiveConnection, error) {
	args := m.Called()
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) GetPropertyPrimaryConnectionType() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m MockNetworkManager) GetPropertyMetered() (nm.NmMetered, error) {
	args := m.Called()
	return args.Get(0).(nm.NmMetered), args.Error(1)
}

func (m MockNetworkManager) GetPropertyActivatingConnection() (nm.ActiveConnection, error) {
	args := m.Called()
	return args.Get(0).(nm.ActiveConnection), args.Error(1)
}

func (m MockNetworkManager) GetPropertyStartup() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyVersion() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m MockNetworkManager) GetPropertyCapabilities() ([]nm.NmCapability, error) {
	args := m.Called()
	return args.Get(0).([]nm.NmCapability), args.Error(1)
}

func (m MockNetworkManager) GetPropertyState() (nm.NmState, error) {
	args := m.Called()
	return args.Get(0).(nm.NmState), args.Error(1)
}

func (m MockNetworkManager) GetPropertyConnectivity() (nm.NmConnectivity, error) {
	args := m.Called()
	return args.Get(0).(nm.NmConnectivity), args.Error(1)
}

func (m MockNetworkManager) GetPropertyConnectivityCheckAvailable() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) GetPropertyConnectivityCheckEnabled() (bool, error) {
	args := m.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (m MockNetworkManager) Subscribe() <-chan *dbus.Signal {
	args := m.Called()
	return args.Get(0).(<-chan *dbus.Signal)
}

func (m MockNetworkManager) Unsubscribe() {
	m.Called()
}

func (m MockNetworkManager) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
