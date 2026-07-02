/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/mocks"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	mocksnetworking "networkservice/internal/networking/mocks/networking"

	nm "github.com/Wifx/gonetworkmanager/v2"
	nmanager "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// dict is a type alias for map[string]interface{} used in tests
type dict map[string]interface{}

// NewGSMHandler test cases

func Test_NewGSMHandler_ReturnsNilWhenNewSettingsFails(t *testing.T) {
	// Arrange
	mockNM := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock newSettings to return error
	patches.ApplyFunc(newSettings, func() (nm.Settings, error) {
		return nil, errors.New("failed to create NetworkManager settings")
	})

	// Act
	result, err := NewGSMHandler(mockNM)

	// Assert
	assert.Nil(t, result, "Expected nil when newSettings fails")
	assert.NotNil(t, err, "Expected error when newSettings fails")
}

func Test_NewGSMHandler_ReturnsInitializedGSMConnectionOnSuccess(t *testing.T) {
	// Arrange
	mockNM := &mockgnm.MockNetworkManager{}
	mockSettings := &mockgnm.MockSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock newSettings to return successful mock
	patches.ApplyFunc(newSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	// Act
	result, err := NewGSMHandler(mockNM)

	// Assert
	assert.NotNil(t, result, "Expected non-nil GSMConnection when newSettings succeeds")
	assert.Nil(t, err, "Expected nil error when newSettings succeeds")
	assert.IsType(t, &GSMConnection{}, result, "Expected result to be of type *GSMConnection")

	// Verify the GSMConnection struct is properly initialized
	gsmConn := result
	assert.Equal(t, mockNM, gsmConn.networkManager, "Expected networkManager to be set correctly")
	assert.Equal(t, mockSettings, gsmConn.nmSettings, "Expected nmSettings to be set correctly")
	assert.NotNil(t, gsmConn.connManager, "Expected connManager to be initialized")
	assert.NotNil(t, gsmConn.fileUtils, "Expected fileUtils to be initialized")

}

func Test_NewGSMHandler_CallsNewConnectionManagerWithCorrectParameter(t *testing.T) {
	// Arrange
	mockNM := &mockgnm.MockNetworkManager{}
	mockSettings := &mockgnm.MockSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock newSettings
	patches.ApplyFunc(newSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	// Act
	result, err := NewGSMHandler(mockNM)

	// Assert
	assert.NotNil(t, result, "Expected non-nil GSMConnection")
	assert.Nil(t, err, "Expected nil error")
	gsmConn := result
	// Verify connManager was created (we can't easily mock the direct function call)
	assert.NotNil(t, gsmConn.connManager, "Expected connManager to be initialized")
}

func Test_CreateConnection_ReturnsErrorWhenNoGSMInterfaceFound(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when creating GSM connection")
	assert.Equal(t, err.Error(), "failed to get GSM interface: GSM connection selected but GSM config is missing")
}

func Test_CreateConnection_ReturnsErrorWhenInvalidGSMSettingsFound(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "test gsm"}}}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when creating Invalid GSM connection")
	assert.Equal(t, err.Error(), "GSM configuration validation error: invalid PIN format")
}

func Test_CreateConnection_ReturnsErrorWhenDevicesNotFound(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	// Mock GetDevices to return a list of devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nmanager.NetworkManager) ([]nmanager.Device, error) {
		return []nmanager.Device{}, errors.New("Failed to get devices: no devices found")
	})
	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when creating Invalid GSM connection")
	assert.Equal(t, err.Error(), "Failed to get devices: no devices found")
}

