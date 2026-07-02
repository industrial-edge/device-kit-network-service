/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package connectionmanager

import (
	"errors"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	"reflect"
	"testing"
	"time"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestConnectionManager_DeactivateConnection_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)
	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConn, nil)

	mockNetworkManager.On("DeactivateConnection", mockActiveConn).Return(nil)

	m := &ConnectionManager{
		nm: mockNetworkManager,
	}

	// Act
	err := m.DeactivateConnection(mockDevice)

	// Assert
	assert.NoError(t, err, "Expected no error when deactivating connection")
}

func TestConnectionManager_DeactivateConnection_NilActiveConn_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDevice.On("GetPropertyActiveConnection").Return(nil, nil)

	m := &ConnectionManager{
		nm: mockNetworkManager,
	}

	// Act
	err := m.DeactivateConnection(mockDevice)

	// Assert
	assert.NoError(t, err, "Expected no error when deactivating connection")
}

func TestConnectionManager_DeactivateConnection_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)
	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConn, nil)

	mockNetworkManager.On("DeactivateConnection", mockActiveConn).Return(errors.New("mock error"))

	m := &ConnectionManager{
		nm: mockNetworkManager,
	}

	// Act
	err := m.DeactivateConnection(mockDevice)

	// Assert
	assert.Error(t, err, "Expected an error when deactivating connection")
}

func TestConnectionManager_GetPropertyActiveConnection_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDevice.On("GetPropertyActiveConnection").Return(nil, errors.New("mock error"))

	m := &ConnectionManager{
		nm: mockNetworkManager,
	}

	// Act
	err := m.DeactivateConnection(mockDevice)

	// Assert
	assert.Error(t, err, "Expected an error when deactivating connection")
}

func TestConnectionManager_DeleteConnection_Success(t *testing.T) {
	// Arrange
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnection.On("Delete").Return(nil)

	m := ConnectionManager{}

	// Act
	err := m.DeleteConnection(mockConnection)

	// Assert
	assert.NoError(t, err, "Expected no error when deleting connection")
}

func TestConnectionManager_DeleteConnection_Error(t *testing.T) {
	// Arrange
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnection.On("Delete").Return(errors.New("mock error"))

	m := ConnectionManager{}

	// Act
	err := m.DeleteConnection(mockConnection)

	// Assert
	assert.Error(t, err, "Expected an error when deleting connection")
}

func TestConnectionManager_ActivateConnection(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)

	m := ConnectionManager{
		nm: mockNetworkManager,
	}

	mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, nil).Return(mockActiveConn, nil)
	mockActiveConn.On("GetPropertyState").Return(nm.NmActiveConnectionStateActivated, nil)

	// Act
	err := m.ActivateConnection(mockConnection, mockDevice)

	// Assert
	assert.NoError(t, err, "Expected no error when activating connection")
}

func TestConnectionManager_ActivateConnection_StateDeactivated(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)

	m := ConnectionManager{
		nm:           mockNetworkManager,
		maxRetries:   1,
		initialDelay: 1 * time.Millisecond,
	}

	mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, nil).Return(mockActiveConn, nil)
	mockActiveConn.On("GetPropertyState").Return(nm.NmActiveConnectionStateDeactivated, nil)

	// Act
	err := m.ActivateConnection(mockConnection, mockDevice)

	// Assert
	assert.Error(t, err, "Expected an error when activating connection")
	assert.EqualError(t, err, "connection not activated", "Expected error message to indicate connection was not activated")
}

func TestConnectionManager_ActivateConnection_GetPropertyState_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)

	m := ConnectionManager{
		nm:           mockNetworkManager,
		maxRetries:   1,
		initialDelay: 1 * time.Millisecond,
	}

	mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, nil).Return(mockActiveConn, nil)
	mockActiveConn.On("GetPropertyState").Return(nm.NmActiveConnectionStateUnknown, errors.New("mock error"))

	// Act
	err := m.ActivateConnection(mockConnection, mockDevice)

	// Assert
	assert.Error(t, err, "Expected an error when activating connection")
	assert.EqualError(t, err, "connection not activated", "Expected error message to indicate connection was not activated")
}

