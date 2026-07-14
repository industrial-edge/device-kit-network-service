/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package factory

import (
	"testing"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/gsm"
	"networkservice/internal/networking/interfaces"
	"networkservice/internal/networking/mocks/gonetworkmanager"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestConnectionFactory_WithGSMConnectionType_ReturnsGSMHandler(t *testing.T) {
	// Arrange
	mockNM := &gonetworkmanager.MockNetworkManager{}
	mockSettings := new(gonetworkmanager.MockSettings)
	connectionType := v1.ConnectionSettings_GSM

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock the NewSettings call that happens inside NewGSMHandler
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	// Act
	handler := ConnectionFactory(connectionType, mockNM)

	// Assert
	assert.NotNil(t, handler, "Expected non-nil handler for GSM connection type")
	assert.IsType(t, &gsm.GSMConnection{}, handler, "Expected GSMConnection type for GSM connection")
}

func TestConnectionFactory_WithGSMConnectionType_ReturnsNilWhenNewSettingsFails(t *testing.T) {
	// Arrange
	mockNM := &gonetworkmanager.MockNetworkManager{}

	connectionType := v1.ConnectionSettings_GSM

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock NewGSMHandler to return error
	patches.ApplyFunc(gsm.NewGSMHandler, func(networkManager nm.NetworkManager) (interfaces.ConnectionHandler, error) {
		return nil, assert.AnError
	})

	// Act
	handler := ConnectionFactory(connectionType, mockNM)

	// Assert
	assert.Nil(t, handler, "Expected nil handler when NewGSMHandler fails")
}

func TestConnectionFactory_WithUnknownConnectionType_ReturnsNil(t *testing.T) {
	// Arrange
	mockNM := &gonetworkmanager.MockNetworkManager{}

	connectionType := v1.ConnectionSettings_UNKNOWN

	// Act
	handler := ConnectionFactory(connectionType, mockNM)

	// Assert
	assert.Nil(t, handler, "Expected nil handler for unknown connection type")
}

func TestConnectionFactory_WithDefaultCase_ReturnsNil(t *testing.T) {
	// Arrange
	mockNM := &gonetworkmanager.MockNetworkManager{}

	// Use an undefined connection type to test default case
	connectionType := v1.ConnectionSettings_ConnectionType(999)

	// Act
	handler := ConnectionFactory(connectionType, mockNM)

	// Assert
	assert.Nil(t, handler, "Expected nil handler for undefined connection type")
}

func TestConnectionFactory_InterfaceCompliance(t *testing.T) {
	// Arrange
	mockNM := &gonetworkmanager.MockNetworkManager{}
	mockSettings := new(gonetworkmanager.MockSettings)
	connectionType := v1.ConnectionSettings_GSM

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	// Act
	handler := ConnectionFactory(connectionType, mockNM)

	// Assert
	assert.NotNil(t, handler, "Handler should not be nil")
	assert.Implements(t, (*interfaces.ConnectionHandler)(nil), handler, "Handler should implement ConnectionHandler interface")

	// Verify the handler has the required methods
	_, hasCreateConnection := handler.(interface {
		CreateConnection(settings *v1.ConnectionSettings) error
	})
	assert.True(t, hasCreateConnection, "Handler should have CreateConnection method")

	_, hasRemoveConnection := handler.(interface {
		RemoveConnection(settings *v1.ConnectionSettings) error
	})
	assert.True(t, hasRemoveConnection, "Handler should have RemoveConnection method")
}