func Test_CreateConnection_ReturnsErrorWhenGSMDevicesNotFound(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	mockDevice1 := &mockgnm.MockDeviceWired{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	// Mock GetDevices to return a list of devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nmanager.NetworkManager) ([]nmanager.Device, error) {
		return []nmanager.Device{mockDevice1, mockDevice2}, nil
	})
	// Mock GetPropertyDeviceType to return Ethernet for both devices
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPropertyDeviceType", func(_ nmanager.Device) (nmanager.NmDeviceType, error) {
		return nmanager.NmDeviceTypeEthernet, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice2), "GetPropertyDeviceType", func(_ nmanager.Device) (nmanager.NmDeviceType, error) {
		return nmanager.NmDeviceTypeEthernet, nil
	})
	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when creating Invalid GSM connection")
	assert.Equal(t, err.Error(), "GSM device lookup failed: no GSM modem device found among network devices")
}

func Test_CreateConnection_ReturnsErrorWhenGSMDeviceInterfaceErrorOccurred(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "testgsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	mockDevice1 := &mockgnm.MockDeviceWireless{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	// Mock GetDevices to return a list of devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nmanager.NetworkManager) ([]nmanager.Device, error) {
		return []nmanager.Device{mockDevice1, mockDevice2}, nil
	})
	// Mock GetPropertyDeviceType to return Ethernet for both devices
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPropertyDeviceType", func(_ nmanager.Device) (nmanager.NmDeviceType, error) {
		return nmanager.NmDeviceTypeModem, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice2), "GetPropertyDeviceType", func(_ nmanager.Device) (nmanager.NmDeviceType, error) {
		return nmanager.NmDeviceTypeEthernet, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPath", func(_ nmanager.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/2"
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPropertyInterface", func(_ nmanager.Device) (string, error) {
		return "", errors.New("no interface found")
	})
	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when creating Invalid GSM connection")
	assert.Equal(t, err.Error(), "failed to prepare GSM connection: could not get GSM device interface name: no interface found")
}

func Test_CreateConnection_ReturnsErrorWhenDeleteExistingGSMConnectionsErrorOccurred(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	newSetting := &v1.ConnectionSettings{
		Name:           "testgsm",
		ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{
			Gsm: &v1.Interface_GsmConf{
				Apn: "my-internet",
				Pin: "1234",
			},
		},
	}

	mockDevice1 := &mockgnm.MockDevice{}
	mockDevice2 := &mockgnm.MockDevice{}
	mockDevices := []nmanager.Device{mockDevice1, mockDevice2}

	// mockSettings := &mockgnm.MockSettings{}
	mockConn := &mockgnm.MockConnection{}
	connections := []nmanager.Connection{mockConn}

	mockNetworkManager.On("GetDevices").Return(mockDevices, nil)

	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)

	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice2.On("GetPropertyInterface").Return("interfacetest", nil)

	patches.ApplyFunc(newSettings, func() (nmanager.Settings, error) {
		return mockSettings, nil
	})

	mockExistingConnSettings := nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "testgsm",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "gsm",
		},
	}

	mockSettings.On("ListConnections").Return(connections, nil)
	mockConn.On("GetSettings").Return(mockExistingConnSettings, nil)
	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))
	mockConn.On("Delete").Return(errors.New("GSM connection"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when delete existing GSM connection")
	assert.Equal(t, err.Error(), "failed to delete existing GSM connections: GSM connection")
}

func Test_CreateConnection_ReturnsErrorWhenListingConnectionErrorOccurred(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	newSetting := &v1.ConnectionSettings{
		Name:           "testgsm",
		ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{
			Gsm: &v1.Interface_GsmConf{
				Apn: "my-internet",
				Pin: "1234",
			},
		},
	}

	mockDevice1 := &mockgnm.MockDevice{}
	mockDevice2 := &mockgnm.MockDevice{}
	mockDevices := []nmanager.Device{mockDevice1, mockDevice2}
	mockConnection := &mockgnm.MockConnection{}

	mockSettings := &mockgnm.MockSettings{}

	mockNetworkManager.On("GetDevices").Return(mockDevices, nil)

	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)

	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice2.On("GetPropertyInterface").Return("interfacetest", nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(newSettings, func() (nmanager.Settings, error) {
		return mockSettings, nil
	})

	mockSettings.On("ListConnections").Return([]nmanager.Connection{}, errors.New("Failed to list connections"))
	mockSettings.On("AddConnection", mock.Anything, mock.Anything).Return(nil, errors.New("Adding Gsm connection failed"))
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)

	assert.NotNil(t, err, "CreateConnection should  return an error when adding GSM connection")
	assert.Equal(t, err.Error(), "failed to add GSM connection: Adding Gsm connection failed")
}

