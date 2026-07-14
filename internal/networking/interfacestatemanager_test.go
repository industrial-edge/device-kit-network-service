/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/devicemanager"
	"networkservice/internal/networking/interfaces"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	"networkservice/internal/networking/mocks/networking"
	"reflect"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestInterfaceStateHandler_Backup_Success(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnSettingsOne := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn1",
			"interface-name": ifaceOne.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5A",
		},
	}
	mockIsActiveOne := true

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnSettingsTwo := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn2",
			"interface-name": ifaceTwo.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 101,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5B",
		},
	}
	mockIsActiveTwo := false

	mockConnectionThree := new(gonetworkmanager.MockConnection)
	mockConnSettingsThree := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn3",
			"interface-name": ifaceThree.InterfaceName,
			"type":           "gsm",
		},
		"ipv4": map[string]any{
			"route-metric": 100,
			"method":       "auto",
		},
		"gsm": map[string]any{
			"apn":      "internet",
			"pin":      "1234",
			"username": "user",
			"password": "pass",
		},
	}
	mockIsActiveThree := true

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		ifaceOne.InterfaceName:   mockDeviceOne,
		ifaceTwo.InterfaceName:   mockDeviceTwo,
		ifaceThree.InterfaceName: mockDeviceThree,
	}, nil).Once()

	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceOne).Return(mockConnectionOne, mockIsActiveOne, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceTwo).Return(mockConnectionTwo, mockIsActiveTwo, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceThree).Return(mockConnectionThree, mockIsActiveThree, nil).Once()

	mockConnectionManager.On("GetConnectionSettings", mockConnectionOne).Return(mockConnSettingsOne, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionTwo).Return(mockConnSettingsTwo, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionThree).Return(mockConnSettingsThree, nil).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      ifaceOne.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceOne.GatewayInterface,
			},
			{
				name:      ifaceTwo.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceTwo.GatewayInterface,
			},
			{
				name:      ifaceThree.InterfaceName,
				ifaceType: common.InterfaceTypeGSM,
				isGateway: ifaceThree.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.NoError(t, err, "Backup should not return an error")
	assert.Len(t, h.backups, 3, "Expected to have 3 backup")
	mockDeviceManager.AssertExpectations(t)
	mockConnectionManager.AssertExpectations(t)

	expectedBackups := []*backup{
		{
			ifaceName:    ifaceOne.InterfaceName,
			ifaceType:    common.InterfaceTypeEthernet,
			wasGateway:   true,
			wasActive:    true,
			device:       mockDeviceOne,
			conn:         mockConnectionOne,
			connSettings: mockConnSettingsOne,
		},
		{
			ifaceName:    ifaceTwo.InterfaceName,
			ifaceType:    common.InterfaceTypeEthernet,
			wasGateway:   false,
			wasActive:    false,
			device:       mockDeviceTwo,
			conn:         mockConnectionTwo,
			connSettings: mockConnSettingsTwo,
		},
		{
			ifaceName:    ifaceThree.InterfaceName,
			ifaceType:    common.InterfaceTypeGSM,
			wasGateway:   false,
			wasActive:    true,
			device:       mockDeviceThree,
			conn:         mockConnectionThree,
			connSettings: mockConnSettingsThree,
		},
	}
	backupComparer := cmp.Comparer(func(a, b *backup) bool {
		return a.ifaceName == b.ifaceName &&
			a.ifaceType == b.ifaceType &&
			a.wasGateway == b.wasGateway &&
			a.wasActive == b.wasActive &&
			reflect.DeepEqual(a.device, b.device) &&
			reflect.DeepEqual(a.conn, b.conn) &&
			compareConnSettingsBackups(a.connSettings, b.connSettings)
	})
	assert.True(t, cmp.Equal(expectedBackups, h.backups, backupComparer),
		"Backups do not match expected values")
}

