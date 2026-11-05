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

type MockModem struct {
	mock.Mock
}

func (m *MockModem) GetObjectPath() dbus.ObjectPath {
	args := m.Called()
	result, _ := args.Get(0).(dbus.ObjectPath)
	return result
}

func (m *MockModem) GetSimpleModem() (mm.ModemSimple, error) {
	panic("unimplemented")
}

func (m *MockModem) Get3gpp() (mm.Modem3gpp, error) {
	panic("unimplemented")
}

func (m *MockModem) GetCdma() (mm.ModemCdma, error) {
	panic("unimplemented")
}

func (m *MockModem) GetTime() (mm.ModemTime, error) {
	panic("unimplemented")
}

func (m *MockModem) GetFirmware() (mm.ModemFirmware, error) {
	panic("unimplemented")
}

func (m *MockModem) GetSignal() (mm.ModemSignal, error) {
	panic("unimplemented")
}

func (m *MockModem) GetOma() (mm.ModemOma, error) {
	panic("unimplemented")
}

func (m *MockModem) GetLocation() (mm.ModemLocation, error) {
	panic("unimplemented")
}

func (m *MockModem) GetMessaging() (mm.ModemMessaging, error) {
	panic("unimplemented")
}

func (m *MockModem) GetVoice() (mm.ModemVoice, error) {
	panic("unimplemented")
}

func (m *MockModem) Enable() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockModem) Disable() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockModem) CreateBearer(property mm.BearerProperty) (mm.Bearer, error) {
	args := m.Called(property)
	if b, ok := args.Get(0).(mm.Bearer); ok {
		return b, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockModem) DeleteBearer(bearer mm.Bearer) error {
	args := m.Called(bearer)
	return args.Error(0)
}

func (m *MockModem) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockModem) FactoryReset(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockModem) SetPowerState(state mm.MMModemPowerState) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockModem) SetCurrentCapabilities(caps []mm.MMModemCapability) error {
	args := m.Called(caps)
	return args.Error(0)
}

func (m *MockModem) SetCurrentModes(mode mm.Mode) error {
	args := m.Called(mode)
	return args.Error(0)
}

func (m *MockModem) SetCurrentBands(bands []mm.MMModemBand) error {
	args := m.Called(bands)
	return args.Error(0)
}

func (m *MockModem) Command(cmd string, timeout uint32) (string, error) {
	args := m.Called(cmd, timeout)
	return args.String(0), args.Error(1)
}

func (m *MockModem) GetSim() (mm.Sim, error) {
	panic("unimplemented")
}

func (m *MockModem) GetBearers() ([]mm.Bearer, error) {
	args := m.Called()
	if bearers, ok := args.Get(0).([]mm.Bearer); ok {
		return bearers, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockModem) GetSupportedCapabilities() ([][]mm.MMModemCapability, error) {
	panic("unimplemented")
}

func (m *MockModem) GetCurrentCapabilities() ([]mm.MMModemCapability, error) {
	panic("unimplemented")
}

func (m *MockModem) GetMaxBearers() (uint32, error) {
	panic("unimplemented")
}

func (m *MockModem) GetMaxActiveBearers() (uint32, error) {
	panic("unimplemented")
}

func (m *MockModem) GetManufacturer() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetModel() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetRevision() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetCarrierConfiguration() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetCarrierConfigurationRevision() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetHardwareRevision() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetDeviceIdentifier() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetDevice() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetDrivers() ([]string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetPlugin() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetPrimaryPort() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetPorts() ([]mm.Port, error) {
	panic("unimplemented")
}

func (m *MockModem) GetEquipmentIdentifier() (string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetUnlockRequired() (mm.MMModemLock, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.MMModemLock)
	return result, args.Error(1)
}

func (m *MockModem) GetUnlockRetries() ([]mm.Pair, error) {
	panic("unimplemented")
}

func (m *MockModem) GetState() (mm.MMModemState, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.MMModemState)
	return result, args.Error(1)
}

func (m *MockModem) GetStateFailedReason() (mm.MMModemStateFailedReason, error) {
	args := m.Called()
	result, _ := args.Get(0).(mm.MMModemStateFailedReason)
	return result, args.Error(1)
}

func (m *MockModem) GetAccessTechnologies() ([]mm.MMModemAccessTechnology, error) {
	args := m.Called()
	if techs, ok := args.Get(0).([]mm.MMModemAccessTechnology); ok {
		return techs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockModem) GetSignalQuality() (uint32, bool, error) {
	args := m.Called()
	percent, _ := args.Get(0).(uint32)
	recent, _ := args.Get(1).(bool)
	return percent, recent, args.Error(2)
}

func (m *MockModem) GetOwnNumbers() ([]string, error) {
	panic("unimplemented")
}

func (m *MockModem) GetPowerState() (mm.MMModemPowerState, error) {
	panic("unimplemented")
}

func (m *MockModem) GetSupportedModes() ([]mm.Mode, error) {
	panic("unimplemented")
}

func (m *MockModem) GetCurrentModes() (mm.Mode, error) {
	panic("unimplemented")
}

func (m *MockModem) GetSupportedBands() ([]mm.MMModemBand, error) {
	panic("unimplemented")
}

func (m *MockModem) GetCurrentBands() ([]mm.MMModemBand, error) {
	panic("unimplemented")
}

func (m *MockModem) GetSupportedIpFamilies() ([]mm.MMBearerIpFamily, error) {
	panic("unimplemented")
}

func (m *MockModem) MarshalJSON() ([]byte, error) {
	args := m.Called()
	result, _ := args.Get(0).([]byte)
	return result, args.Error(1)
}

func (m *MockModem) SubscribeStateChanged() <-chan *dbus.Signal {
	args := m.Called()
	result, _ := args.Get(0).(<-chan *dbus.Signal)
	return result
}

func (m *MockModem) ParseStateChanged(v *dbus.Signal) (mm.MMModemState, mm.MMModemState, mm.MMModemStateChangeReason, error) {
	args := m.Called(v)
	oldState, _ := args.Get(0).(mm.MMModemState)
	newState, _ := args.Get(1).(mm.MMModemState)
	reason, _ := args.Get(2).(mm.MMModemStateChangeReason)
	return oldState, newState, reason, args.Error(3)
}

func (m *MockModem) SubscribePropertiesChanged() <-chan *dbus.Signal {
	args := m.Called()
	result, _ := args.Get(0).(<-chan *dbus.Signal)
	return result
}

func (m *MockModem) ParsePropertiesChanged(v *dbus.Signal) (string, map[string]dbus.Variant, []string, error) {
	args := m.Called(v)
	interfaceName, _ := args.Get(0).(string)
	changedProps, _ := args.Get(1).(map[string]dbus.Variant)
	invalidatedProps, _ := args.Get(2).([]string)
	return interfaceName, changedProps, invalidatedProps, args.Error(3)
}

func (m *MockModem) Unsubscribe() {
	m.Called()
}
