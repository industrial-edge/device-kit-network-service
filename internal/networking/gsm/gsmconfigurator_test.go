/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"errors"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	"networkservice/internal/networking/mocks/networking"
	"reflect"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestGSMConfigurator_retrievingConnectionDetails_UsernameDifferent(t *testing.T) {
	// Arrange
	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnectionOneSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"id": "eno0_DHCP",
		},
	}
	mockConnectionOne.On("GetSettings").Return(mockConnectionOneSettings, nil)

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnectionTwoSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": "test-gsm-device",
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      "internet",
			"username": "user",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"dns":          []uint32{134744072, 16843009},
		},
	}
	mockConnectionTwo.On("GetSettings").Return(mockConnectionTwoSettings, nil)

	mockConnections := []nm.Connection{mockConnectionOne, mockConnectionTwo}
	mockSettings := new(gonetworkmanager.MockSettings)
	mockSettings.On("ListConnections").Return(mockConnections, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return mockSettings, nil
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet", "different-user")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Nil(t, err, "expected no error")
}

func TestGSMConfigurator_retrievingConnectionDetails_ExistingConnectionWithUsername_IncommingConfigWithoutUsername(t *testing.T) {
	// Arrange
	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnectionOneSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"id": "eno0_DHCP",
		},
	}
	mockConnectionOne.On("GetSettings").Return(mockConnectionOneSettings, nil)

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnectionTwoSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": "test-gsm-device",
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      "internet",
			"username": "user",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"dns":          []uint32{134744072, 16843009},
		},
	}
	mockConnectionTwo.On("GetSettings").Return(mockConnectionTwoSettings, nil)

	mockConnections := []nm.Connection{mockConnectionOne, mockConnectionTwo}
	mockSettings := new(gonetworkmanager.MockSettings)
	mockSettings.On("ListConnections").Return(mockConnections, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return mockSettings, nil
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet", "")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Nil(t, err, "expected no error")
}

func TestGSMConfigurator_retrievingConnectionDetails_ExistingConnectionWithoutUsername_IncommingConfigWithUsername(t *testing.T) {
	// Arrange
	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnectionOneSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"id": "eno0_DHCP",
		},
	}
	mockConnectionOne.On("GetSettings").Return(mockConnectionOneSettings, nil)

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnectionTwoSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": "test-gsm-device",
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn": "internet",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"dns":          []uint32{134744072, 16843009},
		},
	}
	mockConnectionTwo.On("GetSettings").Return(mockConnectionTwoSettings, nil)

	mockConnections := []nm.Connection{mockConnectionOne, mockConnectionTwo}
	mockSettings := new(gonetworkmanager.MockSettings)
	mockSettings.On("ListConnections").Return(mockConnections, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return mockSettings, nil
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet", "user")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Nil(t, err, "expected no error")
}

func TestGSMConfigurator_retrievingConnectionDetails_ApnDifferent(t *testing.T) {
	// Arrange
	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnectionOneSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"id": "eno0_DHCP",
		},
	}
	mockConnectionOne.On("GetSettings").Return(mockConnectionOneSettings, nil)

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnectionTwoSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": "test-gsm-device",
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      "internet",
			"username": "user",
		},
		"ipv4": map[string]any{
			"route-metric": 1,
			"dns":          []uint32{134744072, 16843009},
		},
	}
	mockConnectionTwo.On("GetSettings").Return(mockConnectionTwoSettings, nil)

	mockConnections := []nm.Connection{mockConnectionOne, mockConnectionTwo}
	mockSettings := new(gonetworkmanager.MockSettings)
	mockSettings.On("ListConnections").Return(mockConnections, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return mockSettings, nil
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet-other", "user")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Nil(t, err, "expected no error")
}

