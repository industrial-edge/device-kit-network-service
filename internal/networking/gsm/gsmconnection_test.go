/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/mocks"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	mocksnetworking "networkservice/internal/networking/mocks/networking"

	nm "github.com/Wifx/gonetworkmanager/v2"
	nmanager "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/godbus/dbus/v5"

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
	assert.Equal(t, err.Error(), "gsm configuration validation error: invalid PIN format")
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
	assert.Equal(t, err.Error(), "gsm device lookup failed: no GSM modem device found among network devices")
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

	// Rollback will attempt to restore the backed-up connection
	mockRestoredConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/3"))

	// doBackup: connManager.FindConnectionWithStatus returns active connection
	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice2).Return(mockConn, true, nil)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockRestoredConn, mockDevice2).Return(nil)
	mockConnectionManager.On("DeactivateConnection", mockDevice2).Return(nil)
	// No fileUtils set — so no config file backup, connectionMatchesConfigFile not called
	// backupActiveConnection will get settings for backup
	mockConn.On("GetSettings").Return(mockExistingConnSettings, nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("not found"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connManager:    mockConnectionManager,
		fileUtils:      mockFileUtils,
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

// CreateConnection — backup fails when connection lookup fails (Step 5)
func Test_CreateConnection_BackupFailsWhenConnectionLookupFails(t *testing.T) {
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

	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)
	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice2.On("GetPropertyInterface").Return("wwan0", nil)

	// doBackup: FindConnectionWithStatus returns error — CreateConnection must fail
	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice2).Return(nil, false, errors.New("connection lookup failed"))

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet"}`), nil)

	// No existing GSM connections
	mockSettings.On("ListConnections").Return([]nmanager.Connection{}, nil)

	mockNewConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockNewConn, nil)
	mockNewConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/new"))
	mockConnectionManager.On("ActivateConnectionWithRetry", mockNewConn, mockDevice2).Return(nil)

	patches.ApplyPrivateMethod(reflect.TypeOf(&GSMConnection{}), "writeGSMConfigToFile",
		func(_ *GSMConnection, _ *v1.ConnectionSettings_Gsm) error {
			return nil
		})

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		fileUtils:      mockFileUtils,
		connManager:    mockConnectionManager,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should fail when backup connection lookup fails")
	assert.Contains(t, err.Error(), "failed to backup existing GSM state")
	assert.Contains(t, err.Error(), "gsm connection backup failed")
	assert.Contains(t, err.Error(), "failed to find connection with status during backup")
}

