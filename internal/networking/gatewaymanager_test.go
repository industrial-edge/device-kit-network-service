/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/interfaces"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	"networkservice/internal/networking/mocks/networking"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestGatewayManager_Reset_Success(t *testing.T) {
	// Arrange
	mockDeviceManager := new(networking.MockDeviceManager)
	mockConnManager := new(networking.MockConnectionManager)

	interfaceName := "eth0"

	mockDevice := new(gonetworkmanager.MockDevice)
	mockOtherDevice := new(gonetworkmanager.MockDevice)

	mockConn := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{
			common.RouteMetricKey: 1,
		},
	}

	mockUpdatedConnSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{
			common.RouteMetricKey: common.RouteMetricLowestPriority,
		},
	}

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		interfaceName: mockDevice,
		"other0":      mockOtherDevice,
	}, nil)

	mockConnManager.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)
	mockConnManager.On("GetConnectionSettings", mockConn).Return(mockConnSettings, nil)
	mockConnManager.On("UpdateConnection", mockConn, mockUpdatedConnSettings).Return(nil)
	mockConnManager.On("ActivateConnectionWithRetry", mockConn, mockDevice).Return(nil)

	g := &GatewayManager{
		deviceManager: mockDeviceManager,
		connManager:   mockConnManager,
		connUtils:     connectionmanager.NewConnectionUtils(),
	}

	// Apply
	err := g.Reset(interfaceName)

	// Assert
	assert.NoError(t, err, " expected no error from Reset")
	mockDeviceManager.AssertExpectations(t)
	mockConnManager.AssertExpectations(t)
}

func TestGatewayManager_Reset_ActivateConnectionWithRetry_Error(t *testing.T) {
	// Arrange
	mockDeviceManager := new(networking.MockDeviceManager)
	mockConnManager := new(networking.MockConnectionManager)

	interfaceName := "eth0"

	mockDevice := new(gonetworkmanager.MockDevice)
	mockOtherDevice := new(gonetworkmanager.MockDevice)

	mockConn := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{
			common.RouteMetricKey: 1,
		},
	}

	mockUpdatedConnSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{
			common.RouteMetricKey: common.RouteMetricLowestPriority,
		},
	}

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		interfaceName: mockDevice,
		"other0":      mockOtherDevice,
	}, nil)

	mockConnManager.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)
	mockConnManager.On("GetConnectionSettings", mockConn).Return(mockConnSettings, nil)
	mockConnManager.On("UpdateConnection", mockConn, mockUpdatedConnSettings).Return(nil)
	mockConnManager.On("ActivateConnectionWithRetry", mockConn, mockDevice).Return(errors.New("mock error"))

	g := &GatewayManager{
		deviceManager: mockDeviceManager,
		connManager:   mockConnManager,
		connUtils:     connectionmanager.NewConnectionUtils(),
	}

	// Apply
	err := g.Reset(interfaceName)

	// Assert
	assert.Error(t, err, " expected an error from Reset")
	mockDeviceManager.AssertExpectations(t)
	mockConnManager.AssertExpectations(t)
}

func TestGatewayManager_Reset_UpdateConnection_Error(t *testing.T) {
	// Arrange
	mockDeviceManager := new(networking.MockDeviceManager)
	mockConnManager := new(networking.MockConnectionManager)

	interfaceName := "eth0"

	mockDevice := new(gonetworkmanager.MockDevice)
	mockOtherDevice := new(gonetworkmanager.MockDevice)

	mockConn := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{
			common.RouteMetricKey: 1,
		},
	}

	mockUpdatedConnSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{
			common.RouteMetricKey: common.RouteMetricLowestPriority,
		},
	}

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		interfaceName: mockDevice,
		"other0":      mockOtherDevice,
	}, nil)

	mockConnManager.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)
	mockConnManager.On("GetConnectionSettings", mockConn).Return(mockConnSettings, nil)
	mockConnManager.On("UpdateConnection", mockConn, mockUpdatedConnSettings).Return(errors.New("mock error"))

	g := &GatewayManager{
		deviceManager: mockDeviceManager,
		connManager:   mockConnManager,
		connUtils:     connectionmanager.NewConnectionUtils(),
	}

	// Apply
	err := g.Reset(interfaceName)

	// Assert
	assert.Error(t, err, " expected an error from Reset")
	mockDeviceManager.AssertExpectations(t)
	mockConnManager.AssertExpectations(t)
}