func TestGSMConfigurator_retrievingConnectionDetails_GetSettings_Error(t *testing.T) {
	// Arrange
	mockConnectionOne := new(gonetworkmanager.MockConnection)
	mockConnectionOne.On("GetSettings").Return(nil, errors.New("mock error"))

	mockConnectionTwo := new(gonetworkmanager.MockConnection)
	mockConnectionTwo.On("GetSettings").Return(nil, errors.New("mock error"))

	mockConnections := []nm.Connection{mockConnectionOne, mockConnectionTwo}
	mockSettings := new(gonetworkmanager.MockSettings)
	mockSettings.On("ListConnections").Return(mockConnections, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return mockSettings, nil
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet", "user")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Nil(t, err, "expected no error")
}

func TestGSMConfigurator_retrievingConnectionDetails_ListConnections_Error(t *testing.T) {
	// Arrange
	mockSettings := new(gonetworkmanager.MockSettings)
	mockSettings.On("ListConnections").Return(nil, errors.New("mock error"))

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return mockSettings, nil
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet", "user")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Error(t, err, "expected error")
}

func TestGSMConfigurator_retrievingConnectionDetails_newNetworkManagerSettings_Error(t *testing.T) {
	// Arrange
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "newNetworkManagerSettings",
		func(_ *GSMConfigurator) (nm.Settings, error) {
			return nil, errors.New("mock error")
		})

	c := &GSMConfigurator{}

	// Act
	conn, connSettings, err := c.retrievingConnectionDetails("internet", "user")

	// Assert
	assert.Nil(t, conn, "expected no matching connection")
	assert.Nil(t, connSettings, "expected no matching connection settings")
	assert.Error(t, err, "expected error")
}

func TestNewGSMConfigurator_WithDNS_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	iface := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "8.8.8.8",
			SecondaryDNS: "1.1.1.1",
		},
	}

	// Act
	got := NewGSMConfigurator(mockNetworkManager, iface)

	// Assert
	assert.NotNil(t, got, "expected non-nil GSMConnector")

	assert.True(t, reflect.DeepEqual(got.gsmConfig, iface.GsmConfiguration), "gsmConfig do not match expected")
	assert.True(t, reflect.DeepEqual(got.dnsConfig, iface.DNSConfig), "dnsConfig do not match expected")
}

func TestNewGSMConfigurator_WithoutDNS_Success(t *testing.T) {
	// Arrange
	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	iface := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
	}

	// Act
	got := NewGSMConfigurator(mockNetworkManager, iface)

	// Assert
	assert.NotNil(t, got, "expected non-nil GSMConnector")

	assert.True(t, reflect.DeepEqual(got.gsmConfig, iface.GsmConfiguration), "gsmConfig do not match expected")
	assert.Nil(t, got.dnsConfig, "dnsConfig expected to be nil")
}

func TestGSMConfigurator_retrieveApn_RainyDayScenarios(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		settings nm.ConnectionSettings
		want     string
	}{
		{
			name:     "NoGSMSettings",
			settings: nm.ConnectionSettings{},
			want:     "",
		},
		{
			name: "NoAPNInGSMSettings",
			settings: nm.ConnectionSettings{
				"gsm": map[string]any{
					"username": "user",
				},
			},
			want: "",
		},
		{
			name: "APNNotString",
			settings: nm.ConnectionSettings{
				"gsm": map[string]any{
					"apn": 1234,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := new(GSMConfigurator)
			got := c.retrieveApn(tt.settings)

			assert.Equal(t, tt.want, got, "retrieveApn() = %v, want %v", got, tt.want)
		})
	}
}

func TestGSMConfigurator_retrieveUsername_RainyDayScenarios(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		settings nm.ConnectionSettings
		want     string
	}{
		{
			name:     "NoGSMSettings",
			settings: nm.ConnectionSettings{},
			want:     "",
		},
		{
			name: "NoUsernameInGSMSettings",
			settings: nm.ConnectionSettings{
				"gsm": map[string]any{
					"apn": "internet",
				},
			},
			want: "",
		},
		{
			name: "UsernameNotString",
			settings: nm.ConnectionSettings{
				"gsm": map[string]any{
					"username": 1234,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := new(GSMConfigurator)
			got := c.retrieveUsername(tt.settings)

			assert.Equal(t, tt.want, got, "retrieveUsername() = %v, want %v", got, tt.want)
		})
	}
}