func Test_CreateConnection_ReturnsErrorWhenActivatingGSMConnection_Fails(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	newSetting := &v1.ConnectionSettings{
		Name:           "testgsm",
		ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{
			Gsm: &v1.Interface_GsmConf{
				Apn: "my-internet",
				Pin: "1234",
			},
		},
	}

	patches.ApplyFunc(newSettings, func() (nmanager.Settings, error) {
		return mockSettings, nil
	})

	mockDevice1 := &mockgnm.MockDeviceWireless{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	mockConnection := &mockgnm.MockConnection{}
	mockConn1 := &mockgnm.MockConnection{}
	connections := []nmanager.Connection{mockConn1}

	// Set up testify mocks
	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)

	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)

	mockDevice1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice1.On("GetPropertyInterface").Return("interfacetest", nil)

	mockSettings.On("ListConnections").Return(connections, nil)
	mockConn1.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "testgsm",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "gsm",
		},
	}, nil)
	mockConn1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConn1.On("Delete").Return(nil)

	mockConnSettings := nmanager.ConnectionSettings{
		common.ConnectionKey: make(map[string]interface{}),
		common.IPV4Key:       make(map[string]interface{}),
		common.GSMSetting:    make(map[string]interface{}),
	}
	mockSettings.On("AddConnection", mock.Anything, mock.Anything).Return(mockConnection, nil)
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice1).Return(errors.New("could not activate connection"))

	gsmConnection := &GSMConnection{
		networkManager:          mockNetworkManager,
		nmSettings:              mockSettings,
		connectionSettingFromNM: mockConnSettings,
		connManager:             mockConnectionManager,
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when acyivating  GSM connection")
}

func Test_CreateConnection_Success(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	newSetting := &v1.ConnectionSettings{
		Name:           "testgsm",
		ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{
			Gsm: &v1.Interface_GsmConf{
				Apn: "my-internet",
				Pin: "1234",
			},
		},
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(&GSMConnection{}), "writeGSMConfigToFile",
		func(_ *GSMConnection, _ *v1.ConnectionSettings_Gsm) error {
			return nil
		})

	mockDevice1 := &mockgnm.MockDeviceWireless{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	mockConnection := &mockgnm.MockConnection{}
	mockConn1 := &mockgnm.MockConnection{}
	connections := []nmanager.Connection{mockConn1}

	// Set up testify mocks
	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)

	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)

	mockDevice1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice1.On("GetPropertyInterface").Return("interfacetest", nil)

	mockSettings.On("ListConnections").Return(connections, nil)
	mockConn1.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "testgsm",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "gsm",
		},
	}, nil)
	mockConn1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConn1.On("Delete").Return(nil)

	mockConnSettings := nmanager.ConnectionSettings{
		common.ConnectionKey: make(map[string]interface{}),
		common.IPV4Key:       make(map[string]interface{}),
		common.GSMSetting:    make(map[string]interface{}),
	}
	mockSettings.On("AddConnection", mock.Anything, mock.Anything).Return(mockConnection, nil)
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice1).Return(nil)

	gsmConnection := &GSMConnection{
		networkManager:          mockNetworkManager,
		nmSettings:              mockSettings,
		connectionSettingFromNM: mockConnSettings,
		connManager:             mockConnectionManager,
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.Nil(t, err, "CreateConnection should not return an error when activating GSM connection")
}