func compareConnSettingsBackups(expected, actual nm.ConnectionSettings) bool {
	isConnSettingsEqual := expected[common.ConnectionKey]["id"] == actual[common.ConnectionKey]["id"] &&
		expected[common.ConnectionKey]["interface-name"] == actual[common.ConnectionKey]["interface-name"] &&
		expected[common.ConnectionKey]["type"] == actual[common.ConnectionKey]["type"]

	isIPv4SettingsEqual := expected["ipv4"]["route-metric"] == actual["ipv4"]["route-metric"] &&
		expected["ipv4"]["method"] == actual["ipv4"]["method"]

	expectedEthernetSettings, hasExpectedEthernet := expected["802-3-ethernet"]
	actualEthernetSettings, hasActualEthernet := actual["802-3-ethernet"]

	if hasExpectedEthernet && hasActualEthernet {
		isEthernetSettingsEqual := expectedEthernetSettings["mac-address"] == actualEthernetSettings["mac-address"]

		return isConnSettingsEqual && isIPv4SettingsEqual && isEthernetSettingsEqual
	}

	expectedGSMSettings, hasExpectedGSM := expected["gsm"]
	actualGSMSettings, hasActualGSM := actual["gsm"]

	if hasExpectedGSM && hasActualGSM {
		isGSMSettingsEqual := expectedGSMSettings["apn"] == actualGSMSettings["apn"] &&
			expectedGSMSettings["username"] == actualGSMSettings["username"]
		return isConnSettingsEqual && isIPv4SettingsEqual && isGSMSettingsEqual
	}

	return false
}

func TestInterfaceStateHandler_Backup_ConnSettingsNotFound_Success(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnSettingsOne := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn1",
			"interface-name": ifaceOne.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5A",
		},
	}
	mockIsActiveOne := true

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockIsActiveTwo := false

	mockConnectionThree := new(gonetworkmanager.MockConnection)
	mockConnSettingsThree := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn3",
			"interface-name": ifaceThree.InterfaceName,
			"type":           "gsm",
		},
		"ipv4": map[string]any{
			"route-metric": 100,
			"method":       "auto",
		},
		"gsm": map[string]any{
			"apn":      "internet",
			"pin":      "1234",
			"username": "user",
			"password": "pass",
		},
	}
	mockIsActiveThree := true

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		ifaceOne.InterfaceName:   mockDeviceOne,
		ifaceTwo.InterfaceName:   mockDeviceTwo,
		ifaceThree.InterfaceName: mockDeviceThree,
	}, nil).Once()

	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceOne).Return(mockConnectionOne, mockIsActiveOne, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceTwo).Return(mockConnectionTwo, mockIsActiveTwo, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceThree).Return(mockConnectionThree, mockIsActiveThree, nil).Once()

	mockConnectionManager.On("GetConnectionSettings", mockConnectionOne).Return(mockConnSettingsOne, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionTwo).Return(nil, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionThree).Return(mockConnSettingsThree, nil).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      ifaceOne.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceOne.GatewayInterface,
			},
			{
				name:      ifaceTwo.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceTwo.GatewayInterface,
			},
			{
				name:      ifaceThree.InterfaceName,
				ifaceType: common.InterfaceTypeGSM,
				isGateway: ifaceThree.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.NoError(t, err, "Backup should not return an error")
	assert.Len(t, h.backups, 2, "Expected to have 2 backup")
	mockDeviceManager.AssertExpectations(t)
	mockConnectionManager.AssertExpectations(t)

	expectedBackups := []*backup{
		{
			ifaceName:    ifaceOne.InterfaceName,
			ifaceType:    common.InterfaceTypeEthernet,
			wasGateway:   true,
			wasActive:    true,
			device:       mockDeviceOne,
			conn:         mockConnectionOne,
			connSettings: mockConnSettingsOne,
		},
		{
			ifaceName:    ifaceThree.InterfaceName,
			ifaceType:    common.InterfaceTypeGSM,
			wasGateway:   false,
			wasActive:    true,
			device:       mockDeviceThree,
			conn:         mockConnectionThree,
			connSettings: mockConnSettingsThree,
		},
	}
	backupComparer := cmp.Comparer(func(a, b *backup) bool {
		return a.ifaceName == b.ifaceName &&
			a.ifaceType == b.ifaceType &&
			a.wasGateway == b.wasGateway &&
			a.wasActive == b.wasActive &&
			reflect.DeepEqual(a.device, b.device) &&
			reflect.DeepEqual(a.conn, b.conn) &&
			compareConnSettingsBackups(a.connSettings, b.connSettings)
	})
	assert.True(t, cmp.Equal(expectedBackups, h.backups, backupComparer),
		"Backups do not match expected values")
}