// CreateConnection — Delete fails, rollback triggered and succeeds (Step 6)
func Test_CreateConnection_DeleteFailsAndRollbackSucceeds(t *testing.T) {
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

	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)
	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice2.On("GetPropertyInterface").Return("wwan0", nil)

	mockExistingConn := &mockgnm.MockConnection{}

	// doBackup: connManager.FindConnectionWithStatus returns active connection
	// Rollback will re-add the backed-up connection
	mockRestoredConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/restored"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice2).Return(mockExistingConn, true, nil)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockRestoredConn, mockDevice2).Return(nil)
	mockConnectionManager.On("DeactivateConnection", mockDevice2).Return(nil)

	mockExistingConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))
	// GetSettings returns both connection and gsm sections (used by backup validation and deleteExisting)
	mockExistingConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "old-gsm",
			"uuid": "old-uuid",
			"type": "gsm",
		},
		"gsm": map[string]interface{}{
			"apn": "my-internet",
		},
	}, nil)
	mockExistingConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	mockSettings.On("ListConnections").Return([]nmanager.Connection{mockExistingConn}, nil)
	mockExistingConn.On("Delete").Return(errors.New("delete failed"))

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet"}`), nil)
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600)).Return(nil)
	mockFileUtils.On("RemoveAll", common.GSMConfigPath).Return(nil)

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		fileUtils:      mockFileUtils,
		connManager:    mockConnectionManager,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when delete existing GSM connection")
	assert.Equal(t, err.Error(), "failed to delete existing GSM connections: delete failed")
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
	mockSettings.On("AddConnection", mock.Anything).Return(nil, errors.New("Adding Gsm connection failed"))
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	// doBackup: connManager.FindConnectionWithStatus returns no active connection
	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice2).Return(nil, false, errors.New("no active connection"))
	mockConnectionManager.On("DeactivateConnection", mockDevice2).Return(nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("not found"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connManager:    mockConnectionManager,
		fileUtils:      mockFileUtils,
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

	mockSettings.On("AddConnection", mock.Anything).Return(mockConnection, nil)
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	// doBackup: connManager.FindConnectionWithStatus returns no active connection
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice1).Return(nil, false, errors.New("no active connection"))
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice1).Return(errors.New("could not activate connection"))
	mockConnectionManager.On("DeactivateConnection", mockDevice1).Return(nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("no file"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connectionSettingFromNM: nmanager.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
		connManager: mockConnectionManager,
		fileUtils:   mockFileUtils,
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

	mockSettings.On("AddConnection", mock.Anything).Return(mockConnection, nil)
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	// doBackup: connManager.FindConnectionWithStatus returns no active connection
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice1).Return(nil, false, errors.New("no active connection"))
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice1).Return(nil)
	mockConnectionManager.On("DeactivateConnection", mockDevice1).Return(nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("no file"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connectionSettingFromNM: nmanager.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
		connManager: mockConnectionManager,
		fileUtils:   mockFileUtils,
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

	mockSettings.On("AddConnection", mock.Anything).Return(mockConnection, nil)
	mockConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	// doBackup: connManager.FindConnectionWithStatus returns no active connection
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice1).Return(nil, false, errors.New("no active connection"))
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice1).Return(nil)
	mockConnectionManager.On("DeactivateConnection", mockDevice1).Return(nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("no file"))

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		fileUtils:      mockFileUtils,
		connManager:    mockConnectionManager,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should return an error when writeGSMConfigToFile fails")
	assert.Contains(t, err.Error(), "failed to persist GSM configuration")
	assert.Contains(t, err.Error(), "file write error")
}

// CreateConnection — writeGSMConfigToFile fails with active backup, triggers rollback
func Test_CreateConnection_writeGSMConfigToFile_Error_TriggersRollback(t *testing.T) {
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

	mockNewConnection := &mockgnm.MockConnection{}
	mockNewConnection.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/new"))
	mockExistingConn := &mockgnm.MockConnection{}
	connections := []nmanager.Connection{mockExistingConn}

	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)

	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)

	mockDevice1.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice1.On("GetPropertyInterface").Return("interfacetest", nil)

	mockSettings.On("ListConnections").Return(connections, nil)
	mockExistingConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":   "testgsm",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "gsm",
		},
		"gsm": map[string]interface{}{
			"apn": "my-internet",
		},
	}, nil)
	mockExistingConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/2"))
	// Delete fails — triggers rollback
	mockExistingConn.On("Delete").Return(errors.New("delete failed"))
	mockExistingConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	// Rollback will re-add the backed-up connection
	mockRestoredConn := &mockgnm.MockConnection{}
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/3"))
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice1).Return(mockExistingConn, true, nil)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockRestoredConn, mockDevice1).Return(nil)
	mockConnectionManager.On("DeactivateConnection", mockDevice1).Return(nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet"}`), nil)
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600)).Return(nil)

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		fileUtils:      mockFileUtils,
		connManager:    mockConnectionManager,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err, "CreateConnection should  return an error when delete existing GSM connection")
	assert.Equal(t, err.Error(), "failed to delete existing GSM connections: delete failed")
}

// --- CreateConnection: doBackup fails — returns error without triggering rollback ---
func Test_CreateConnection_DoBackupFails(t *testing.T) {
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

	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)
	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice2.On("GetPropertyInterface").Return("wwan0", nil)

	// doBackup fails because config file exists but cannot be read
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte{}, errors.New("permission denied"))

	mockConnectionManager := new(mocksnetworking.MockConnectionManager)

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connManager:    mockConnectionManager,
		fileUtils:      mockFileUtils,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to backup existing GSM state")
}