func TestGSMConfigurator_Configure_Gateway_WithDNS_Success(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "1.1.1.1",
	}

	testDeviceName := "test-gsm-device"

	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockNewConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
		"ipv4": map[string]any{
			"route-metric":    common.RouteMetricHighestPriority,
			"dns":             []uint32{134744072, 16843009},
			"ignore-auto-dns": 1,
		},
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("UpdateConnection", mockConnection, mockNewConnectionSettings).Return(nil)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDeviceManager := new(networking.MockDeviceManager)
	mockDeviceManager.On("GetGSMDevice", testDeviceName).Return(mockDevice, nil)

	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return mockConnection, mockConnectionSettings, nil
		})

	c := &GSMConfigurator{
		isGateway:     true,
		gsmConfig:     gsmConfig,
		dnsConfig:     dnsConfig,
		connUtils:     connectionmanager.NewConnectionUtils(),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.NoError(t, err, "expected no error from Configure")
}

func TestGSMConfigurator_Configure_NotGateway_WithoutDNS_Success(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	testDeviceName := "test-gsm-device"

	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockNewConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("UpdateConnection", mockConnection, mockNewConnectionSettings).Return(nil)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDeviceManager := new(networking.MockDeviceManager)
	mockDeviceManager.On("GetGSMDevice", testDeviceName).Return(mockDevice, nil)

	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return mockConnection, mockConnectionSettings, nil
		})

	c := &GSMConfigurator{
		gsmConfig:     gsmConfig,
		dnsConfig:     nil,
		connUtils:     connectionmanager.NewConnectionUtils(),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.NoError(t, err, "expected no error from Configure")
}

func TestGSMConfigurator_Configure_WithOnlyPrimaryDNS_Success(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS: "8.8.8.8",
	}

	testDeviceName := "test-gsm-device"

	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockNewConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
		"ipv4": map[string]any{
			"dns":             []uint32{134744072},
			"ignore-auto-dns": 1,
		},
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("UpdateConnection", mockConnection, mockNewConnectionSettings).Return(nil)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDeviceManager := new(networking.MockDeviceManager)
	mockDeviceManager.On("GetGSMDevice", testDeviceName).Return(mockDevice, nil)

	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return mockConnection, mockConnectionSettings, nil
		})

	c := &GSMConfigurator{
		gsmConfig:     gsmConfig,
		dnsConfig:     dnsConfig,
		connUtils:     connectionmanager.NewConnectionUtils(),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.NoError(t, err, "expected no error from Configure")
}

func TestGSMConfigurator_Configure_ActivateConnectionWithRetry_Error(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "1.1.1.1",
	}

	testDeviceName := "test-gsm-device"

	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockNewConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
		"ipv4": map[string]any{
			"route-metric":    common.RouteMetricHighestPriority,
			"dns":             []uint32{134744072, 16843009},
			"ignore-auto-dns": 1,
		},
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("UpdateConnection", mockConnection, mockNewConnectionSettings).Return(nil)

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDeviceManager := new(networking.MockDeviceManager)
	mockDeviceManager.On("GetGSMDevice", testDeviceName).Return(mockDevice, nil)

	mockConnectionManager.On("ActivateConnectionWithRetry", mockConnection, mockDevice).Return(errors.New("mock error"))

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return mockConnection, mockConnectionSettings, nil
		})

	c := &GSMConfigurator{
		isGateway:     true,
		gsmConfig:     gsmConfig,
		dnsConfig:     dnsConfig,
		connUtils:     connectionmanager.NewConnectionUtils(),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error from Configure")
}

