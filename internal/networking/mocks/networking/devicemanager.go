/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/stretchr/testify/mock"
)

type MockDeviceManager struct {
	mock.Mock
}

func (m *MockDeviceManager) GetAvailableDevicesMap() (map[string]nm.Device, error) {
	args := m.Called()
	device := args.Get(0)
	if device == nil {
		return nil, args.Error(1)
	}
	return device.(map[string]nm.Device), args.Error(1)
}

func (m *MockDeviceManager) GetGSMDevice(name string) (nm.Device, error) {
	args := m.Called(name)
	device := args.Get(0)
	if device == nil {
		return nil, args.Error(1)
	}
	return device.(nm.Device), args.Error(1)
}