// --- CreateConnection: post-backup step fails and doRollback also fails ---
func Test_CreateConnection_RollbackFails(t *testing.T) {
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

	mockNetworkManager.On("GetDevices").Return([]nmanager.Device{mockDevice1, mockDevice2}, nil)
	mockDevice1.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nmanager.NmDeviceTypeModem, nil)
	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Devices/2"))
	mockDevice2.On("GetPropertyInterface").Return("wwan0", nil)

	// doBackup succeeds with config file + active connection
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet"}`), nil)

	mockExistingConn := &mockgnm.MockConnection{}
	mockConnectionManager := new(mocksnetworking.MockConnectionManager)
	mockConnectionManager.On("FindConnectionWithStatus", mockDevice2).Return(mockExistingConn, true, nil)
	mockConnectionManager.On("DeactivateConnection", mockDevice2).Return(nil)

	mockExistingConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))
	mockExistingConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{"id": "old-gsm", "type": "gsm"},
		"gsm":        map[string]interface{}{"apn": "my-internet"},
	}, nil)
	mockExistingConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	// deleteExistingGSMConnections finds and fails to delete
	mockSettings.On("ListConnections").Return([]nmanager.Connection{mockExistingConn}, nil)
	mockExistingConn.On("Delete").Return(errors.New("delete failed"))

	// Rollback: restoreConnection fails (AddConnection returns error)
	mockSettings.On("AddConnection", mock.Anything).Return(nil, errors.New("rollback add failed"))

	// Rollback: restoreConfigFile �� needed since hasFileConfig is true
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600)).Return(nil)

	gsmConnection := &GSMConnection{
		networkManager: mockNetworkManager,
		nmSettings:     mockSettings,
		connManager:    mockConnectionManager,
		fileUtils:      mockFileUtils,
		connectionSettingFromNM: nm.ConnectionSettings{
			common.ConnectionKey: make(map[string]interface{}),
			common.IPV4Key:       make(map[string]interface{}),
			common.GSMSetting:    make(map[string]interface{}),
		},
	}

	err := gsmConnection.CreateConnection(newSetting)
	// The original error is returned even though rollback also failed
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to delete existing GSM connections: delete failed")
}

// --- backupConfigFile: no GSM section in backup ---
func Test_overrideBackupWithFileConfig_NoGSMSection(t *testing.T) {
	gsmConn := &GSMConnection{}
	backup := &gsmSnapshot{
		fileConfig:         map[string]string{"Apn": "new-apn", "Pin": "5678"},
		connectionSettings: nmanager.ConnectionSettings{},
	}

	gsmConn.overrideBackupWithFileConfig(backup)

	gsmSection := backup.connectionSettings[common.GSMSetting]
	assert.Equal(t, "new-apn", gsmSection[common.APNKey])
	assert.Equal(t, "5678", gsmSection[common.PINKey])
}

// --- overrideBackupWithFileConfig: optional fields absent in file get removed ---
func Test_overrideBackupWithFileConfig_RemovesAbsentOptionalFields(t *testing.T) {
	gsmConn := &GSMConnection{}
	backup := &gsmSnapshot{
		fileConfig: map[string]string{"Apn": "new-apn"},
		connectionSettings: nmanager.ConnectionSettings{
			"gsm": map[string]interface{}{
				"apn":      "old-apn",
				"username": "old-user",
				"password": "old-pass",
				"pin":      "0000",
			},
		},
	}

	gsmConn.overrideBackupWithFileConfig(backup)

	gsmSection := backup.connectionSettings["gsm"]
	assert.Equal(t, "new-apn", gsmSection[common.APNKey])
	_, hasUser := gsmSection[common.UsernameKey]
	_, hasPass := gsmSection[common.PasswordKey]
	_, hasPin := gsmSection[common.PINKey]
	assert.False(t, hasUser, "Username should be removed since not in file")
	assert.False(t, hasPass, "Password should be removed since not in file")
	assert.False(t, hasPin, "PIN should be removed since not in file")
}

// --- connectionMatchesConfigFile: match ---
func Test_connectionMatchesConfigFile_Match(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet", "username": "user1"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"pin": "1234", "password": "pass1"},
	}, nil)

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn,
		map[string]string{"Apn": "my-internet", "Username": "user1", "Pin": "1234", "Password": "pass1"})

	assert.True(t, result)
}

// --- connectionMatchesConfigFile: APN mismatch ---
func Test_connectionMatchesConfigFile_APNMismatch(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "wrong-apn"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn, map[string]string{"Apn": "my-internet"})

	assert.False(t, result)
}