func TestInterfaceStateHandler_Backup_ConnectionNotFound_Success(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnSettingsOne := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn1",
			"interface-name": ifaceOne.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5A",
		},
	}
	mockIsActiveOne := true

	mockConnectionThree := new(gonetworkmanager.MockConnection)
	mockConnSettingsThree := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn3",
			"interface-name": ifaceThree.InterfaceName,
			"type":           "gsm",
		},
		"ipv4": map[string]any{
			"route-metric": 100,
			"method":       "auto",
		},
		"gsm": map[string]any{
			"apn":      "internet",
			"pin":      "1234",
			"username": "user",
			"password": "pass",
		},
	}
	mockIsActiveThree := true

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		ifaceOne.InterfaceName:   mockDeviceOne,
		ifaceTwo.InterfaceName:   mockDeviceTwo,
		ifaceThree.InterfaceName: mockDeviceThree,
	}, nil).Once()

	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceOne).Return(mockConnectionOne, mockIsActiveOne, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceTwo).Return(nil, false, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceThree).Return(mockConnectionThree, mockIsActiveThree, nil).Once()

	mockConnectionManager.On("GetConnectionSettings", mockConnectionOne).Return(mockConnSettingsOne, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionThree).Return(mockConnSettingsThree, nil).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      ifaceOne.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceOne.GatewayInterface,
			},
			{
				name:      ifaceTwo.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceTwo.GatewayInterface,
			},
			{
				name:      ifaceThree.InterfaceName,
				ifaceType: common.InterfaceTypeGSM,
				isGateway: ifaceThree.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.NoError(t, err, "Backup should not return an error")
	assert.Len(t, h.backups, 2, "Expected to have 2 backup")
	mockDeviceManager.AssertExpectations(t)
	mockConnectionManager.AssertExpectations(t)

	expectedBackups := []*backup{
		{
			ifaceName:    ifaceOne.InterfaceName,
			ifaceType:    common.InterfaceTypeEthernet,
			wasGateway:   true,
			wasActive:    true,
			device:       mockDeviceOne,
			conn:         mockConnectionOne,
			connSettings: mockConnSettingsOne,
		},
		{
			ifaceName:    ifaceThree.InterfaceName,
			ifaceType:    common.InterfaceTypeGSM,
			wasGateway:   false,
			wasActive:    true,
			device:       mockDeviceThree,
			conn:         mockConnectionThree,
			connSettings: mockConnSettingsThree,
		},
	}
	backupComparer := cmp.Comparer(func(a, b *backup) bool {
		return a.ifaceName == b.ifaceName &&
			a.ifaceType == b.ifaceType &&
			a.wasGateway == b.wasGateway &&
			a.wasActive == b.wasActive &&
			reflect.DeepEqual(a.device, b.device) &&
			reflect.DeepEqual(a.conn, b.conn) &&
			compareConnSettingsBackups(a.connSettings, b.connSettings)
	})
	assert.True(t, cmp.Equal(expectedBackups, h.backups, backupComparer),
		"Backups do not match expected values")
}