func Test_CreateConnection_writeGSMConfigToFile_Error(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	newSetting := &v1.ConnectionSettings{
		Name:           "testgsm",
		ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{
			Gsm: &v1.Interface_GsmConf{
				Apn: "my-internet",
				Pin: "1234",
			},
		},
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(&GSMConnection{}), "writeGSMConfigToFile",
		func(_ *GSMConnection, _ *v1.ConnectionSettings_Gsm) error {
			return errors.New("file write error")
		})

	mockDevice1 := &mockgnm.MockDeviceWireless{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	mockConnection := &mockgnm.MockConnection{}
	mockConn1 := &mockgnm.MockConnection{}
	connections := []nmanager.Connection{mockConn1}

	// Set up testify mocks
	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)

	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)

	mockDevice1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice1.On("GetPropertyInterface").Return("interfacetest", nil)

	mockSettings.On("ListConnections").Return(connections, nil)
	mockConn1.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "testgsm",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "gsm",
		},
	}, nil)
	mockConn1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConn1.On("Delete").Return(nil)

	mockConnSettings := nmanager.ConnectionSettings{
		common.ConnectionKey: make(map[string]interface{}),
		common.IPV4Key:       make(map[string]interface{}),
		common.GSMSetting:    make(map[string]interface{}),
	}
	mockSettings.On("AddConnection", mock.Anything, mock.Anything).Return(mockConnection, nil)
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice1).Return(nil)

	gsmConnection := &GSMConnection{
		networkManager:          mockNetworkManager,
		nmSettings:              mockSettings,
		connectionSettingFromNM: mockConnSettings,
		connManager:             mockConnectionManager,
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.Nil(t, err, "CreateConnection should not return an error when activating GSM connection")
}

func Test_RemoveConnection_ReturnsErrorWhenNoGSMInterfaceFound(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm"}

	err := gsmConnection.RemoveConnection(newSetting)
	assert.NotNil(t, err, "RemoveConnection should  return an error when removing GSM connection")
	assert.Equal(t, err.Error(), "minimal parameters(connection type gsm and name) required for GSM connection removal are missing")
}

func Test_RemoveConnection_ReturnsNoErrorInConnectionListing(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	patches.ApplyMethod(reflect.TypeOf(mockSettings), "ListConnections", func(_ nmanager.Settings) ([]nmanager.Connection, error) {
		return nil, errors.New("Failed to list connections")
	})
	err := gsmConnection.RemoveConnection(newSetting)
	assert.Nil(t, err, nil)
	// assert.Equal(t, err.Error(), "connection not found: testgsm_dhcp")
}

func Test_RemoveConnection_ReturnsNoError_WhenConnectionNotFound(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	mockConn1 := &mockgnm.MockConnection{}
	mockConnections := []nmanager.Connection{mockConn1}
	gsmConnection.networkManager = mockNetworkManager

	patches.ApplyMethod(reflect.TypeOf(mockSettings), "ListConnections", func(_ nmanager.Settings) ([]nmanager.Connection, error) {
		return mockConnections, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockConn1), "GetSettings", func(_ nmanager.Connection) (nmanager.ConnectionSettings, error) {
		connection := nmanager.ConnectionSettings{
			"connection": map[string]interface{}{
				"id":   "testgsm1",
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"type": "gsm",
			},
		}
		return connection, nil

	})

	err := gsmConnection.RemoveConnection(newSetting)
	assert.Nil(t, err, nil)

}

func Test_RemoveConnection_ReturnsErrorInConnectionDelete(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	testConnSettings := &v1.ConnectionSettings{
		Name:           "testgsm",
		ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{
			Gsm: &v1.Interface_GsmConf{
				Apn: "my-internet",
				Pin: "1234",
			},
		},
	}

	mockConn := &mockgnm.MockConnection{}
	mockConnections := []nmanager.Connection{mockConn}

	mockConnSettings := nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "testgsm",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "gsm",
		},
	}

	mockSettings.On("ListConnections").Return(mockConnections, nil)
	mockConn.On("GetSettings").Return(mockConnSettings, nil)
	mockConn.On("Delete").Return(errors.New("failed to delete connection"))

	g := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
	}

	err := g.RemoveConnection(testConnSettings)
	assert.NotNil(t, err, "RemoveConnection should  return an error when removing GSM connection")
	assert.Equal(t, err.Error(), "GSM connection deletion failed: failed to delete connection")
}