func TestGSMConfigurator_Configure_UpdateConnection_Error(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "1.1.1.1",
	}

	testDeviceName := "test-gsm-device"

	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockNewConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
		"ipv4": map[string]any{
			"route-metric":    common.RouteMetricHighestPriority,
			"dns":             []uint32{134744072, 16843009},
			"ignore-auto-dns": 1,
		},
	}

	mockConnectionManager := new(networking.MockConnectionManager)
	mockConnectionManager.On("UpdateConnection", mockConnection, mockNewConnectionSettings).Return(errors.New("mock error"))

	mockDevice := new(gonetworkmanager.MockDevice)
	mockDeviceManager := new(networking.MockDeviceManager)
	mockDeviceManager.On("GetGSMDevice", testDeviceName).Return(mockDevice, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return mockConnection, mockConnectionSettings, nil
		})

	c := &GSMConfigurator{
		isGateway:     true,
		gsmConfig:     gsmConfig,
		dnsConfig:     dnsConfig,
		connUtils:     connectionmanager.NewConnectionUtils(),
		connManager:   mockConnectionManager,
		deviceManager: mockDeviceManager,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error from Configure")
}

func TestGSMConfigurator_Configure_GetGSMDevice_Error(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "1.1.1.1",
	}

	testDeviceName := "test-gsm-device"

	mockConnection := new(gonetworkmanager.MockConnection)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": testDeviceName,
			"id":             "iedk-gsm-connection",
			"type":           "gsm",
		},
		"gsm": map[string]any{
			"apn":      testApn,
			"username": testUsername,
		},
	}

	mockDeviceManager := new(networking.MockDeviceManager)
	mockDeviceManager.On("GetGSMDevice", testDeviceName).Return(nil, errors.New("mock error"))

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return mockConnection, mockConnectionSettings, nil
		})

	c := &GSMConfigurator{
		isGateway:     true,
		gsmConfig:     gsmConfig,
		dnsConfig:     dnsConfig,
		connUtils:     connectionmanager.NewConnectionUtils(),
		deviceManager: mockDeviceManager,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error from Configure")
}

func TestGSMConfigurator_Configure_GSMConnectionNotExist_Error(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "1.1.1.1",
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return nil, nil, nil
		})

	c := &GSMConfigurator{
		isGateway: true,
		gsmConfig: gsmConfig,
		dnsConfig: dnsConfig,
		connUtils: connectionmanager.NewConnectionUtils(),
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error from Configure")
	assert.EqualError(t, err, "no GSM connection found; use CreateConnection RPC first",
		"error message does not match expected")
}

func TestGSMConfigurator_Configure_retrievingConnectionDetails_Error(t *testing.T) {
	// Arrange
	testApn := "internet"
	testUsername := "user"

	gsmConfig := &v1.Interface_GsmConf{
		Apn:      testApn,
		Pin:      "1234",
		Username: testUsername,
		Password: "pass",
	}

	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "1.1.1.1",
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*GSMConfigurator](), "retrievingConnectionDetails",
		func(_ *GSMConfigurator, apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
			assert.Equal(t, testApn, apn, "APN passed to retrievingConnectionDetails does not match expected")
			assert.Equal(t, testUsername, username, "Username passed to retrievingConnectionDetails does not match expected")
			return nil, nil, errors.New("mock error")
		})

	c := &GSMConfigurator{
		isGateway: true,
		gsmConfig: gsmConfig,
		dnsConfig: dnsConfig,
		connUtils: connectionmanager.NewConnectionUtils(),
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error from Configure")
}

func TestGSMConfigurator_isGSMConnection_NoConnectionSettings(t *testing.T) {
	// Arrange
	settings := nm.ConnectionSettings{}
	c := &GSMConfigurator{}

	// Act
	got := c.isGSMConnection(settings)

	// Assert
	assert.False(t, got, "expected isGSMConnection to return false when no connection settings are present")
}