func TestInterfaceStateHandler_Backup_GetConnectionSettings_Error(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnSettingsOne := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn1",
			"interface-name": ifaceOne.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5A",
		},
	}
	mockIsActiveOne := true

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnSettingsTwo := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn2",
			"interface-name": ifaceTwo.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 101,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5B",
		},
	}
	mockIsActiveTwo := false

	mockConnectionThree := new(gonetworkmanager.MockConnection)
	mockIsActiveThree := true

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		ifaceOne.InterfaceName:   mockDeviceOne,
		ifaceTwo.InterfaceName:   mockDeviceTwo,
		ifaceThree.InterfaceName: mockDeviceThree,
	}, nil).Once()

	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceOne).Return(mockConnectionOne, mockIsActiveOne, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceTwo).Return(mockConnectionTwo, mockIsActiveTwo, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceThree).Return(mockConnectionThree, mockIsActiveThree, nil).Once()

	mockConnectionManager.On("GetConnectionSettings", mockConnectionOne).Return(mockConnSettingsOne, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionTwo).Return(mockConnSettingsTwo, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionThree).Return(nil, errors.New("mock error")).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      ifaceOne.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceOne.GatewayInterface,
			},
			{
				name:      ifaceTwo.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceTwo.GatewayInterface,
			},
			{
				name:      ifaceThree.InterfaceName,
				ifaceType: common.InterfaceTypeGSM,
				isGateway: ifaceThree.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.Error(t, err, "Backup should return an error")
	assert.Len(t, h.backups, 0, "Expected to have 0 backup")
	mockDeviceManager.AssertExpectations(t)
	mockConnectionManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_Backup_FindConnectionWithStatus_Error(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceOne := new(gonetworkmanager.MockDevice)
	mockDeviceTwo := new(gonetworkmanager.MockDevice)
	mockDeviceThree := new(gonetworkmanager.MockDevice)

	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnSettingsOne := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn1",
			"interface-name": ifaceOne.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5A",
		},
	}
	mockIsActiveOne := true

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnSettingsTwo := nm.ConnectionSettings{
		"connection": map[string]any{
			"id":             "conn2",
			"interface-name": ifaceTwo.InterfaceName,
			"type":           "802-3-ethernet",
		},
		"ipv4": map[string]any{
			"route-metric": 101,
			"method":       "auto",
		},
		"802-3-ethernet": map[string]any{
			"mac-address": "00:11:22:33:44:5B",
		},
	}
	mockIsActiveTwo := false

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		ifaceOne.InterfaceName:   mockDeviceOne,
		ifaceTwo.InterfaceName:   mockDeviceTwo,
		ifaceThree.InterfaceName: mockDeviceThree,
	}, nil).Once()

	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceOne).Return(mockConnectionOne, mockIsActiveOne, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceTwo).Return(mockConnectionTwo, mockIsActiveTwo, nil).Once()
	mockConnectionManager.On("FindConnectionWithStatus", mockDeviceThree).Return(nil, false, errors.New("mock error")).Once()

	mockConnectionManager.On("GetConnectionSettings", mockConnectionOne).Return(mockConnSettingsOne, nil).Once()
	mockConnectionManager.On("GetConnectionSettings", mockConnectionTwo).Return(mockConnSettingsTwo, nil).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      ifaceOne.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceOne.GatewayInterface,
			},
			{
				name:      ifaceTwo.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceTwo.GatewayInterface,
			},
			{
				name:      ifaceThree.InterfaceName,
				ifaceType: common.InterfaceTypeGSM,
				isGateway: ifaceThree.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.Error(t, err, "Backup should return an error")
	assert.Len(t, h.backups, 0, "Expected to have 0 backup")
	mockDeviceManager.AssertExpectations(t)
	mockConnectionManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_Backup_DeviceNotExists(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		"testDeviceOne": new(gonetworkmanager.MockDevice),
	}, nil).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      iface.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: iface.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.NoError(t, err, "Backup should not return an error")
	assert.Len(t, h.backups, 0, "Expected to have 0 backup")
	mockDeviceManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_Backup_GetAvailableDevicesMap_Error(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
	}

	mockDeviceManager := new(networking.MockDeviceManager)

	mockDeviceManager.On("GetAvailableDevicesMap").Return(nil, errors.New("mock error")).Once()

	h := &InterfaceStateHandler{
		existingIfaceInfos: []*ifaceInfo{
			{
				name:      ifaceOne.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceOne.GatewayInterface,
			},
			{
				name:      ifaceTwo.InterfaceName,
				ifaceType: common.InterfaceTypeEthernet,
				isGateway: ifaceTwo.GatewayInterface,
			},
			{
				name:      ifaceThree.InterfaceName,
				ifaceType: common.InterfaceTypeGSM,
				isGateway: ifaceThree.GatewayInterface,
			},
		},
		backups:       make([]*backup, 0),
		deviceManager: mockDeviceManager,
	}

	// Act
	err := h.Backup()

	// Assert
	assert.Error(t, err, "Backup should return an error")
	assert.Len(t, h.backups, 0, "Expected to have 0 backup")
	mockDeviceManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_Restore_EthernetGateway_Success(t *testing.T) {
	// Arrange
	backupOne := &backup{
		ifaceName:  "eth0",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  true,
	}

	backupTwo := &backup{
		ifaceName:  "eth1",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: true,
		wasActive:  false,
	}

	backupThree := &backup{
		ifaceName:  "gsm0",
		ifaceType:  common.InterfaceTypeGSM,
		wasGateway: false,
		wasActive:  true,
	}

	backups := []*backup{backupOne, backupTwo, backupThree}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callOrder := make([]string, 0)

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreEthernetConnection_"+backup.ifaceName)
			return nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreGSMConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreGSMConnection_"+backup.ifaceName)
			return nil
		})

	h := &InterfaceStateHandler{
		backups: backups,
	}

	// Act
	err := h.Restore()

	// Assert
	assert.NoError(t, err, "Restore should not return an error")

	wantCallOrder := []string{
		"restoreEthernetConnection_eth1",
		"restoreEthernetConnection_eth0",
		"restoreGSMConnection_gsm0",
	}
	assert.Equal(t, wantCallOrder, callOrder, "unexpected mismatch in call order")

}