func Test_RemoveConnection_SuccessInConnectionDelete(t *testing.T) {

	mockNetworkManager := &mockgnm.MockNetworkManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockSettings := &mockgnm.MockSettings{}
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	gsmConnection, _ := NewGSMHandler(mockNetworkManager)

	newSetting := &v1.ConnectionSettings{
		Name: "test-gsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	mockConn1 := &mockgnm.MockConnection{}
	mockConnections := []nmanager.Connection{mockConn1}

	patches.ApplyMethod(reflect.TypeOf(mockSettings), "ListConnections", func(_ nmanager.Settings) ([]nmanager.Connection, error) {
		return mockConnections, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockConn1), "GetSettings", func(_ nmanager.Connection) (nmanager.ConnectionSettings, error) {
		connection := nmanager.ConnectionSettings{
			"connection": map[string]interface{}{
				"id":   "testgsm",
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"type": "gsm",
			},
		}
		return connection, nil

	})
	patches.ApplyMethod(reflect.TypeOf(mockConn1), "Delete", func(_ nmanager.Connection) error {
		return nil
	})
	err := gsmConnection.RemoveConnection(newSetting)
	assert.Nil(t, err, nil)
}

// TestWriteGSMConfigToFile_Success tests successful GSM config file writing scenarios
func TestWriteGSMConfigToFile_SuccessWithAllFields(t *testing.T) {
	gsmData := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.provider.com",
			Pin:      "1234",
			Username: "user123",
			Password: "pass456",
		},
	}

	expectedConfig := map[string]string{
		"Apn":      "internet.provider.com",
		"Pin":      "1234",
		"Username": "user123",
		"Password": "pass456",
	}
	mockNM := &mockgnm.MockNetworkManager{}
	gsmConn := &GSMConnection{
		networkManager: mockNM,
		connectionSettingFromNM: nmanager.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
		connManager: connectionmanager.NewConnectionManager(mockNM),
	}

	mockFileUtils := &mocks.MockFileSystem{}
	gsmConn.fileUtils = mockFileUtils

	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("RemoveAll", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)

	// Capture the data written to file
	var capturedData []byte
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.AnythingOfType("[]uint8"), os.FileMode(0600)).
		Run(func(args mock.Arguments) {
			capturedData = args.Get(1).([]byte)
		}).Return(nil)

	// Act
	err := gsmConn.writeGSMConfigToFile(gsmData)

	// Assert
	assert.NoError(t, err, "Should write GSM config without error")
	// Verify the JSON content
	var actualConfig map[string]string
	err = json.Unmarshal(capturedData, &actualConfig)
	assert.NoError(t, err, "Should unmarshal written JSON data")
	assert.Equal(t, expectedConfig, actualConfig, "Written config should match expected")

}

func TestWriteGSMConfigToFile_FailureInRemoveAll(t *testing.T) {
	gsmData := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.provider.com",
			Pin:      "1234",
			Username: "user123",
			Password: "pass456",
		},
	}

	mockNM := &mockgnm.MockNetworkManager{}
	gsmConn := &GSMConnection{
		networkManager: mockNM,
		connectionSettingFromNM: nmanager.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
		connManager: connectionmanager.NewConnectionManager(mockNM),
	}

	mockFileUtils := &mocks.MockFileSystem{}
	gsmConn.fileUtils = mockFileUtils

	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("RemoveAll", common.GSMConfigPath).Return(fmt.Errorf("failed to remove directory"))
	// mockFileUtils.On("MkdirAll", constants.GSMConfigParentPath, os.FileMode(0755)).Return(nil)

	// // Capture the data written to file
	// var capturedData []byte
	// mockFileUtils.On("WriteFile", constants.GSMConfigPath, mock.AnythingOfType("[]uint8"), os.FileMode(0600)).
	// 	Run(func(args mock.Arguments) {
	// 		capturedData = args.Get(1).([]byte)
	// 	}).Return(nil)

	// Act
	err := gsmConn.writeGSMConfigToFile(gsmData)

	// Assert
	assert.Error(t, err, "Should return error when RemoveAll fails")
	assert.Contains(t, err.Error(), "failed to remove existing GSM config file")

}

