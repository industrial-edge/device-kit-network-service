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

type MockModem3gpp struct {
	mock.Mock
}

func (m *MockModem3gpp) GetObjectPath() dbus.ObjectPath {
	args := m.Called()
	result, _ := args.Get(0).(dbus.ObjectPath)
	return result
}

func (m *MockModem3gpp) GetUssd() (mm.Ussd, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) Register(operatorId string) error {
	args := m.Called(operatorId)
	return args.Error(0)
}

func (m *MockModem3gpp) Scan() ([]mm.Network3Gpp, error) {
	args := m.Called()
	if networks, ok := args.Get(0).([]mm.Network3Gpp); ok {
		return networks, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockModem3gpp) RequestScan() {
	m.Called()
}

func (m *MockModem3gpp) GetScanResults() (mm.NetworkScanResult, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.NetworkScanResult)
	return result, args.Error(1)
}

func (m *MockModem3gpp) SetEpsUeModeOperation(mode mm.MMModem3gppEpsUeModeOperation) error {
	args := m.Called(mode)
	return args.Error(0)
}

func (m *MockModem3gpp) SetInitialEpsBearerSettings(property mm.BearerProperty) error {
	args := m.Called(property)
	return args.Error(0)
}

func (m *MockModem3gpp) GetImei() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockModem3gpp) GetRegistrationState() (mm.MMModem3gppRegistrationState, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.MMModem3gppRegistrationState)
	return result, args.Error(1)
}

func (m *MockModem3gpp) GetOperatorCode() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockModem3gpp) GetMcc() (string, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) GetMnc() (string, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) GetOperatorName() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockModem3gpp) GetEnabledFacilityLocks() ([]mm.MMModem3gppFacility, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) GetEpsUeModeOperation() (mm.MMModem3gppEpsUeModeOperation, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) GetPco() ([]mm.RawPcoData, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) GetInitialEpsBearer() (mm.Bearer, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) GetInitialEpsBearerSettings() (mm.BearerProperty, error) {
	panic("unimplemented")
}

func (m *MockModem3gpp) MarshalJSON() ([]byte, error) {
	args := m.Called()
	result, _ := args.Get(0).([]byte)
	return result, args.Error(1)
}