func TestInterfaceStateHandler_Restore_GSMGateway_Success(t *testing.T) {
	// Arrange
	backupOne := &backup{
		ifaceName:  "eth0",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  true,
	}

	backupTwo := &backup{
		ifaceName:  "eth1",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  false,
	}

	backupThree := &backup{
		ifaceName:  "gsm0",
		ifaceType:  common.InterfaceTypeGSM,
		wasGateway: true,
		wasActive:  true,
	}

	backups := []*backup{backupOne, backupTwo, backupThree}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callOrder := make([]string, 0)

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreEthernetConnection_"+backup.ifaceName)
			return nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreGSMConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreGSMConnection_"+backup.ifaceName)
			return nil
		})

	h := &InterfaceStateHandler{
		backups: backups,
	}

	// Act
	err := h.Restore()

	// Assert
	assert.NoError(t, err, "Restore should not return an error")

	wantCallOrder := []string{
		"restoreGSMConnection_gsm0",
		"restoreEthernetConnection_eth1",
		"restoreEthernetConnection_eth0",
	}
	assert.Equal(t, wantCallOrder, callOrder, "unexpected mismatch in call order")

}

func TestInterfaceStateHandler_Restore_NoReorderRequired_Success(t *testing.T) {
	// Arrange
	backupOne := &backup{
		ifaceName:  "eth0",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  true,
	}

	backupTwo := &backup{
		ifaceName:  "eth1",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  false,
	}

	backupThree := &backup{
		ifaceName:  "gsm0",
		ifaceType:  common.InterfaceTypeGSM,
		wasGateway: true,
		wasActive:  true,
	}

	backups := []*backup{backupThree, backupOne, backupTwo}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callOrder := make([]string, 0)

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreEthernetConnection_"+backup.ifaceName)
			return nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreGSMConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreGSMConnection_"+backup.ifaceName)
			return nil
		})

	h := &InterfaceStateHandler{
		backups: backups,
	}

	// Act
	err := h.Restore()

	// Assert
	assert.NoError(t, err, "Restore should not return an error")

	wantCallOrder := []string{
		"restoreGSMConnection_gsm0",
		"restoreEthernetConnection_eth0",
		"restoreEthernetConnection_eth1",
	}
	assert.Equal(t, wantCallOrder, callOrder, "unexpected mismatch in call order")

}

func TestInterfaceStateHandler_Restore_RestoreEthernetConnection_Error(t *testing.T) {
	// Arrange
	backupOne := &backup{
		ifaceName:  "eth0",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  true,
	}

	backupTwo := &backup{
		ifaceName:  "eth1",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: true,
		wasActive:  false,
	}

	backupThree := &backup{
		ifaceName:  "gsm0",
		ifaceType:  common.InterfaceTypeGSM,
		wasGateway: false,
		wasActive:  true,
	}

	backups := []*backup{backupOne, backupTwo, backupThree}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callOrder := make([]string, 0)
	callCount := 0

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreEthernetConnection_"+backup.ifaceName)
			callCount++

			if callCount == 2 {
				return errors.New("mock error")
			}
			return nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreGSMConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreGSMConnection_"+backup.ifaceName)
			return nil
		})

	h := &InterfaceStateHandler{
		backups: backups,
	}

	// Act
	err := h.Restore()

	// Assert
	assert.Error(t, err, "Restore should return an error")

	wantCallOrder := []string{
		"restoreEthernetConnection_eth1",
		"restoreEthernetConnection_eth0",
	}
	assert.Equal(t, wantCallOrder, callOrder, "unexpected mismatch in call order")
}