// --- connectionMatchesConfigFile: GetSettings fails ---
func Test_connectionMatchesConfigFile_GetSettingsFails(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{}, errors.New("D-Bus error"))

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn, map[string]string{"Apn": "my-internet"})

	assert.False(t, result)
}

// --- connectionMatchesConfigFile: GetSecrets fails ---
func Test_connectionMatchesConfigFile_GetSecretsFails(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{}, errors.New("secrets error"))

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn, map[string]string{"Apn": "my-internet"})

	assert.False(t, result)
}

// --- connectionMatchesConfigFile: optional fields match when absent in file ---
func Test_connectionMatchesConfigFile_OptionalFieldsAbsentInFile(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet", "username": "some-user"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"pin": "9999", "password": "secret"},
	}, nil)

	gsmConn := &GSMConnection{}
	// File only has APN — optional fields should be considered matching
	result := gsmConn.connectionMatchesConfigFile(mockConn, map[string]string{"Apn": "my-internet"})

	assert.True(t, result)
}

// --- connectionMatchesConfigFile: optional field mismatch (Username) ---
func Test_connectionMatchesConfigFile_OptionalFieldMismatch_Username(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet", "username": "wrong-user"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn,
		map[string]string{"Apn": "my-internet", "Username": "correct-user"})

	assert.False(t, result)
}

// --- connectionMatchesConfigFile: optional field mismatch (Pin) ---
func Test_connectionMatchesConfigFile_OptionalFieldMismatch_Pin(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"pin": "0000"},
	}, nil)

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn,
		map[string]string{"Apn": "my-internet", "Pin": "1234"})

	assert.False(t, result)
}

// --- connectionMatchesConfigFile: optional field mismatch (Password) ---
func Test_connectionMatchesConfigFile_OptionalFieldMismatch_Password(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"password": "wrong-pass"},
	}, nil)

	gsmConn := &GSMConnection{}
	result := gsmConn.connectionMatchesConfigFile(mockConn,
		map[string]string{"Apn": "my-internet", "Password": "correct-pass"})

	assert.False(t, result)
}

// --- connectionMatchesConfigFile: optional field empty in file skips comparison ---
func Test_connectionMatchesConfigFile_OptionalFieldEmptyInFileSkipped(t *testing.T) {
	mockConn := &mockgnm.MockConnection{}
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet", "username": "some-user"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"pin": "9999", "password": "secret"},
	}, nil)

	gsmConn := &GSMConnection{}
	// File has optional fields but they are empty — should skip comparison and match
	result := gsmConn.connectionMatchesConfigFile(mockConn,
		map[string]string{"Apn": "my-internet", "Username": "", "Pin": "", "Password": ""})

	assert.True(t, result)
}

// --- restoreConfigFile: MkdirAll fails ---
func Test_restoreConfigFile_MkdirFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(errors.New("mkdir failed"))

	gsmConn := &GSMConnection{
		fileUtils: mockFileUtils,
		gsmBackup: &gsmSnapshot{
			fileConfig: map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.restoreConfigFile()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create GSM config directory")
}

// --- restoreConfigFile: WriteFile fails ---
func Test_restoreConfigFile_WriteFileFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, []byte(`{"Apn":"test"}`), os.FileMode(0600)).Return(errors.New("write failed"))

	gsmConn := &GSMConnection{
		fileUtils: mockFileUtils,
		gsmBackup: &gsmSnapshot{
			fileConfig: map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.restoreConfigFile()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write GSM config file")
}

// --- restoreConfigFile: json.Marshal fails ---
func Test_restoreConfigFile_MarshalFails(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(json.Marshal, func(_ interface{}) ([]byte, error) {
		return nil, errors.New("marshal error")
	})

	gsmConn := &GSMConnection{
		gsmBackup: &gsmSnapshot{
			fileConfig: map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.restoreConfigFile()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal GSM config for restore")
}

// --- restoreConfigFile: success ---
func Test_restoreConfigFile_Success(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, []byte(`{"Apn":"test"}`), os.FileMode(0600)).Return(nil)

	gsmConn := &GSMConnection{
		fileUtils: mockFileUtils,
		gsmBackup: &gsmSnapshot{
			fileConfig: map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.restoreConfigFile()

	assert.NoError(t, err)
}

// --- backupConfigFile: ReadFile fails ---
func Test_backupConfigFile_ReadFileFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte{}, errors.New("permission denied"))

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}
	backup := &gsmSnapshot{}

	err := gsmConn.backupConfigFile(backup)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read GSM config file for backup")
	assert.False(t, backup.hasFileConfig)
}

// --- backupConfigFile: file is empty — logs warning, skips backup ---
func Test_backupConfigFile_EmptyFile(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte{}, nil)

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}
	backup := &gsmSnapshot{}

	err := gsmConn.backupConfigFile(backup)

	assert.NoError(t, err, "Empty file should not return error, just skip backup")
	assert.False(t, backup.hasFileConfig)
}

