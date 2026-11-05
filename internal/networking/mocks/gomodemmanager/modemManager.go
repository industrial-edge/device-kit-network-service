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

type MockModemManager struct {
	mock.Mock
}

func (m *MockModemManager) ScanDevices() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockModemManager) GetModems() ([]mm.Modem, error) {
	args := m.Called()
	if modems, ok := args.Get(0).([]mm.Modem); ok {
		return modems, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockModemManager) SetLogging(level mm.MMLoggingLevel) error {
	args := m.Called(level)
	return args.Error(0)
}

func (m *MockModemManager) ReportKernelEvent(props mm.EventProperties) error {
	args := m.Called(props)
	return args.Error(0)
}

func (m *MockModemManager) InhibitDevice(uid string, inhibit bool) error {
	args := m.Called(uid, inhibit)
	return args.Error(0)
}

func (m *MockModemManager) GetVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockModemManager) MarshalJSON() ([]byte, error) {
	args := m.Called()
	result, _ := args.Get(0).([]byte)
	return result, args.Error(1)
}

func (m *MockModemManager) SubscribePropertiesChanged() <-chan *dbus.Signal {
	args := m.Called()
	result, _ := args.Get(0).(<-chan *dbus.Signal)
	return result
}

func (m *MockModemManager) ParsePropertiesChanged(v *dbus.Signal) (string, map[string]dbus.Variant, []string, error) {
	args := m.Called(v)
	interfaceName, _ := args.Get(0).(string)
	changedProps, _ := args.Get(1).(map[string]dbus.Variant)
	invalidatedProps, _ := args.Get(2).([]string)
	return interfaceName, changedProps, invalidatedProps, args.Error(3)
}

func (m *MockModemManager) Unsubscribe() {
	m.Called()
}