func TestWriteGSMConfigToFile_FailureInCreateDir(t *testing.T) {
	gsmData := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.provider.com",
			Pin:      "1234",
			Username: "user123",
			Password: "pass456",
		},
	}

	mockNM := &mockgnm.MockNetworkManager{}
	gsmConn := &GSMConnection{
		networkManager: mockNM,
		connectionSettingFromNM: nmanager.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
		connManager: connectionmanager.NewConnectionManager(mockNM),
	}

	mockFileUtils := &mocks.MockFileSystem{}
	gsmConn.fileUtils = mockFileUtils

	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("RemoveAll", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(fmt.Errorf("failed to create directory"))

	// Act
	err := gsmConn.writeGSMConfigToFile(gsmData)

	// Assert
	assert.Error(t, err, "Should return error when MkdirAll fails")
	assert.Equal(t, "failed to create GSM config directory: failed to create directory", err.Error())

}

func TestWriteGSMConfigToFile_FailureInWriteFile(t *testing.T) {
	gsmData := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.provider.com",
			Pin:      "1234",
			Username: "user123",
			Password: "pass456",
		},
	}

	mockNM := &mockgnm.MockNetworkManager{}
	gsmConn := &GSMConnection{
		networkManager: mockNM,
		connectionSettingFromNM: nmanager.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
		connManager: connectionmanager.NewConnectionManager(mockNM),
	}

	mockFileUtils := &mocks.MockFileSystem{}
	gsmConn.fileUtils = mockFileUtils

	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("RemoveAll", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.AnythingOfType("[]uint8"), os.FileMode(0600)).Return(fmt.Errorf("failed to write file"))

	// Act
	err := gsmConn.writeGSMConfigToFile(gsmData)

	// Assert
	assert.Error(t, err, "Should return error when WriteFile fails")
	assert.Equal(t, "failed to write GSM config to file: failed to write file", err.Error())
}

// SetConnectionDetailsWithGSM test cases

func TestSetConnectionDetailsWithGSM_WithValidParameters(t *testing.T) {
	// Arrange
	connection := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.test",
			Pin:      "1234",
			Username: "testuser",
			Password: "testpass",
		},
	}

	deviceName := "wwan0"
	connectionName := "TestGSMConnection"
	ip4Method := "auto"

	// Act
	SetConnectionDetailsWithGSM(connection, gsmConfig, deviceName, connectionName)

	// Assert
	assert.Equal(t, connectionName, connection[common.ConnectionKey][common.IDKey], "Connection ID should match")
	assert.Equal(t, common.GSMSetting, connection[common.ConnectionKey][common.TypeKey], "Type should be GSM")
	assert.Equal(t, true, connection[common.ConnectionKey][common.AutoConnectKey], "AutoConnect should be true")
	assert.Equal(t, deviceName, connection[common.ConnectionKey][common.InterfaceNameKey], "Interface name should match")
	assert.Equal(t, ip4Method, connection[common.IPV4Key][common.MethodKey], "ip4 method type should be auto")

	// GSM specific settings
	assert.Equal(t, gsmConfig.Gsm.Apn, connection[common.GSMSetting][common.APNKey], "APN should be set")
	assert.Equal(t, gsmConfig.Gsm.Pin, connection[common.GSMSetting][common.PINKey], "PIN should be set")
	assert.Equal(t, gsmConfig.Gsm.Username, connection[common.GSMSetting][common.UsernameKey], "Username should be set")
	assert.Equal(t, gsmConfig.Gsm.Password, connection[common.GSMSetting][common.PasswordKey], "Password should be set")
}