// --- backupConfigFile: invalid JSON — logs warning, skips backup ---
func Test_backupConfigFile_InvalidJSON(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{not valid json`), nil)

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}
	backup := &gsmSnapshot{}

	err := gsmConn.backupConfigFile(backup)

	assert.NoError(t, err, "Invalid JSON should not return error, just skip backup")
	assert.False(t, backup.hasFileConfig)
}

// --- backupConfigFile: validation fails (e.g. missing APN) — logs warning, skips backup ---
func Test_backupConfigFile_ValidationFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	// Valid JSON but missing required "Apn" key
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Pin":"1234"}`), nil)

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}
	backup := &gsmSnapshot{}

	err := gsmConn.backupConfigFile(backup)

	assert.NoError(t, err, "Validation failure should not return error, just skip backup")
	assert.False(t, backup.hasFileConfig)
}

// --- backupConfigFile: success ---
func Test_backupConfigFile_Success(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet","Pin":"1234"}`), nil)

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}
	backup := &gsmSnapshot{}

	err := gsmConn.backupConfigFile(backup)

	assert.NoError(t, err)
	assert.True(t, backup.hasFileConfig)
	assert.Equal(t, "my-internet", backup.fileConfig["Apn"])
	assert.Equal(t, "1234", backup.fileConfig["Pin"])
}

// --- backupConfigFile: file does not exist — skip backup ---
func Test_backupConfigFile_FileNotFound(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("file not found"))

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}
	backup := &gsmSnapshot{}

	err := gsmConn.backupConfigFile(backup)

	assert.NoError(t, err)
	assert.False(t, backup.hasFileConfig)
	assert.Nil(t, backup.fileConfig)
}

// --- backupConnection: no file config — skips connection backup ---
func Test_backupConnection_NoFileConfig(t *testing.T) {
	gsmConn := &GSMConnection{}
	backup := &gsmSnapshot{hasFileConfig: false}

	err := gsmConn.backupConnection(backup)

	assert.NoError(t, err)
	assert.False(t, backup.hasConnection)
}

// --- backupConnection: FindConnectionWithStatus returns error ---
func Test_backupConnection_FindConnectionWithStatusError(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(nil, false, errors.New("dbus timeout"))

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "test"}}

	err := gsmConn.backupConnection(backup)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find connection with status during backup")
}

// --- backupConnection: active connection found, storeConnectionBackup fails ---
func Test_backupConnection_ActiveConnectionStoreBackupFails(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockConnection{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)

	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{}, errors.New("dbus error"))

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "test"}}

	err := gsmConn.backupConnection(backup)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store active connection backup")
}

// --- backupConnection: active connection matches config file ---
func Test_backupConnection_ActiveConnectionMatchesFile(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockConnection{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)

	connSettings := nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet"},
	}
	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))
	mockConn.On("GetSettings").Return(connSettings, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "my-internet"}}

	err := gsmConn.backupConnection(backup)

	assert.NoError(t, err)
	assert.True(t, backup.hasConnection)
	assert.True(t, backup.wasActive)
}

// --- backupConnection: active connection does NOT match config file — override applied ---
func Test_backupConnection_ActiveConnectionMismatchOverride(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockConnection{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)

	connSettings := nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "old-apn"},
	}
	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))
	mockConn.On("GetSettings").Return(connSettings, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "new-apn"}}

	err := gsmConn.backupConnection(backup)

	assert.NoError(t, err)
	assert.True(t, backup.hasConnection)
	assert.True(t, backup.wasActive)
	// Verify override was applied — GSM section should have the file's APN
	assert.Equal(t, "new-apn", backup.connectionSettings["gsm"][common.APNKey])
}

