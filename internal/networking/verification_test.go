package networking

import (
	"errors"
	_ "errors"
	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"strings"
	"testing"
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
		func(_ *NetworkConfigurator, _ string) nm.DeviceWired { return new(mockgnm.MockDeviceWired) })

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
		func(_ *NetworkConfigurator, _ string) nm.DeviceWired { return nil })
	defer patches.Reset()

	verifyMAC(input, result, configurator)

	assert.False(t, result.retVal, "verifyMAC should return false when the device does not exist")
	assert.Equal(t, result.builder.String(), "device does not exist: mac address 20:87:56:b5:ed:e0 \n", "verifyMAC should append an error message when the device does not exist")
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
	input := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{
				Label:      "",
				MacAddress: "20:87:56:b5:ed:e0",
			},
			{
				Label: "valid-label",
				Static: &v1.Interface_StaticConf{
					IPv4:    "192.168.0.2",
					NetMask: "255.255.255.0",
					Gateway: "192.168.0.1",
				},
				DNSConfig: &v1.Interface_Dns{
					PrimaryDNS:   "1.1.1.1",
					SecondaryDNS: "8.8.8.8",
				},
			},
		},
	}
	configurator := &NetworkConfigurator{}

	patches := gomonkey.ApplyFunc((*NetworkConfigurator).getDeviceWithMac, func(_ *NetworkConfigurator, _ string) nm.DeviceWired {
		return new(mockgnm.MockDeviceWired)
	})
	defer patches.Reset()

	valid, err := verify(input, configurator)

	assert.True(t, valid, "verify should return true when all conditions are valid")
	assert.NoError(t, err, "verify should not return an error when all conditions are valid")
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
	patches := gomonkey.ApplyFunc((*NetworkConfigurator).getDeviceWithMac,
		func(_ *NetworkConfigurator, _ string) nm.DeviceWired {
			return nil
		},
	)
	defer patches.Reset()

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
	patches := gomonkey.ApplyFunc((*NetworkConfigurator).getDeviceWithMac,
		func(_ *NetworkConfigurator, _ string) nm.DeviceWired {
			return new(mockgnm.MockDeviceWired)
		})
	defer patches.Reset()

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
	patches := gomonkey.ApplyFunc((*NetworkConfigurator).getDeviceWithMac,
		func(_ *NetworkConfigurator, _ string) nm.DeviceWired {
			return new(mockgnm.MockDeviceWired)
		})
	defer patches.Reset()

	valid, err := verify(input, configurator)

	assert.False(t, valid, "verify should return false when DNSConfig is invalid")
	assert.Error(t, err, "verify should return an error when DNS")
}