func TestConnectionManager_ActivateConnection_ActivateConnection_Error(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)

	m := ConnectionManager{
		nm: mockNetworkManager,
	}

	mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, nil).Return(nil, errors.New("mock error"))

	// Act
	err := m.ActivateConnection(mockConnection, mockDevice)

	// Assert
	assert.Error(t, err, "Expected an error when activating connection")
}

func TestConnectionManager_ActivateConnectionWithRetry_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)

	m := ConnectionManager{
		nm:           mockNetworkManager,
		maxRetries:   7,
		initialDelay: 1 * time.Second,
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*ConnectionManager](), "ActivateConnection",
		func(_ *ConnectionManager, connection nm.Connection, device nm.Device) error {
			assert.Equal(t, mockConnection, connection, "Expected the same connection to be passed to ActivateConnection")
			assert.Equal(t, mockDevice, device, "Expected the same device to be passed to ActivateConnection")
			return nil
		})

	// Act
	err := m.ActivateConnection(mockConnection, mockDevice)

	// Assert
	assert.NoError(t, err, "Expected no error when activating connection")
}

func TestConnectionManager_ActivateConnectionWithRetry_SuccessAfterFourFailures(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)

	m := ConnectionManager{
		nm:           mockNetworkManager,
		maxRetries:   7,
		initialDelay: 1 * time.Second,
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callCount := 0

	patches.ApplyMethod(reflect.TypeFor[*ConnectionManager](), "ActivateConnection",
		func(_ *ConnectionManager, connection nm.Connection, device nm.Device) error {
			assert.Equal(t, mockConnection, connection, "Expected the same connection to be passed to ActivateConnection")
			assert.Equal(t, mockDevice, device, "Expected the same device to be passed to ActivateConnection")
			callCount++
			if callCount < 5 {
				return errors.New("mock error")
			}
			return nil
		})

	wantDelays := []time.Duration{
		1 * time.Second,
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
	}
	patches.ApplyFunc(time.Sleep, func(d time.Duration) {
		assert.Equal(t, wantDelays[callCount-1], d, "Expected delay to match Fibonacci sequence")
	})

	// Act
	err := m.ActivateConnectionWithRetry(mockConnection, mockDevice)

	// Assert
	assert.NoError(t, err, "Expected no error when activating connection")
}

func TestConnectionManager_ActivateConnectionWithRetry_Failures(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockDevice := new(gonetworkmanager.MockDevice)

	m := ConnectionManager{
		nm:           mockNetworkManager,
		maxRetries:   7,
		initialDelay: 1 * time.Second,
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*ConnectionManager](), "ActivateConnection",
		func(_ *ConnectionManager, connection nm.Connection, device nm.Device) error {
			assert.Equal(t, mockConnection, connection, "Expected the same connection to be passed to ActivateConnection")
			assert.Equal(t, mockDevice, device, "Expected the same device to be passed to ActivateConnection")
			return errors.New("mock error")
		})

	callCount := 0
	wantDelays := []time.Duration{
		1 * time.Second,
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
		8 * time.Second,
		13 * time.Second,
	}
	patches.ApplyFunc(time.Sleep, func(d time.Duration) {
		assert.Equal(t, wantDelays[callCount], d, "Expected delay to match Fibonacci sequence")
		callCount++
	})

	// Act
	err := m.ActivateConnectionWithRetry(mockConnection, mockDevice)

	// Assert
	assert.Error(t, err, "Expected error when activating connection")
}