// --- backupConnection: no active connection, inactive match found ---
func Test_backupConnection_InactiveConnectionMatchesFile(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(nil, false, nil)

	mockSettings := &mockgnm.MockSettings{}
	mockConn := &mockgnm.MockConnection{}
	mockSettings.On("ListConnections").Return([]nmanager.Connection{mockConn}, nil)
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"connection": map[string]interface{}{"type": "gsm"},
		"gsm":        map[string]interface{}{"apn": "my-internet"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)
	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice, nmSettings: mockSettings}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "my-internet"}}

	err := gsmConn.backupConnection(backup)

	assert.NoError(t, err)
	assert.True(t, backup.hasConnection)
	assert.False(t, backup.wasActive)
}

// --- backupConnection: no active connection, no inactive match ---
func Test_backupConnection_NoMatchingConnection(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(nil, false, nil)

	mockSettings := &mockgnm.MockSettings{}
	mockSettings.On("ListConnections").Return([]nmanager.Connection{}, nil)

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice, nmSettings: mockSettings}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "my-internet"}}

	err := gsmConn.backupConnection(backup)

	assert.NoError(t, err)
	assert.False(t, backup.hasConnection)
}

// --- backupConnection: no active connection, inactive match found but storeConnectionBackup fails ---
func Test_backupConnection_InactiveConnectionStoreBackupFails(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(nil, false, nil)

	mockSettings := &mockgnm.MockSettings{}
	mockConn := &mockgnm.MockConnection{}
	mockSettings.On("ListConnections").Return([]nmanager.Connection{mockConn}, nil)

	// First GetSettings call (from findConnectionBySettingValue) succeeds
	// Second GetSettings call (from connectionMatchesConfigFile via extractGSMConnFields) succeeds
	// Third GetSettings call (from storeConnectionBackup) fails
	connSettings := nmanager.ConnectionSettings{
		"connection": map[string]interface{}{"type": "gsm"},
		"gsm":        map[string]interface{}{"apn": "my-internet"},
	}
	mockConn.On("GetSettings").Return(connSettings, nil).Times(2)
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{}, errors.New("dbus error")).Once()
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)
	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice, nmSettings: mockSettings}
	backup := &gsmSnapshot{hasFileConfig: true, fileConfig: map[string]string{"Apn": "my-internet"}}

	err := gsmConn.backupConnection(backup)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store inactive connection backup")
}

// --- doBackup: success with no prior config file (fresh device) ---
func Test_doBackup_Success_NoConfigFile(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("not found"))

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}

	err := gsmConn.doBackup()

	assert.NoError(t, err)
	assert.NotNil(t, gsmConn.gsmBackup)
	assert.False(t, gsmConn.gsmBackup.hasFileConfig)
	assert.False(t, gsmConn.gsmBackup.hasConnection)
}

// --- doBackup: success with config file and active connection ---
func Test_doBackup_Success_WithFileAndActiveConnection(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet"}`), nil)

	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockConnection{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)

	mockConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/1"))
	mockConn.On("GetSettings").Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{"apn": "my-internet"},
	}, nil)
	mockConn.On("GetSecrets", common.GSMSetting).Return(nmanager.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}, nil)

	gsmConn := &GSMConnection{
		fileUtils:   mockFileUtils,
		connManager: mockConnMgr,
		gsmDevice:   mockDevice,
	}

	err := gsmConn.doBackup()

	assert.NoError(t, err)
	assert.NotNil(t, gsmConn.gsmBackup)
	assert.True(t, gsmConn.gsmBackup.hasFileConfig)
	assert.True(t, gsmConn.gsmBackup.hasConnection)
	assert.True(t, gsmConn.gsmBackup.wasActive)
}

// --- doBackup: backupConfigFile fails ---
func Test_doBackup_BackupConfigFileFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte{}, errors.New("permission denied"))

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}

	err := gsmConn.doBackup()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gsm Config file backup failed")
	assert.Nil(t, gsmConn.gsmBackup)
}

