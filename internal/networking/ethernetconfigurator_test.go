/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	"reflect"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestEthernetConfigurator_Configure_Success(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		Label:            "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
		GatewayInterface: true,
	}

	mockDevice := new(gonetworkmanager.MockDeviceWired)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"id": "eno0",
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "prepareSettings",
		func(_ *NetworkConfigurator, iface *v1.Interface, device nm.DeviceWired) (nm.ConnectionSettings, error) {
			return mockConnectionSettings, nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "updateConnections",
		func(_ *NetworkConfigurator, device nm.DeviceWired, settings nm.ConnectionSettings) error {
			return nil
		})

	c := &EthernetConfigurator{
		device: mockDevice,
		iface:  iface,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.NoError(t, err, "unexpected error")
}

func TestEthernetConfigurator_Configure_updateConnections_Error(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		Label:            "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
		GatewayInterface: true,
	}

	mockDevice := new(gonetworkmanager.MockDeviceWired)
	mockConnectionSettings := nm.ConnectionSettings{
		"connection": map[string]any{
			"id": "eno0",
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "prepareSettings",
		func(_ *NetworkConfigurator, iface *v1.Interface, device nm.DeviceWired) (nm.ConnectionSettings, error) {
			return mockConnectionSettings, nil
		})

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "updateConnections",
		func(_ *NetworkConfigurator, device nm.DeviceWired, settings nm.ConnectionSettings) error {
			return errors.New("mock error")
		})

	c := &EthernetConfigurator{
		device: mockDevice,
		iface:  iface,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error")
}

func TestEthernetConfigurator_Configure_prepareSettings_Error(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		Label:            "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
		GatewayInterface: true,
	}

	mockDevice := new(gonetworkmanager.MockDeviceWired)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "prepareSettings",
		func(_ *NetworkConfigurator, iface *v1.Interface, device nm.DeviceWired) (nm.ConnectionSettings, error) {
			return nil, errors.New("mock error")
		})

	c := &EthernetConfigurator{
		device: mockDevice,
		iface:  iface,
	}

	// Act
	err := c.Configure()

	// Assert
	assert.Error(t, err, "expected error")
}

func TestNewEthernetConfigurator_Success(t *testing.T) {
	// Arrange
	nc := new(NetworkConfigurator)
	iface := &v1.Interface{
		Label:      "eth0",
		MacAddress: "00:0A:95:9D:68:16",
		DHCP:       "enabled",
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	mockDevice := new(gonetworkmanager.MockDeviceWired)
	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "getDeviceBy",
		func(_ *NetworkConfigurator, iface *v1.Interface) (nm.DeviceWired, error) {
			return mockDevice, nil
		},
	)

	// Act
	got, gotErr := NewEthernetConfigurator(nc, iface)

	// Assert
	assert.NoError(t, gotErr, "unexpected error")
	assert.NotNil(t, got, "expected non-nil EthernetConfigurator")
	assert.Equal(t, nc, got.nc, "NetworkConfigurator does not match")
	assert.Equal(t, iface, got.iface, "Interface does not match")
	assert.NotNil(t, got.device, "expected non-nil device")
}

func TestNewEthernetConfigurator_GetDevice_Error(t *testing.T) {
	// Arrange
	nc := new(NetworkConfigurator)
	iface := &v1.Interface{
		Label:      "eth0",
		MacAddress: "00:0A:95:9D:68:16",
		DHCP:       "enabled",
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "getDeviceBy",
		func(_ *NetworkConfigurator, iface *v1.Interface) (nm.DeviceWired, error) {
			return nil, errors.New("mock error")
		},
	)

	// Act
	got, gotErr := NewEthernetConfigurator(nc, iface)

	// Assert
	assert.Error(t, gotErr, "expected error")
	assert.Nil(t, got, "expected nil EthernetConfigurator")
}
