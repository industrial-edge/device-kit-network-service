/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package devicemanager

import (
	"errors"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/stretchr/testify/assert"
)

func TestDeviceManager_GetAvailableDevicesMap_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameOne := "eth0"
	deviceNameTwo := "gsm0"
	deviceNameThree := "wlan0"

	mockDeviceOne.On("GetPropertyInterface").Return(deviceNameOne, nil).Once()
	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()
	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()

	mockDeviceThree.On("GetPropertyInterface").Return(deviceNameThree, nil).Once()
	mockDeviceThree.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeWifi, nil).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Act
	got, gotErr := d.GetAvailableDevicesMap()

	// Assert
	assert.NoError(t, gotErr, "Expected no error when getting available devices map")
	expected := map[string]nm.Device{
		deviceNameOne: mockDeviceOne,
		deviceNameTwo: mockDeviceTwo,
	}
	assert.Equal(t, expected, got, "Expected device map does not match actual")
}

func TestDeviceManager_GetAvailableDevicesMap_GetPropertyDeviceType_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameOne := "eth0"
	deviceNameTwo := "gsm0"
	deviceNameThree := "wlan0"

	mockDeviceOne.On("GetPropertyInterface").Return(deviceNameOne, nil).Once()
	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()
	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()

	mockDeviceThree.On("GetPropertyInterface").Return(deviceNameThree, nil).Once()
	mockDeviceThree.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeUnknown, errors.New("mock error")).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Act
	got, gotErr := d.GetAvailableDevicesMap()

	// Assert
	assert.Error(t, gotErr, "Expected error when getting available devices map")
	assert.Nil(t, got, "unexpected non-nil device map")
}

func TestDeviceManager_GetAvailableDevicesMap_GetPropertyInterface_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameOne := "eth0"
	deviceNameTwo := "gsm0"
	deviceNameThree := "wlan0"

	mockDeviceOne.On("GetPropertyInterface").Return(deviceNameOne, nil).Once()
	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()
	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()

	mockDeviceThree.On("GetPropertyInterface").Return(deviceNameThree, errors.New("mock error")).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Act
	got, gotErr := d.GetAvailableDevicesMap()

	// Assert
	assert.Error(t, gotErr, "Expected error when getting available devices map")
	assert.Nil(t, got, "unexpected non-nil device map")
}

func TestDeviceManager_GetAvailableDevicesMap_GetDevices_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockNetworkManager.On("GetDevices").Return([]nm.Device{}, errors.New("mock error"))

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Act
	got, gotErr := d.GetAvailableDevicesMap()

	// Assert
	assert.Error(t, gotErr, "Expected error when getting available devices map")
	assert.Nil(t, got, "unexpected non-nil device map")
}

func TestNewDeviceManager(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	// Act
	got := NewDeviceManager(mockNetworkManager)

	// Assert
	assert.NotNil(t, got, "Expected non-nil DeviceManager instance")
}

func TestDeviceManager_GetGSMDevice_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameTwo := "gsm0"
	deviceNameThree := "gsm1"

	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()

	mockDeviceThree.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceThree.On("GetPropertyInterface").Return(deviceNameThree, nil).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Arrange
	got, gotErr := d.GetGSMDevice(deviceNameThree)

	// Assert
	assert.NoError(t, gotErr, "Expected no error when getting GSM device")
	assert.Equal(t, mockDeviceThree, got, "Expected GSM device does not match actual")
}

func TestDeviceManager_GetGSMDevice_GetPropertyInterface_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameTwo := "gsm0"
	deviceNameThree := "gsm1"

	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()

	mockDeviceThree.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceThree.On("GetPropertyInterface").Return("", errors.New("mock error")).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Arrange
	got, gotErr := d.GetGSMDevice(deviceNameThree)

	// Assert
	assert.Error(t, gotErr, "Expected error when getting GSM device")
	assert.Nil(t, got, "unexpected non-nil GSM device")
}

func TestDeviceManager_GetGSMDevice_GetPropertyDeviceType_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameTwo := "gsm0"
	deviceNameThree := "gsm1"

	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()

	mockDeviceThree.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeUnknown, errors.New("mock error")).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Arrange
	got, gotErr := d.GetGSMDevice(deviceNameThree)

	// Assert
	assert.Error(t, gotErr, "Expected error when getting GSM device")
	assert.Nil(t, got, "unexpected non-nil GSM device")
}

func TestDeviceManager_GetGSMDevice_DeviceNotExists(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	deviceNameTwo := "gsm0"
	deviceNameThree := "gsm1"

	mockDeviceOne.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil).Once()

	mockDeviceTwo.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceTwo.On("GetPropertyInterface").Return(deviceNameTwo, nil).Once()

	mockDeviceThree.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil).Once()
	mockDeviceThree.On("GetPropertyInterface").Return(deviceNameThree, nil).Once()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{
		mockDeviceOne,
		mockDeviceTwo,
		mockDeviceThree,
	}, nil)

	d := &DeviceManager{
		nm: mockNetworkManager,
	}

	// Arrange
	got, gotErr := d.GetGSMDevice("gsm2")

	// Assert
	assert.Error(t, gotErr, "Expected error when getting GSM device")
	assert.EqualError(t, gotErr, "Failed to retrieve GSM device: gsm2", "unexpected error message mismatch")
	assert.Nil(t, got, "unexpected non-nil GSM device")
}