// --- doBackup: backupConnection fails ---
func Test_doBackup_BackupConnectionFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte(`{"Apn":"my-internet"}`), nil)

	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("FindConnectionWithStatus", mockDevice).Return(nil, false, errors.New("dbus timeout"))

	gsmConn := &GSMConnection{
		fileUtils:   mockFileUtils,
		connManager: mockConnMgr,
		gsmDevice:   mockDevice,
	}

	err := gsmConn.doBackup()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gsm connection backup failed")
	assert.Nil(t, gsmConn.gsmBackup)
}

// --- activateRestoredConnection: success ---
func Test_activateRestoredConnection_Success(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockConnection{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("ActivateConnectionWithRetry", mockConn, mockDevice).Return(nil)

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice}

	err := gsmConn.activateRestoredConnection(mockConn)

	assert.NoError(t, err)
	mockConnMgr.AssertCalled(t, "ActivateConnectionWithRetry", mockConn, mockDevice)
}

// --- activateRestoredConnection: ActivateConnectionWithRetry fails ---
func Test_activateRestoredConnection_Fails(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockConnection{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("ActivateConnectionWithRetry", mockConn, mockDevice).Return(errors.New("activation timeout"))

	gsmConn := &GSMConnection{connManager: mockConnMgr, gsmDevice: mockDevice}

	err := gsmConn.activateRestoredConnection(mockConn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to activate restored GSM connection")
	assert.Contains(t, err.Error(), "activation timeout")
}

// --- restoreConnection: AddConnection fails ---
func Test_restoreConnection_AddConnectionFails(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockSettings.On("AddConnection", mock.Anything).Return(nil, errors.New("add failed"))

	gsmConn := &GSMConnection{
		nmSettings: mockSettings,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			wasActive: false,
		},
	}

	err := gsmConn.restoreConnection()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not restore GSM connection")
}

// --- restoreConnection: wasActive=false, skips activation ---
func Test_restoreConnection_InactiveSkipsActivation(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockRestoredConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/restored"))

	mockConnMgr := new(mocksnetworking.MockConnectionManager)

	gsmConn := &GSMConnection{
		nmSettings:  mockSettings,
		connManager: mockConnMgr,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			wasActive: false,
		},
	}

	err := gsmConn.restoreConnection()

	assert.NoError(t, err)
	mockConnMgr.AssertNotCalled(t, "ActivateConnectionWithRetry", mock.Anything, mock.Anything)
}

// --- restoreConnection: wasActive=true, activation succeeds ---
func Test_restoreConnection_ActiveAndActivationSucceeds(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockRestoredConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/restored"))

	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("ActivateConnectionWithRetry", mockRestoredConn, mockDevice).Return(nil)

	gsmConn := &GSMConnection{
		nmSettings:  mockSettings,
		connManager: mockConnMgr,
		gsmDevice:   mockDevice,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			wasActive: true,
		},
	}

	err := gsmConn.restoreConnection()

	assert.NoError(t, err)
	mockConnMgr.AssertCalled(t, "ActivateConnectionWithRetry", mockRestoredConn, mockDevice)
}

// --- restoreConnection: wasActive=true, activation fails ---
func Test_restoreConnection_ActiveAndActivationFails(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockRestoredConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/restored"))

	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("ActivateConnectionWithRetry", mockRestoredConn, mockDevice).Return(errors.New("activation failed"))

	gsmConn := &GSMConnection{
		nmSettings:  mockSettings,
		connManager: mockConnMgr,
		gsmDevice:   mockDevice,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			wasActive: true,
		},
	}

	err := gsmConn.restoreConnection()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not activate GSM connection")
}

// --- doRollback: nothing to restore ---
func Test_doRollback_NothingToRestore(t *testing.T) {
	gsmConn := &GSMConnection{
		gsmBackup: &gsmSnapshot{},
	}

	err := gsmConn.doRollback()

	assert.NoError(t, err)
}

// --- doRollback: restoreConnection fails ---
func Test_doRollback_RestoreConnectionFails(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockSettings.On("AddConnection", mock.Anything).Return(nil, errors.New("add failed"))

	gsmConn := &GSMConnection{
		nmSettings: mockSettings,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			wasActive: false,
		},
	}

	err := gsmConn.doRollback()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
	assert.Contains(t, err.Error(), "restore connection")
}