func TestSetConnectionDetailsWithGSM_WithEmptyUsernameAndPassword(t *testing.T) {
	// Arrange
	connection := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.test",
			Pin:      "1234",
			Username: "",
			Password: "",
		},
	}

	deviceName := "wwan0"
	connectionName := "TestGSMConnection"
	ip4Method := "auto"

	// Act
	SetConnectionDetailsWithGSM(connection, gsmConfig, deviceName, connectionName)

	// Assert
	assert.Equal(t, connectionName, connection[common.ConnectionKey][common.IDKey], "Connection ID should match")
	assert.Equal(t, common.GSMSetting, connection[common.ConnectionKey][common.TypeKey], "Type should be GSM")
	assert.Equal(t, true, connection[common.ConnectionKey][common.AutoConnectKey], "AutoConnect should be true")
	assert.Equal(t, deviceName, connection[common.ConnectionKey][common.InterfaceNameKey], "Interface name should match")
	assert.Equal(t, ip4Method, connection[common.IPV4Key][common.MethodKey], "ip4 method type should be auto")

	// GSM specific settings
	assert.Equal(t, gsmConfig.Gsm.Apn, connection[common.GSMSetting][common.APNKey], "APN should be set")
	assert.Equal(t, gsmConfig.Gsm.Pin, connection[common.GSMSetting][common.PINKey], "PIN should be set")

	// Username and Password should not be set when empty
	assert.Nil(t, connection[common.GSMSetting][common.UsernameKey], "Username should not be set when empty")
	assert.Nil(t, connection[common.GSMSetting][common.PasswordKey], "Password should not be set when empty")
}

func TestSetConnectionDetailsWithGSM_WithEmptyUsername(t *testing.T) {
	// Arrange
	connection := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.test",
			Pin:      "1234",
			Username: "",
			Password: "testpass",
		},
	}

	deviceName := "wwan0"
	connectionName := "TestGSMConnection"
	ip4Method := "auto"
	// Act
	SetConnectionDetailsWithGSM(connection, gsmConfig, deviceName, connectionName)

	// Assert
	// Username and Password should not be set when username is empty
	assert.Nil(t, connection[common.GSMSetting][common.UsernameKey], "Username should not be set when empty")
	assert.Nil(t, connection[common.GSMSetting][common.PasswordKey], "Password should not be set when username is empty")
	assert.Equal(t, ip4Method, connection[common.IPV4Key][common.MethodKey], "ip4 method type should be auto")
}

func TestSetConnectionDetailsWithGSM_WithEmptyPassword(t *testing.T) {
	// Arrange
	connection := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "internet.test",
			Pin:      "1234",
			Username: "testuser",
			Password: "",
		},
	}

	deviceName := "wwan0"
	connectionName := "TestGSMConnection"
	ip4Method := "auto"

	// Act
	SetConnectionDetailsWithGSM(connection, gsmConfig, deviceName, connectionName)

	// Assert
	// Username and Password should not be set when password is empty
	assert.Nil(t, connection[common.GSMSetting][common.UsernameKey], "Username should not be set when password is empty")
	assert.Nil(t, connection[common.GSMSetting][common.PasswordKey], "Password should not be set when empty")
	assert.Equal(t, ip4Method, connection[common.IPV4Key][common.MethodKey], "ip4 method type should be auto")
}

