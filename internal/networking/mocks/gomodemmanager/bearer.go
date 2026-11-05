/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gomodemmanager

import (
	"github.com/godbus/dbus/v5"
	mm "github.com/maltegrosse/go-modemmanager"
	"github.com/stretchr/testify/mock"
)

type MockBearer struct {
	mock.Mock
}

func (m *MockBearer) GetObjectPath() dbus.ObjectPath {
	args := m.Called()
	result, _ := args.Get(0).(dbus.ObjectPath)
	return result
}

func (m *MockBearer) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBearer) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBearer) MarshalJSON() ([]byte, error) {
	args := m.Called()
	result, _ := args.Get(0).([]byte)
	return result, args.Error(1)
}

func (m *MockBearer) GetInterface() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockBearer) GetConnected() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockBearer) GetSuspended() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockBearer) GetIp4Config() (mm.BearerIpConfig, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.BearerIpConfig)
	return result, args.Error(1)
}

func (m *MockBearer) GetIp6Config() (mm.BearerIpConfig, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.BearerIpConfig)
	return result, args.Error(1)
}

func (m *MockBearer) GetStats() (mm.BearerStats, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.BearerStats)
	return result, args.Error(1)
}

func (m *MockBearer) GetIpTimeout() (uint32, error) {
	args := m.Called()
	result, _ := args.Get(0).(uint32)
	return result, args.Error(1)
}

func (m *MockBearer) GetBearerType() (mm.MMBearerType, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.MMBearerType)
	return result, args.Error(1)
}

func (m *MockBearer) GetProperties() (mm.BearerProperty, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.BearerProperty)
	return result, args.Error(1)
}

func (m *MockBearer) SubscribePropertiesChanged() <-chan *dbus.Signal {
	args := m.Called()
	result, _ := args.Get(0).(<-chan *dbus.Signal)
	return result
}

func (m *MockBearer) ParsePropertiesChanged(v *dbus.Signal) (string, map[string]dbus.Variant, []string, error) {
	args := m.Called(v)
	interfaceName, _ := args.Get(0).(string)
	changedProps, _ := args.Get(1).(map[string]dbus.Variant)
	invalidatedProps, _ := args.Get(2).([]string)
	return interfaceName, changedProps, invalidatedProps, args.Error(3)
}

func (m *MockBearer) Unsubscribe() {
	m.Called()
}
