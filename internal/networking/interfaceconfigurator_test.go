/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/gsm"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewInterfaceConfigurator_Successful_EthernetInterface(t *testing.T) {
	// Arrange
	ifaceType := v1.Interface_ETHERNET

	iface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:11:22:33:44:55",
		DHCP:             "enabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.1.10",
			NetMask: "255.255.255.0",
			Gateway: "192.168.1.1",
		},
		InterfaceType: &ifaceType,
	}

	nc := new(NetworkConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(NewEthernetConfigurator, func(gotnc *NetworkConfigurator, gotIface *v1.Interface) (*EthernetConfigurator, error) {
		assert.Equal(t, nc, gotnc, "NetworkConfigurator parameter mismatch")
		assert.Equal(t, iface, gotIface, "interface parameter mismatch")
		return &EthernetConfigurator{}, nil
	})

	// Act
	got, gotErr := NewInterfaceConfigurator(nc, iface)

	// Assert
	assert.NoError(t, gotErr, "unexpected error creating Ethernet configurator")
	assert.NotNil(t, got, "expected non-nil configurator")
	assert.IsType(t, &EthernetConfigurator{}, got, "expected EthernetConfigurator type")
}

func TestNewInterfaceConfigurator_Successful_EthernetInterface_WithoutType(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:11:22:33:44:55",
		DHCP:             "enabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.1.10",
			NetMask: "255.255.255.0",
			Gateway: "192.168.1.1",
		},
	}

	nc := new(NetworkConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(NewEthernetConfigurator, func(gotnc *NetworkConfigurator, gotIface *v1.Interface) (*EthernetConfigurator, error) {
		assert.Equal(t, nc, gotnc, "NetworkConfigurator parameter mismatch")
		assert.Equal(t, iface, gotIface, "interface parameter mismatch")
		return &EthernetConfigurator{}, nil
	})

	// Act
	got, gotErr := NewInterfaceConfigurator(nc, iface)

	// Assert
	assert.NoError(t, gotErr, "unexpected error creating Ethernet configurator")
	assert.NotNil(t, got, "expected non-nil configurator")
	assert.IsType(t, &EthernetConfigurator{}, got, "expected EthernetConfigurator type")
}

func TestNewInterfaceConfigurator_Failure_EthernetInterface(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:11:22:33:44:55",
		DHCP:             "enabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.1.10",
			NetMask: "255.255.255.0",
			Gateway: "192.168.1.1",
		},
	}
	nc := new(NetworkConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(NewEthernetConfigurator, func(gotnc *NetworkConfigurator, gotIface *v1.Interface) (*EthernetConfigurator, error) {
		assert.Equal(t, nc, gotnc, "NetworkConfigurator parameter mismatch")
		assert.Equal(t, iface, gotIface, "interface parameter mismatch")
		return nil, errors.New("failed to create EthernetConfigurator")
	})

	// Act
	got, gotErr := NewInterfaceConfigurator(nc, iface)

	// Assert
	assert.Error(t, gotErr, "expected error creating Ethernet configurator")
	assert.Nil(t, got, "expected nil configurator")
}

func TestNewInterfaceConfigurator_Successful_GSMInterface(t *testing.T) {
	// Arrange
	ifaceType := v1.Interface_GSM

	iface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
		InterfaceType: &ifaceType,
	}
	gnm, _ := nm.NewNetworkManager()
	nc := &NetworkConfigurator{
		gnm: gnm,
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(gsm.NewGSMConfigurator, func(gotgnm nm.NetworkManager, gotIface *v1.Interface) *gsm.GSMConfigurator {
		assert.Equal(t, gnm, gotgnm, "NetworkManager parameter mismatch")
		assert.Equal(t, iface, gotIface, "interface parameter mismatch")
		return &gsm.GSMConfigurator{}
	})

	// Act
	got, gotErr := NewInterfaceConfigurator(nc, iface)

	// Assert
	assert.NoError(t, gotErr, "unexpected error creating GSM configurator")
	assert.NotNil(t, got, "expected non-nil configurator")
	assert.IsType(t, &gsm.GSMConfigurator{}, got, "expected GSMConfigurator type")
}

func TestNewInterfaceConfigurator_Successful_GSMInterface_WithoutType(t *testing.T) {
	// Arrange
	iface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}
	gnm, _ := nm.NewNetworkManager()
	nc := &NetworkConfigurator{
		gnm: gnm,
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(gsm.NewGSMConfigurator, func(gotgnm nm.NetworkManager, gotIface *v1.Interface) *gsm.GSMConfigurator {
		assert.Equal(t, gnm, gotgnm, "NetworkManager parameter mismatch")
		assert.Equal(t, iface, gotIface, "interface parameter mismatch")
		return &gsm.GSMConfigurator{}
	})

	// Act
	got, gotErr := NewInterfaceConfigurator(nc, iface)

	// Assert
	assert.NoError(t, gotErr, "unexpected error creating GSM configurator")
	assert.NotNil(t, got, "expected non-nil configurator")
	assert.IsType(t, &gsm.GSMConfigurator{}, got, "expected GSMConfigurator type")
}