func TestGatewayManager_Reset_GetConnectionSettings_Error(t *testing.T) {
	// Arrange
	mockDeviceManager := new(networking.MockDeviceManager)
	mockConnManager := new(networking.MockConnectionManager)

	interfaceName := "eth0"

	mockDevice := new(gonetworkmanager.MockDevice)
	mockOtherDevice := new(gonetworkmanager.MockDevice)

	mockConn := new(gonetworkmanager.MockConnection)

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		interfaceName: mockDevice,
		"other0":      mockOtherDevice,
	}, nil)

	mockConnManager.On("FindConnectionWithStatus", mockDevice).Return(mockConn, true, nil)
	mockConnManager.On("GetConnectionSettings", mockConn).Return(nil, errors.New("mock error"))

	g := &GatewayManager{
		deviceManager: mockDeviceManager,
		connManager:   mockConnManager,
		connUtils:     connectionmanager.NewConnectionUtils(),
	}

	// Apply
	err := g.Reset(interfaceName)

	// Assert
	assert.Error(t, err, " expected an error from Reset")
	mockDeviceManager.AssertExpectations(t)
	mockConnManager.AssertExpectations(t)
}

func TestGatewayManager_Reset_FindConnectionWithStatus_Error(t *testing.T) {
	// Arrange
	mockDeviceManager := new(networking.MockDeviceManager)
	mockConnManager := new(networking.MockConnectionManager)

	interfaceName := "eth0"

	mockDevice := new(gonetworkmanager.MockDevice)
	mockOtherDevice := new(gonetworkmanager.MockDevice)

	mockDeviceManager.On("GetAvailableDevicesMap").Return(map[string]nm.Device{
		interfaceName: mockDevice,
		"other0":      mockOtherDevice,
	}, nil)

	mockConnManager.On("FindConnectionWithStatus", mockDevice).Return(nil, false, errors.New("mock error"))

	g := &GatewayManager{
		deviceManager: mockDeviceManager,
		connManager:   mockConnManager,
		connUtils:     connectionmanager.NewConnectionUtils(),
	}

	// Apply
	err := g.Reset(interfaceName)

	// Assert
	assert.Error(t, err, " expected an error from Reset")
	mockDeviceManager.AssertExpectations(t)
	mockConnManager.AssertExpectations(t)
}

func TestGatewayManager_Reset_GetAvailableDevicesMap_Error(t *testing.T) {
	// Arrange
	mockDeviceManager := new(networking.MockDeviceManager)

	interfaceName := "eth0"

	mockDeviceManager.On("GetAvailableDevicesMap").Return(nil, errors.New("mock error"))

	g := &GatewayManager{
		deviceManager: mockDeviceManager,
		connUtils:     connectionmanager.NewConnectionUtils(),
	}

	// Apply
	err := g.Reset(interfaceName)

	// Assert
	assert.Error(t, err, " expected an error from Reset")
	mockDeviceManager.AssertExpectations(t)
}

func TestNewGatewayManager(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	patches := gomonkey.ApplyFunc(connectionmanager.NewConnectionManager, func(nm nm.NetworkManager) interfaces.ConnectionHelper {
		return new(networking.MockConnectionManager)
	})
	defer patches.Reset()

	// Act
	got := NewGatewayManager(mockNetworkManager)

	// Assert
	assert.NotNil(t, got, " expected NewGatewayManager to return a non-nil value")
}