func TestInterfaceStateHandler_Restore_RestoreGSMConnection_Error(t *testing.T) {
	// Arrange
	backupOne := &backup{
		ifaceName:  "eth0",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: false,
		wasActive:  true,
	}

	backupTwo := &backup{
		ifaceName:  "eth1",
		ifaceType:  common.InterfaceTypeEthernet,
		wasGateway: true,
		wasActive:  false,
	}

	backupThree := &backup{
		ifaceName:  "gsm0",
		ifaceType:  common.InterfaceTypeGSM,
		wasGateway: false,
		wasActive:  true,
	}

	backups := []*backup{backupOne, backupTwo, backupThree}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callOrder := make([]string, 0)

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreEthernetConnection_"+backup.ifaceName)
			return nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreGSMConnection",
		func(_ *InterfaceStateHandler, backup *backup) error {
			callOrder = append(callOrder, "restoreGSMConnection_"+backup.ifaceName)
			return errors.New("mock error")
		})

	h := &InterfaceStateHandler{
		backups: backups,
	}

	// Act
	err := h.Restore()

	// Assert
	assert.Error(t, err, "Restore should return an error")

	wantCallOrder := []string{
		"restoreEthernetConnection_eth1",
		"restoreEthernetConnection_eth0",
		"restoreGSMConnection_gsm0",
	}
	assert.Equal(t, wantCallOrder, callOrder, "unexpected mismatch in call order")
}

func TestInterfaceStateHandler_restoreGSMConnection_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{}

	backup := &backup{
		ifaceName:    "gsm0",
		device:       mockDevice,
		conn:         mockConnection,
		connSettings: mockConnSettings,
	}

	mockConnectionManager := new(networking.MockConnectionManager)

	mockConnectionManager.On("UpdateConnection", mockConnection, mockConnSettings).Return(nil).Once()
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(nil).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreGSMConnection(backup)

	// Assert
	assert.NoError(t, err, "restoreGSMConnection should not return an error")
	mockConnectionManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_restoreGSMConnection_ActivateConnectionWithRetry_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{}

	backup := &backup{
		ifaceName:    "gsm0",
		device:       mockDevice,
		conn:         mockConnection,
		connSettings: mockConnSettings,
	}

	mockConnectionManager := new(networking.MockConnectionManager)

	mockConnectionManager.On("UpdateConnection", mockConnection, mockConnSettings).Return(nil).Once()
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreGSMConnection(backup)

	// Assert
	assert.Error(t, err, "restoreGSMConnection should return an error")
	mockConnectionManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_restoreGSMConnection_UpdateConnection_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{}

	backup := &backup{
		ifaceName:    "gsm0",
		device:       mockDevice,
		conn:         mockConnection,
		connSettings: mockConnSettings,
	}

	mockConnectionManager := new(networking.MockConnectionManager)

	mockConnectionManager.On("UpdateConnection", mockConnection, mockConnSettings).Return(errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreGSMConnection(backup)

	// Assert
	assert.Error(t, err, "restoreGSMConnection should return an error")
	mockConnectionManager.AssertExpectations(t)
}