func TestConnectionManager_AddConnection_Success(t *testing.T) {
	// Arrange
	mockSettings := new(gonetworkmanager.MockSettings)
	mockConn := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{}

	m := ConnectionManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*ConnectionManager](), "newNetworkManagerSettings",
		func(_ *ConnectionManager) (nm.Settings, error) {
			return mockSettings, nil
		})

	mockSettings.On("AddConnection", mockConnectionSettings).Return(mockConn, nil)

	// Act
	got, err := m.AddConnection(mockConnectionSettings)

	// Assert
	assert.NoError(t, err, "Expected no error when adding connection")
	assert.Equal(t, mockConn, got, "Expected the returned connection to match the mock connection")
}

func TestConnectionManager_AddConnection_Error(t *testing.T) {
	// Arrange
	mockSettings := new(gonetworkmanager.MockSettings)
	mockConnectionSettings := nm.ConnectionSettings{}

	m := ConnectionManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*ConnectionManager](), "newNetworkManagerSettings",
		func(_ *ConnectionManager) (nm.Settings, error) {
			return mockSettings, nil
		})

	mockSettings.On("AddConnection", mockConnectionSettings).Return(nil, errors.New("mock error"))

	// Act
	got, err := m.AddConnection(mockConnectionSettings)

	// Assert
	assert.Error(t, err, "Expected error when adding connection")
	assert.Nil(t, got, "Expected the returned connection to be nil")
}

func TestConnectionManager_AddConnection_newNetworkManagerSettings_Error(t *testing.T) {
	// Arrange
	mockConnectionSettings := nm.ConnectionSettings{}

	m := ConnectionManager{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*ConnectionManager](), "newNetworkManagerSettings",
		func(_ *ConnectionManager) (nm.Settings, error) {
			return nil, errors.New("mock error")
		})

	// Act
	got, err := m.AddConnection(mockConnectionSettings)

	// Assert
	assert.Error(t, err, "Expected error when adding connection")
	assert.Nil(t, got, "Expected the returned connection to be nil")
}

func TestConnectionManager_UpdateConnection_Success(t *testing.T) {
	// Arrange
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{}

	mockConnection.On("Update", mockConnSettings).Return(nil)
	m := ConnectionManager{}

	// Act
	err := m.UpdateConnection(mockConnection, mockConnSettings)

	// Assert
	assert.NoError(t, err, "Expected no error when updating connection")
}

func TestConnectionManager_UpdateConnection_Error(t *testing.T) {
	// Arrange
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{}

	mockConnection.On("Update", mockConnSettings).Return(errors.New("mock error"))
	m := ConnectionManager{}

	// Act
	err := m.UpdateConnection(mockConnection, mockConnSettings)

	// Assert
	assert.Error(t, err, "Expected error when updating connection")
}

func TestConnectionManager_GetConnectionSettings_Success(t *testing.T) {
	// Arrange
	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnSettings := nm.ConnectionSettings{}

	mockConnection.On("GetSettings").Return(mockConnSettings, nil)
	m := ConnectionManager{}

	// Act
	got, err := m.GetConnectionSettings(mockConnection)
	// Assert
	assert.NoError(t, err, "Expected no error when getting connection settings")
	assert.Equal(t, mockConnSettings, got, "Expected the returned connection settings to match the mock settings")
}

func TestConnectionManager_GetConnectionSettings_Error(t *testing.T) {
	// Arrange
	mockConnection := new(gonetworkmanager.MockConnection)

	mockConnection.On("GetSettings").Return(nil, errors.New("mock error"))
	m := ConnectionManager{}

	// Act
	got, err := m.GetConnectionSettings(mockConnection)

	// Assert
	assert.Error(t, err, "Expected error when getting connection settings")
	assert.Nil(t, got, "Expected the returned connection settings to be nil")
}

func TestConnectionManager_FindConnectionWithStatus_ActiveConnection_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)
	mockConnection := new(gonetworkmanager.MockConnection)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)

	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConn, nil)
	mockActiveConn.On("GetPropertyConnection").Return(mockConnection, nil)
	mockActiveConn.On("GetPropertyState").Return(nm.NmActiveConnectionStateActivated, nil)

	m := ConnectionManager{}

	// Act
	got, got2, gotErr := m.FindConnectionWithStatus(mockDevice)

	// Assert
	assert.NoError(t, gotErr, "Expected no error when finding active connection")
	assert.Equal(t, mockConnection, got, "Expected the returned connection to match the mock connection")
	assert.True(t, got2, "Expected the returned state to be Activated")
}