// --- doRollback: restoreConnection fails, but restoreConfigFile is still attempted ---
func Test_doRollback_RestoreConnectionFails_StillRestoresConfigFile(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockSettings.On("AddConnection", mock.Anything).Return(nil, errors.New("add failed"))

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600)).Return(nil)

	gsmConn := &GSMConnection{
		nmSettings: mockSettings,
		fileUtils:  mockFileUtils,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			hasFileConfig: true,
			fileConfig:    map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.doRollback()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
	assert.Contains(t, err.Error(), "restore connection")
	mockFileUtils.AssertCalled(t, "MkdirAll", common.GSMConfigParentPath, os.FileMode(0755))
	mockFileUtils.AssertCalled(t, "WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600))
}

// --- doRollback: restoreConfigFile fails ---
func Test_doRollback_RestoreConfigFileFails(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(errors.New("mkdir failed"))

	gsmConn := &GSMConnection{
		fileUtils: mockFileUtils,
		gsmBackup: &gsmSnapshot{
			hasFileConfig: true,
			fileConfig:    map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.doRollback()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
	assert.Contains(t, err.Error(), "restore config file")
}

// --- doRollback: both connection and file restored successfully ---
func Test_doRollback_BothConnectionAndFileRestored(t *testing.T) {
	mockSettings := &mockgnm.MockSettings{}
	mockRestoredConn := &mockgnm.MockConnection{}
	mockSettings.On("AddConnection", mock.Anything).Return(mockRestoredConn, nil)
	mockRestoredConn.On("GetPath").Return(dbus.ObjectPath("/org/freedesktop/NetworkManager/Connection/restored"))

	mockDevice := &mockgnm.MockDevice{}
	mockConnMgr := new(mocksnetworking.MockConnectionManager)
	mockConnMgr.On("ActivateConnectionWithRetry", mockRestoredConn, mockDevice).Return(nil)

	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600)).Return(nil)

	gsmConn := &GSMConnection{
		nmSettings:  mockSettings,
		connManager: mockConnMgr,
		gsmDevice:   mockDevice,
		fileUtils:   mockFileUtils,
		gsmBackup: &gsmSnapshot{
			hasConnection: true,
			connectionSettings: nmanager.ConnectionSettings{
				"connection": map[string]interface{}{
					"id":   "old-gsm",
					"uuid": "old-uuid",
				},
			},
			wasActive:     true,
			hasFileConfig: true,
			fileConfig:    map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.doRollback()

	assert.NoError(t, err)
}

// --- doRollback: only file config, no connection ---
func Test_doRollback_OnlyFileConfigRestored(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(nil)
	mockFileUtils.On("WriteFile", common.GSMConfigPath, mock.Anything, os.FileMode(0600)).Return(nil)

	gsmConn := &GSMConnection{
		fileUtils: mockFileUtils,
		gsmBackup: &gsmSnapshot{
			hasFileConfig: true,
			fileConfig:    map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.doRollback()

	assert.NoError(t, err)
}

func Test_doBackup_Success(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(errors.New("not found"))

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}

	err := gsmConn.doBackup()

	assert.NoError(t, err)
	assert.NotNil(t, gsmConn.gsmBackup)
}

func Test_doBackup_Error(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("Stat", common.GSMConfigPath).Return(nil)
	mockFileUtils.On("ReadFile", common.GSMConfigPath).Return([]byte{}, errors.New("read error"))

	gsmConn := &GSMConnection{fileUtils: mockFileUtils}

	err := gsmConn.doBackup()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gsm Config file backup failed")
}

func Test_doRollback_Success(t *testing.T) {
	gsmConn := &GSMConnection{
		gsmBackup: &gsmSnapshot{},
	}

	err := gsmConn.doRollback()

	assert.NoError(t, err)
}

func Test_doRollback_Error(t *testing.T) {
	mockFileUtils := &mocks.MockFileSystem{}
	mockFileUtils.On("MkdirAll", common.GSMConfigParentPath, os.FileMode(0755)).Return(errors.New("mkdir failed"))

	gsmConn := &GSMConnection{
		fileUtils: mockFileUtils,
		gsmBackup: &gsmSnapshot{
			hasFileConfig: true,
			fileConfig:    map[string]string{"Apn": "test"},
		},
	}

	err := gsmConn.doRollback()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
}