func TestInterfaceStateHandler_restoreEthernetConnection_WithExistingConnection_Success(t *testing.T) {
	// Arrange
	mockConnectionManager := new(networking.MockConnectionManager)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockExistingConnection := new(gonetworkmanager.MockConnection)
	mockNewConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{}

	b := &backup{
		ifaceName:    "eth0",
		device:       mockDevice,
		connSettings: mockConnectionSettings,
	}

	mockConnectionManager.On("FindConnectionWithStatus", mockDevice).Return(mockExistingConnection, false, nil).Once()
	mockConnectionManager.On("DeleteConnection", mockExistingConnection).Return(nil).Once()
	mockConnectionManager.On("AddConnection", mockConnectionSettings).Return(mockNewConnection, nil).Once()

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnState",
		func(_ *InterfaceStateHandler, _ *backup, _ bool, conn nm.Connection) error {
			assert.Equal(t, mockNewConnection, conn, "should pass new connection")
			return nil
		})

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnection(b)

	// Assert
	assert.NoError(t, err, "restoreEthernetConnection should not return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnection_NoExistingConnection_Success(t *testing.T) {
	// Arrange
	mockConnectionManager := new(networking.MockConnectionManager)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockNewConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{}

	b := &backup{
		ifaceName:    "eth0",
		device:       mockDevice,
		connSettings: mockConnectionSettings,
	}

	mockConnectionManager.On("FindConnectionWithStatus", mockDevice).Return(nil, false, nil).Once()
	mockConnectionManager.On("AddConnection", mockConnectionSettings).Return(mockNewConnection, nil).Once()

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnState",
		func(_ *InterfaceStateHandler, _ *backup, _ bool, conn nm.Connection) error {
			assert.Equal(t, mockNewConnection, conn, "should pass new connection")
			return nil
		})

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnection(b)

	// Assert
	assert.NoError(t, err, "restoreEthernetConnection should not return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnection_restoreEthernetConnState_Error(t *testing.T) {
	// Arrange
	mockConnectionManager := new(networking.MockConnectionManager)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockExistingConnection := new(gonetworkmanager.MockConnection)
	mockNewConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{}

	b := &backup{
		ifaceName:    "eth0",
		device:       mockDevice,
		connSettings: mockConnectionSettings,
	}

	mockConnectionManager.On("FindConnectionWithStatus", mockDevice).Return(mockExistingConnection, false, nil).Once()
	mockConnectionManager.On("DeleteConnection", mockExistingConnection).Return(nil).Once()
	mockConnectionManager.On("AddConnection", mockConnectionSettings).Return(mockNewConnection, nil).Once()

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*InterfaceStateHandler](), "restoreEthernetConnState",
		func(_ *InterfaceStateHandler, _ *backup, _ bool, conn nm.Connection) error {
			assert.Equal(t, mockNewConnection, conn, "should pass new connection")
			return errors.New("restore error")
		})

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnection(b)

	// Assert
	assert.Error(t, err, "restoreEthernetConnection should return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnection_AddConnection_Error(t *testing.T) {
	// Arrange
	mockConnectionManager := new(networking.MockConnectionManager)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockExistingConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{}

	b := &backup{
		ifaceName:    "eth0",
		device:       mockDevice,
		connSettings: mockConnectionSettings,
	}

	mockConnectionManager.On("FindConnectionWithStatus", mockDevice).Return(mockExistingConnection, false, nil).Once()
	mockConnectionManager.On("DeleteConnection", mockExistingConnection).Return(nil).Once()
	mockConnectionManager.On("AddConnection", mockConnectionSettings).Return(nil, errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnection(b)

	// Assert
	assert.Error(t, err, "restoreEthernetConnection should return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnection_DeleteConnection_Error(t *testing.T) {
	// Arrange
	mockConnectionManager := new(networking.MockConnectionManager)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockExistingConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{}

	b := &backup{
		ifaceName:    "eth0",
		device:       mockDevice,
		connSettings: mockConnectionSettings,
	}

	mockConnectionManager.On("FindConnectionWithStatus", mockDevice).Return(mockExistingConnection, false, nil).Once()
	mockConnectionManager.On("DeleteConnection", mockExistingConnection).Return(errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnection(b)

	// Assert
	assert.Error(t, err, "restoreEthernetConnection should return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnection_FindConnectionWithStatus_Error(t *testing.T) {
	// Arrange
	mockConnectionManager := new(networking.MockConnectionManager)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockConnectionSettings := nm.ConnectionSettings{}

	b := &backup{
		ifaceName:    "eth0",
		device:       mockDevice,
		connSettings: mockConnectionSettings,
	}

	mockConnectionManager.On("FindConnectionWithStatus", mockDevice).Return(nil, false, errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnection(b)

	// Assert
	assert.Error(t, err, "restoreEthernetConnection should return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnState_IsActiveAndWasActive_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	b := &backup{
		wasActive: true,
		device:    mockDevice,
	}
	mockConnection := new(gonetworkmanager.MockConnection)
	isActive := true

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(nil).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnState(b, isActive, mockConnection)

	// Assert
	assert.NoError(t, err, "restoreEthernetConnState should not return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnState_IsNotActiveAndWasActive_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	b := &backup{
		wasActive: true,
		device:    mockDevice,
	}
	mockConnection := new(gonetworkmanager.MockConnection)
	isActive := false

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(nil).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnState(b, isActive, mockConnection)

	// Assert
	assert.NoError(t, err, "restoreEthernetConnState should not return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnState_IsActiveAndWasNotActive_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	b := &backup{
		wasActive: false,
		device:    mockDevice,
	}
	mockConnection := new(gonetworkmanager.MockConnection)
	isActive := true

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("DeactivateConnection", mockDevice).Return(nil).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnState(b, isActive, mockConnection)

	// Assert
	assert.NoError(t, err, "restoreEthernetConnState should not return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnState_IsNotActiveAndWasNotActive_Success(t *testing.T) {
	// Arrange
	b := &backup{
		wasActive: false,
		device:    new(gonetworkmanager.MockDevice),
	}
	mockConnection := new(gonetworkmanager.MockConnection)
	isActive := false

	h := InterfaceStateHandler{}

	// Act
	err := h.restoreEthernetConnState(b, isActive, mockConnection)

	// Assert
	assert.NoError(t, err, "restoreEthernetConnState should not return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnState_IsActiveAndWasNotActive_DeactivateConnection_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	b := &backup{
		wasActive: false,
		device:    mockDevice,
	}
	mockConnection := new(gonetworkmanager.MockConnection)
	isActive := true

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("DeactivateConnection", mockDevice).Return(errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnState(b, isActive, mockConnection)

	// Assert
	assert.Error(t, err, "restoreEthernetConnState should return an error")
}

func TestInterfaceStateHandler_restoreEthernetConnState_IsNotActiveAndWasActive_ActivateConnection_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	b := &backup{
		wasActive: true,
		device:    mockDevice,
	}
	mockConnection := new(gonetworkmanager.MockConnection)
	isActive := false

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(errors.New("mock error")).Once()

	h := InterfaceStateHandler{
		connManager: mockConnectionManager,
	}

	// Act
	err := h.restoreEthernetConnState(b, isActive, mockConnection)

	// Assert
	assert.Error(t, err, "restoreEthernetConnState should return an error")
}

func TestNewInterfaceStateHandler(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		InterfaceType:    v1.Interface_ETHERNET.Enum(),
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
		InterfaceType: v1.Interface_ETHERNET.Enum(),
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		InterfaceType: v1.Interface_GSM.Enum(),
	}

	ifaces := []*v1.Interface{ifaceOne, ifaceTwo, ifaceThree}
	nm := new(gonetworkmanager.MockNetworkManager)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	mockConnectionManager := new(networking.MockConnectionManager)
	patches.ApplyFunc(connectionmanager.NewConnectionManager, func() interfaces.ConnectionHelper {
		return mockConnectionManager
	})

	mockDeviceManager := new(networking.MockDeviceManager)
	patches.ApplyFunc(devicemanager.NewDeviceManager, func() interfaces.DeviceHelper {
		return mockDeviceManager
	})

	// Act
	got := NewInterfaceStateHandler(nm, ifaces)

	// Assert
	assert.NotNil(t, got, "NewInterfaceStateHandler should not return nil")
}

func Test_buildIfaceInfoList(t *testing.T) {
	// Arrange
	ifaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		InterfaceType:    v1.Interface_ETHERNET.Enum(),
	}

	ifaceTwo := &v1.Interface{
		InterfaceName: "eth1",
		InterfaceType: v1.Interface_ETHERNET.Enum(),
	}

	ifaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		InterfaceType: v1.Interface_GSM.Enum(),
	}

	ifaces := []*v1.Interface{ifaceOne, ifaceTwo, ifaceThree}

	// Act
	got := buildIfaceInfoList(ifaces)

	// Assert
	assert.Len(t, got, 3, "Expected to have 3 ifaceInfo entries")

	expected := []*ifaceInfo{
		{
			name:      ifaceOne.InterfaceName,
			ifaceType: common.InterfaceTypeEthernet,
			isGateway: ifaceOne.GatewayInterface,
		},
		{
			name:      ifaceTwo.InterfaceName,
			ifaceType: common.InterfaceTypeEthernet,
			isGateway: ifaceTwo.GatewayInterface,
		},
		{
			name:      ifaceThree.InterfaceName,
			ifaceType: common.InterfaceTypeGSM,
			isGateway: ifaceThree.GatewayInterface,
		},
	}
	ifaceInfoComparer := cmp.Comparer(func(a, b *ifaceInfo) bool {
		return a.name == b.name &&
			a.ifaceType == b.ifaceType &&
			a.isGateway == b.isGateway
	})
	assert.True(t, cmp.Equal(expected, got, ifaceInfoComparer),
		"ifaceInfo list does not match expected values")
}