func TestConnectionManager_FindConnectionWithStatus_ActiveConnection_GetPropertyConnection_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)
	mockActiveConn := new(gonetworkmanager.MockActiveConnection)

	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConn, nil)
	mockActiveConn.On("GetPropertyConnection").Return(nil, errors.New("mock error"))

	m := ConnectionManager{}

	// Act
	got, got2, gotErr := m.FindConnectionWithStatus(mockDevice)

	// Assert
	assert.Error(t, gotErr, "Expected error when finding active connection")
	assert.Nil(t, got, "Expected the returned connection to be nil")
	assert.False(t, got2, "Expected the returned state to not be Activated")
}

func TestConnectionManager_FindConnectionWithStatus_ActiveConnection_GetPropertyActiveConnection_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	mockDevice.On("GetPropertyActiveConnection").Return(nil, errors.New("mock error"))

	m := ConnectionManager{}

	// Act
	got, got2, gotErr := m.FindConnectionWithStatus(mockDevice)

	// Assert
	assert.Error(t, gotErr, "Expected error when finding active connection")
	assert.Nil(t, got, "Expected the returned connection to be nil")
	assert.False(t, got2, "Expected the returned state to not be Activated")
}

func TestConnectionManager_FindConnectionWithStatus_InactiveConnection_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)
	mockConnection := new(gonetworkmanager.MockConnection)

	mockDevice.On("GetPropertyActiveConnection").Return(nil, nil)
	mockDevice.On("GetPropertyAvailableConnections").Return([]nm.Connection{mockConnection}, nil)

	m := ConnectionManager{}

	// Act
	got, got2, gotErr := m.FindConnectionWithStatus(mockDevice)

	// Assert
	assert.NoError(t, gotErr, "Expected no error when finding active connection")
	assert.Equal(t, mockConnection, got, "Expected the returned connection to match the mock connection")
	assert.False(t, got2, "Expected the returned state to not be Activated")
}

func TestConnectionManager_FindConnectionWithStatus_NoConnection_Success(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	mockDevice.On("GetPropertyActiveConnection").Return(nil, nil)
	mockDevice.On("GetPropertyAvailableConnections").Return([]nm.Connection{}, nil)

	m := ConnectionManager{}

	// Act
	got, got2, gotErr := m.FindConnectionWithStatus(mockDevice)

	// Assert
	assert.NoError(t, gotErr, "Expected no error when finding active connection")
	assert.Nil(t, got, "Expected the returned connection to be nil")
	assert.False(t, got2, "Expected the returned state to not be Activated")
}

func TestConnectionManager_FindConnectionWithStatus_GetPropertyAvailableConnections_Error(t *testing.T) {
	// Arrange
	mockDevice := new(gonetworkmanager.MockDevice)

	mockDevice.On("GetPropertyActiveConnection").Return(nil, nil)
	mockDevice.On("GetPropertyAvailableConnections").Return([]nm.Connection{}, errors.New("mock error"))

	m := ConnectionManager{}

	// Act
	got, got2, gotErr := m.FindConnectionWithStatus(mockDevice)

	// Assert
	assert.Error(t, gotErr, "Expected error when finding active connection")
	assert.Nil(t, got, "Expected the returned connection to be nil")
	assert.False(t, got2, "Expected the returned state to not be Activated")
}

func TestNewConnectionManager(t *testing.T) {
	// Arrange
	nm := new(gonetworkmanager.MockNetworkManager)

	// Act
	got := NewConnectionManager(nm)

	// Assert
	assert.NotNil(t, got, "Expected the ConnectionManager instance to not be nil")
}
