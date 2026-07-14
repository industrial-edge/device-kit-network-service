/*
 * Copyright © Siemens 2024 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	_ "errors"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"reflect"
	"strings"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func createMockVerifyResult(retVal bool) *verifyResult {
	return &verifyResult{
		retVal:  retVal,
		builder: strings.Builder{},
	}
}

// createMockCustomInterface sets a specific field in the input v1.Interface structure based on the provided key and value.
func createMockCustomInterface(input *v1.Interface, key string, value string) (*v1.Interface, error) {
	input.Static = &v1.Interface_StaticConf{}
	input.DNSConfig = &v1.Interface_Dns{}
	switch key {
	case "MacAddress":
		input.MacAddress = value
	case "IPv4":
		input.Static.IPv4 = value
	case "Gateway":
		input.Static.Gateway = value
	case "NetMask":
		input.Static.NetMask = value
	case "PrimaryDNS":
		input.DNSConfig.PrimaryDNS = value
	case "SecondaryDNS":
		input.DNSConfig.SecondaryDNS = value
	default:
		return nil, errors.New("unknown key")
	}
	return input, nil
}

func TestVerifyMAC(t *testing.T) {
	configurator := &NetworkConfigurator{}
	input, _ := createMockCustomInterface(&v1.Interface{}, "MacAddress", "20:87:56:b5:ed:e0")
	result := createMockVerifyResult(true)

	patches := gomonkey.ApplyFunc((*NetworkConfigurator).getDeviceWithMac,
		func(_ *NetworkConfigurator, _ string) nm.Device { return new(mockgnm.MockDevice) })

	defer patches.Reset()

	verifyMAC(input, result, configurator)

	assert.True(t, result.retVal, "verifyMAC should return true for a valid MAC address")
	assert.Empty(t, result.builder.String(), "verifyMAC should not append error message for a valid MAC address")
}

func TestVerifyMAC_ParseMac_Error(t *testing.T) {
	input, _ := createMockCustomInterface(&v1.Interface{}, "MacAddress", "invalid-mac-address")
	result := createMockVerifyResult(true)

	verifyMAC(input, result, &NetworkConfigurator{})

	assert.False(t, result.retVal, "verifyMAC should return false for an invalid MAC address")
	assert.Equal(t, result.builder.String(), "wrong mac address invalid-mac-address \n", "verifyMAC should append an error message for an invalid MAC address")
}

func TestVerifyMAC_DeviceDoesNotExist(t *testing.T) {
	configurator := &NetworkConfigurator{}
	input, _ := createMockCustomInterface(&v1.Interface{}, "MacAddress", "20:87:56:b5:ed:e0")
	result := createMockVerifyResult(true)

	patches := gomonkey.ApplyFunc((*NetworkConfigurator).getDeviceWithMac,
		func(_ *NetworkConfigurator, _ string) nm.Device { return nil })
	defer patches.Reset()

	verifyMAC(input, result, configurator)

	assert.False(t, result.retVal, "verifyMAC should return false when the device does not exist")
	assert.Equal(t, result.builder.String(), "no network interface found for the specified mac address: mac address 20:87:56:b5:ed:e0 \n", "verifyMAC should append an error message when the device does not exist")
}

func TestVerifyStaticConf_IPv4Invalid(t *testing.T) {
	input, _ := createMockCustomInterface(&v1.Interface{}, "IPv4", "invalid-ip-address")
	result := createMockVerifyResult(true)

	verifyStaticConf(input, result)

	assert.False(t, result.retVal, "verifyStaticConf should return false when IPv4 address is invalid")
	assert.Equal(t, result.builder.String(), "wrong ip address invalid-ip-address \n", "verifyStaticConf should append an error message for invalid IPv4 address")
}

func TestVerifyStaticConf_GatewayInvalid(t *testing.T) {
	input, _ := createMockCustomInterface(&v1.Interface{}, "Gateway", "invalid-ip-address")
	result := createMockVerifyResult(true)

	verifyStaticConf(input, result)

	assert.False(t, result.retVal, "verifyStaticConf should return false when Gateway address is invalid")
	assert.Equal(t, result.builder.String(), "wrong gateway address invalid-ip-address \n", "verifyStaticConf should append an error message for invalid Gateway address")
}

func TestVerifyStaticConf_NetMaskInvalid(t *testing.T) {
	input, _ := createMockCustomInterface(&v1.Interface{}, "NetMask", "invalid-ip-address")
	result := createMockVerifyResult(true)

	verifyStaticConf(input, result)

	assert.False(t, result.retVal, "verifyStaticConf should return false when NetMask address is invalid")
	assert.Contains(t, result.builder.String(), "wrong netmask address invalid-ip-address \n", "verifyStaticConf should append an error message for invalid NetMask address")
}

func TestVerifyDNS_PrimaryDNSInvalid(t *testing.T) {
	input, _ := createMockCustomInterface(&v1.Interface{}, "PrimaryDNS", "invalid-ip-address")
	result := createMockVerifyResult(true)

	verifyDNS(input, result)

	assert.False(t, result.retVal, "verifyDNS should return false when PrimaryDNS address is invalid")
	assert.Equal(t, result.builder.String(), "wrong dns address invalid-ip-address \n", "verifyDNS should append an error message for invalid PrimaryDNS address")
}

func TestVerifyDNS_SecondaryDNSInvalid(t *testing.T) {
	input, _ := createMockCustomInterface(&v1.Interface{}, "SecondaryDNS", "invalid-ip-address")
	result := createMockVerifyResult(true)

	verifyDNS(input, result)

	assert.False(t, result.retVal, "verifyDNS should return false when SecondaryDNS address is invalid")
	assert.Equal(t, result.builder.String(), "wrong dns address invalid-ip-address \n", "verifyDNS should append an error message for invalid SecondaryDNS address")
}

func TestVerify_AllConditionsValid(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				InterfaceType: v1.Interface_ETHERNET.Enum(),
				MacAddress:    "20:87:56:b5:ed:e0",
				DHCP:          "enabled",
				Static: &v1.Interface_StaticConf{
					IPv4:    "192.168.0.1",
					NetMask: "255.255.255.0",
					Gateway: "192.168.0.1",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.1.1.1",
					SecondaryDNS: "8.8.8.8",
				},
			},
			{
				GatewayInterface: true,
				InterfaceType:    v1.Interface_GSM.Enum(),
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "user",
					Password: "pass",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	mockDeviceWired := new(gonetworkmanager.MockDeviceWired)

	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	configurator := &NetworkConfigurator{
		gnm: mockNetworkManager,
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "getDeviceWithMac", func(_ string) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})
	defer patches.Reset()

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.True(t, valid, "verify should return true when all conditions are valid")
	assert.NoError(t, err, "verify should not return an error when all conditions are valid")
}

func TestVerify_MultipleErrors(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				MacAddress:    "mac",
				DHCP:          "enabled",
				Static: &v1.Interface_StaticConf{
					IPv4:    "192.168.0.256",
					NetMask: "255.255.255.0",
					Gateway: "192.168.0.1",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.1.1.1",
					SecondaryDNS: "8.8.8.8",
				},
			},
			{
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "user",
					Password: "pass",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	mockDeviceWired := new(gonetworkmanager.MockDeviceWired)

	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	configurator := &NetworkConfigurator{
		gnm: mockNetworkManager,
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "getDeviceWithMac", func(_ string) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})
	defer patches.Reset()

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.False(t, valid, "verify should return false when multiple errors are present")
	assert.Error(t, err, "verify should return an error when multiple errors are present")

	expectedErrorMessages := []string{
		"wrong ip address 192.168.0.256",
		"wrong mac address mac",
	}
	for _, msg := range expectedErrorMessages {
		assert.Contains(t, err.Error(), msg)
	}
}

func TestVerify_GSMConfigValidWithNoInterfaceType(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				MacAddress:    "mac",
				DHCP:          "enabled",

				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "",
					Password: "",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.True(t, valid, "verify should return true when APN is provided in GSM configuration")
	assert.Nil(t, err, "verify should not return an error when APN is provided in GSM configuration")

}

func TestVerify_GSMConfigValidWithEthernetInterfaceType(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				MacAddress:    "mac",
				DHCP:          "enabled",
				InterfaceType: v1.Interface_ETHERNET.Enum(),
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "",
					Password: "",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.True(t, valid, "verify should return true when APN is provided in GSM configuration")
	assert.Nil(t, err, "verify should not return an error when APN is provided in GSM configuration")

}

func TestVerify_GSMConfigEmptyWithGSMInterfaceType(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				MacAddress:    "mac",
				DHCP:          "enabled",
				InterfaceType: v1.Interface_GSM.Enum(),
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "",
					Pin:      "",
					Username: "",
					Password: "",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.False(t, valid, "verify should return false when APN is empty in GSM configuration")
	expectedErrorMessage := "Minimum GSM configuration is empty - APN is missing"
	assert.Contains(t, err.Error(), expectedErrorMessage)

}

func TestVerify_GSMConfigValidWithGSMInterfaceType(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				MacAddress:    "mac",
				DHCP:          "enabled",
				InterfaceType: v1.Interface_GSM.Enum(),
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "",
					Password: "",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.True(t, valid, "verify should return true when APN is provided in GSM configuration")
	assert.Nil(t, err, "verify should not return an error when APN is provided in GSM configuration")

}

func TestVerify_GSMConfigEmptyWithEthernet(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName: "eno1",
				MacAddress:    "mac",
				DHCP:          "enabled",
				InterfaceType: v1.Interface_ETHERNET.Enum(),
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "",
					Pin:      "",
					Username: "",
					Password: "",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.False(t, valid, "verify should return false when APN is empty in GSM configuration")
	expectedErrorMessage := "Minimum GSM configuration is empty - APN is missing"
	assert.Contains(t, err.Error(), expectedErrorMessage)
}

func TestVerify_MultipleDefaultGateways(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				InterfaceName:    "eno1",
				MacAddress:       "20:87:56:b5:ed:e0",
				DHCP:             "enabled",
				GatewayInterface: true,
				Static: &v1.Interface_StaticConf{
					IPv4:    "192.168.0.1",
					NetMask: "255.255.255.0",
					Gateway: "192.168.0.1",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.1.1.1",
					SecondaryDNS: "8.8.8.8",
				},
			},
			{
				Label:            "gsm0",
				DHCP:             "enabled",
				GatewayInterface: true,
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "user",
					Password: "pass",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	mockDeviceWired := new(gonetworkmanager.MockDeviceWired)

	mockNetworkManager := new(gonetworkmanager.MockNetworkManager)

	configurator := &NetworkConfigurator{
		gnm: mockNetworkManager,
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "getDeviceWithMac", func(_ string) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})
	defer patches.Reset()

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.False(t, valid, "verify should return false when multiple default gateways are set")
	assert.Error(t, err, "verify should return an error when multiple default gateways are set")

	wantErrorMessage := "more than one default gateway interface is not allowed"
	assert.Contains(t, err.Error(), wantErrorMessage)
}

func TestVerify_MacAddressInvalid(t *testing.T) {
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				Label:      "",
				MacAddress: "invalid-mac",
			},
		},
	}

	configurator := &NetworkConfigurator{}

	valid, err := verify(input, configurator)

	assert.False(t, valid, "verify should return false when MAC address is invalid")
	assert.Error(t, err, "verify should return an error when MAC address is invalid")
}

func TestVerify_StaticConfInvalid(t *testing.T) {
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				Label: "valid-label",
				Static: &v1.Interface_StaticConf{
					IPv4: "invalid-ip-address",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	valid, err := verify(input, configurator)

	assert.False(t, valid, "verify should return false when StaticConf is invalid")
	assert.Error(t, err, "verify should return an error when StaticConf is invalid")
}

func TestVerify_DNSInvalid(t *testing.T) {
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				Label: "valid-label",
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS: "invalid-ip-address",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	valid, err := verify(input, configurator)

	assert.False(t, valid, "verify should return false when DNSConfig is invalid")
	assert.Error(t, err, "verify should return an error when DNS")
}

func TestVerify_ValidGSMConfigOnly(t *testing.T) {
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				Label:            "gsm0",
				DHCP:             "enabled",
				GatewayInterface: true,
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "internet",
					Pin:      "1234",
					Username: "user",
					Password: "pass",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.0.0.1",
					SecondaryDNS: "8.8.4.4",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	valid, err := verify(input, configurator)

	assert.True(t, valid, "verify should return true when only GSMConfig is set")
	assert.NoError(t, err, "verify should not return an error when only GSMConfig is set")
}

func TestVerifyRoutes_InvalidDestination(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "invalid-ip", Netmask: "255.255.255.0", NextHop: "192.168.0.1", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route destination address invalid-ip")
}

func TestVerifyRoutes_EmptyDestination(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "", Netmask: "255.255.255.0", NextHop: "192.168.0.1", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route destination address")
}

func TestVerifyRoutes_InvalidNetmask(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "invalid-netmask", NextHop: "192.168.0.1", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route netmask address invalid-netmask")
}

func TestVerifyRoutes_EmptyNetmask(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "", NextHop: "192.168.0.1", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route netmask address")
}

func TestVerifyRoutes_InvalidNextHop(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "255.255.255.0", NextHop: "invalid-nexthop", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route next-hop address invalid-nexthop")
}

func TestVerifyRoutes_EmptyNextHop(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "255.255.255.0", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.True(t, result.retVal)
	assert.Empty(t, result.builder.String())
}

func TestVerifyRoutes_MissingMetric(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "255.255.255.0", NextHop: "192.168.0.2"},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route metric 0")
}

func TestVerifyRoutes_InvalidMetric(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "255.255.255.0", NextHop: "192.168.0.2", Metric: 1},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.False(t, result.retVal)
	assert.Contains(t, result.builder.String(), "wrong route metric 1")
}

func TestVerifyRoutes_ValidRoute(t *testing.T) {
	input := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{Destination: "192.168.0.1", Netmask: "255.255.255.0", NextHop: "192.168.0.2", Metric: 10},
		},
	}
	result := createMockVerifyResult(true)
	verifyRoutes(input, result)
	assert.True(t, result.retVal)
	assert.Empty(t, result.builder.String())
}

func Test_verifyGSM_NonDefaultGateway(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "1.0.0.1",
			SecondaryDNS: "8.8.4.4",
		},
	}
	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}
	result := createMockVerifyResult(true)
	configurator := &NetworkConfigurator{}

	// Act:
	verify(newSetting, configurator)

	// Assert:
	assert.True(t, result.retVal)
}

func Test_verifyGSM_InvalidPrimaryDNS(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "invalid-dns",
			SecondaryDNS: "8.8.4.4",
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)
	// Assert:
	assert.False(t, validConfig)
	assert.Contains(t, err.Error(), "wrong dns address invalid-dns")
}

func Test_verifyGSM_InvalidSecondaryDNS(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "1.0.0.1",
			SecondaryDNS: "invalid-dns",
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.False(t, validConfig)
	assert.Contains(t, err.Error(), "wrong dns address invalid-dns")
}

func Test_verifyGSM_EmptyDNSConfig(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		Label:            "gsm0",
		DHCP:             "enabled",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.True(t, validConfig, "verifyGSM should pass when DNSConfig is empty for GSM interface")
	assert.Nil(t, err, " DNS config is optional for GSM interface")
}

func Test_verifyGSM_ValidConfiguration(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "8.8.8.8",
			SecondaryDNS: "8.8.4.4",
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.True(t, validConfig, "verifyGSM should pass with valid configuration")
	assert.Nil(t, err, "Should have no error messages")
}

func Test_verifyGSM_ValidConfigurationNoDNS(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: nil,
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}
	result := createMockVerifyResult(true)
	configurator := &NetworkConfigurator{}

	// Act:
	verify(newSetting, configurator)

	// Assert:
	assert.True(t, result.retVal, "verifyGSM should pass when DNS config is nil")
	assert.Empty(t, result.builder.String(), "Should have no error messages")
}

func Test_verifyGSM_MultipleErrors(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: false, // Not set as default gateway
		InterfaceType:    v1.Interface_GSM.Enum(),
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},

		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "invalid-primary",
			SecondaryDNS: "invalid-secondary",
		},
	}
	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.False(t, validConfig, "verifyGSM should fail with multiple errors")
	assert.Contains(t, err.Error(), "wrong dns address invalid-primary")
	assert.Contains(t, err.Error(), "wrong dns address invalid-secondary")

}

func Test_verifyGSM_OnlyPrimaryDNS(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "8.8.8.8",
			SecondaryDNS: "", // Empty secondary DNS
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.True(t, validConfig, "verifyGSM should pass with only primary DNS")
	assert.Nil(t, err, "Should have no error messages")
}

func Test_verifyGSM_OnlySecondaryDNS(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "", // Empty primary DNS
			SecondaryDNS: "8.8.4.4",
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.True(t, validConfig, "verifyGSM should pass with only secondary DNS")
	assert.Nil(t, err, "Should have no error messages")
}

func Test_verifyGSM_EmptyDNSAddresses(t *testing.T) {
	// Arrange:
	element := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "user",
			Password: "pass",
		},
		DNSConfig: &v1.Interface_Dns{
			PrimaryDNS:   "",
			SecondaryDNS: "",
		},
	}

	newSetting := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			element,
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	validConfig, err := verify(newSetting, configurator)

	// Assert:
	assert.True(t, validConfig, "verifyGSM should pass with empty DNS addresses")
	assert.Nil(t, err, "Should have no error messages when DNS addresses are empty")
}

func TestVerify_MixedInterfaceTypesWithErrors(t *testing.T) {
	// Arrange:
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				// Ethernet interface with invalid MAC
				Label:      "test-ethernet",
				MacAddress: "invalid-mac",
				Static: &v1.Interface_StaticConf{
					IPv4:    "192.168.1.100",
					NetMask: "255.255.255.0",
					Gateway: "192.168.1.1",
				},
			},
			{
				// GSM interface with empty APN
				GsmConfiguration: &v1.Interface_GsmConf{
					Apn:      "", // Invalid: empty APN
					Pin:      "1234",
					Username: "user",
					Password: "pass",
				},
			},
		},
	}

	configurator := &NetworkConfigurator{}

	// Act:
	valid, err := verify(input, configurator)

	// Assert:
	assert.False(t, valid, "verify should return false with mixed interface errors")
	assert.Error(t, err, "verify should return an error with mixed interface errors")

	assert.Contains(t, err.Error(), "Minimum GSM configuration is empty - APN is missing")
}