func TestSetConnectionDetailsWithGSM_WithEmptyPin(t *testing.T) {
	// Arrange
	connection := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "mobile.data.com",
			Pin:      "",
			Username: "username",
			Password: "p@ssw0rd!",
		},
	}

	deviceName := "wwan1"
	connectionName := "empty pin connection"
	ip4Method := "auto"

	// Act
	SetConnectionDetailsWithGSM(connection, gsmConfig, deviceName, connectionName)

	// Assert
	assert.Equal(t, connectionName, connection[common.ConnectionKey][common.IDKey], "Connection ID with spaces should be handled")
	assert.Equal(t, deviceName, connection[common.ConnectionKey][common.InterfaceNameKey], "Device name should match")
	assert.Equal(t, ip4Method, connection[common.IPV4Key][common.MethodKey], "ip4 method type should be auto")

	// GSM specific settings with special characters
	assert.Equal(t, gsmConfig.Gsm.Apn, connection[common.GSMSetting][common.APNKey], "APN with dots should be set")
	assert.Nil(t, connection[common.GSMSetting][common.PINKey], "PIN should not be set")
	assert.Equal(t, gsmConfig.Gsm.Username, connection[common.GSMSetting][common.UsernameKey], "Username should be set")
	assert.Equal(t, gsmConfig.Gsm.Password, connection[common.GSMSetting][common.PasswordKey], "Password should be set")
}

func TestSetConnectionDetailsWithGSM_ModifiesExistingConnection(t *testing.T) {
	// Arrange - connection with existing data
	existingData := "existing-value"
	connection := nm.ConnectionSettings{
		common.ConnectionKey: dict{
			"existing-key": existingData,
		},
		common.GSMSetting: dict{
			"existing-gsm-key": "existing-gsm-value",
		},
		common.IPV4Key: make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn:      "new.apn.com",
			Pin:      "9876",
			Username: "newuser",
			Password: "newpass",
		},
	}

	deviceName := "wwan0"
	connectionName := "ModifiedConnection"
	ip4Method := "auto"

	// Act
	SetConnectionDetailsWithGSM(connection, gsmConfig, deviceName, connectionName)

	// Assert
	// Existing data should be preserved
	assert.Equal(t, existingData, connection[common.ConnectionKey]["existing-key"], "Existing connection data should be preserved")
	assert.Equal(t, "existing-gsm-value", connection[common.GSMSetting]["existing-gsm-key"], "Existing GSM data should be preserved")

	// New data should be set
	assert.Equal(t, connectionName, connection[common.ConnectionKey][common.IDKey], "New connection ID should be set")
	assert.Equal(t, gsmConfig.Gsm.Apn, connection[common.GSMSetting][common.APNKey], "New APN should be set")
	assert.Equal(t, gsmConfig.Gsm.Pin, connection[common.GSMSetting][common.PINKey], "New PIN should be set")
	assert.Equal(t, gsmConfig.Gsm.Username, connection[common.GSMSetting][common.UsernameKey], "New username should be set")
	assert.Equal(t, gsmConfig.Gsm.Password, connection[common.GSMSetting][common.PasswordKey], "New password should be set")
	assert.Equal(t, ip4Method, connection[common.IPV4Key][common.MethodKey], "ip4 method type should be auto")
}

func TestSetConnectionDetailsWithGSM_UUIDGeneration(t *testing.T) {
	// Arrange
	connection1 := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}
	connection2 := nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.GSMSetting:    make(dict),
		common.IPV4Key:       make(dict),
	}

	gsmConfig := &v1.ConnectionSettings_Gsm{
		Gsm: &v1.Interface_GsmConf{
			Apn: "internet",
			Pin: "1234",
		},
	}

	// Act
	SetConnectionDetailsWithGSM(connection1, gsmConfig, "wwan0", "Connection1")
	SetConnectionDetailsWithGSM(connection2, gsmConfig, "wwan1", "Connection2")

	// Assert
	uuid1 := connection1[common.ConnectionKey][common.UUIDKey].(string)
	uuid2 := connection2[common.ConnectionKey][common.UUIDKey].(string)

	assert.NotEqual(t, uuid1, uuid2, "UUIDs should be unique for different connections")
	assert.NotEmpty(t, uuid1, "UUID should not be empty")
	assert.NotEmpty(t, uuid2, "UUID should not be empty")

	// Validate UUID format (basic check)
	_, err1 := uuid.Parse(uuid1)
	_, err2 := uuid.Parse(uuid2)
	assert.NoError(t, err1, "First UUID should be valid")
	assert.NoError(t, err2, "Second UUID should be valid")
}
