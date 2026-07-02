/*
 * Copyright © Siemens 2024 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"bytes"
	"errors"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"os"
	"reflect"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func getMockInterfaceStaticConf() *v1.Interface_StaticConf {
	return &v1.Interface_StaticConf{
		IPv4:    "192.168.1.1",
		NetMask: "255.255.255.0",
		Gateway: "192.168.1.254",
	}
}

func getMockInterfaceL2Config() *v1.Interface_L2 {
	return &v1.Interface_L2{
		NetMask:             "255.255.255.0",
		StartingAddressIPv4: "192.168.1.1",
		Range:               "256",
		Gateway:             "192.168.1.254",
		AuxiliaryAddresses:  map[string]string{"AAAA": "BBBB"},
	}
}

func getMockInterfaceDNSConfig() *v1.Interface_Dns {
	return &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "8.8.4.4",
	}
}

func getMockIP4NsData() []nm.IP4NameserverData {
	return []nm.IP4NameserverData{
		{Address: "8.8.8.8"},
		{Address: "8.8.4.4"},
		{Address: ""},
	}
}

func getMockRouteData() []map[string]interface{} {
	return []map[string]interface{}{
		{
			common.DestinationKey: "10.0.0.0",
			common.PrefixKey:      uint32(18),
			common.NextHopKey:     "1.2.3.5",
			common.MetricKey:      uint32(5),
		},
	}
}

func Test_ParseStaticIPConfig_ReturnsCorrectConfig(t *testing.T) {
	expected := getMockInterfaceStaticConf()
	connection := nm.ConnectionSettings{
		common.IPV4Key: map[string]interface{}{
			common.AddressDataKey: []map[string]interface{}{
				{
					common.AddressKey: "192.168.1.1",
					common.PrefixKey:  uint32(24),
				},
			},
			common.GatewayKey: "192.168.1.254",
		},
	}

	config := parseStaticIPConfig(connection)

	assert.Equal(t, expected, config, "Parsed config should match the expected config")
}

func Test_ParseDHCPIPv4Config_ReturnsCorrectConfig(t *testing.T) {
	expected := getMockInterfaceStaticConf()
	mockIP4Config := new(mockgnm.MockIP4Config)

	mockIP4Config.On("GetPropertyAddressData").Return(
		[]nm.IP4AddressData{{Address: "192.168.1.1", Prefix: 24}}, nil)
	mockIP4Config.On("GetPropertyGateway").Return("192.168.1.254", nil)

	config := parseDHCPIPv4Config(mockIP4Config)

	assert.Equal(t, expected, config, "Parsed config should match the expected config")
}

func Test_ParseDns_ReturnsCorrectDnsConfig(t *testing.T) {
	dnsArray := getMockIP4NsData()

	expected := &v1.Interface_Dns{
		PrimaryDNS:   "8.8.8.8",
		SecondaryDNS: "8.8.4.4",
	}

	dns := parseDns(dnsArray)

	assert.Equal(t, expected, dns, "Parsed DNS should match the expected DNS")
}

func Test_ParseRoutes_ReturnsCorrectRouteConfig(t *testing.T) {
	routeArray := getMockRouteData()

	expected := []*v1.Interface_Route{
		{
			Destination: "10.0.0.0",
			Netmask:     net.IP(net.CIDRMask(int(routeArray[0][common.PrefixKey].(uint32)), 32)).String(),
			NextHop:     "1.2.3.5",
			Metric:      5,
		},
	}

	route := parseRoutes(routeArray)
	assert.Equal(t, expected, route, "Parsed Route should match the expected Route")
}

func Test_ParseRoutes_NilValuesForKeys(t *testing.T) {
	nilRouteArray := []map[string]any{
		{
			common.DestinationKey: nil,
			common.PrefixKey:      nil,
			common.NextHopKey:     nil,
			common.MetricKey:      nil,
		},
	}
	nilExpected := []*v1.Interface_Route{
		{
			Destination: "",
			Netmask:     net.IP(net.CIDRMask(0, 32)).String(),
			NextHop:     "",
			Metric:      0,
		},
	}
	nilRoute := parseRoutes(nilRouteArray)
	assert.Equal(t, nilExpected, nilRoute, "Parsed Route with nil values should match the expected Route v4")
}

func Test_ParseRoutes_InvalidPrefixDstCombinations(t *testing.T) {
	invalidRouteArray := []map[string]any{
		{
			common.DestinationKey: "10.0.0.0",
			common.PrefixKey:      uint32(40), // Invalid for IPv4
			common.NextHopKey:     "1.2.3.5",
			common.MetricKey:      uint32(5),
		},
	}
	invalidExpected := []*v1.Interface_Route{
		{
			Destination: "10.0.0.0",
			Netmask:     net.IP(net.CIDRMask(40, 32)).String(), // Will log error, but still returns
			NextHop:     "1.2.3.5",
			Metric:      5,
		},
	}
	invalidRoute := parseRoutes(invalidRouteArray)
	assert.Equal(t, invalidExpected, invalidRoute, "Parsed Route with invalid prefix/dst should match the expected Route v4")
}

func Test_ParseRoutes_TypeCastErrors(t *testing.T) {
	typeCastErrorArray := []map[string]any{
		{
			common.DestinationKey: 12345,          // Not a string
			common.PrefixKey:      "not-a-uint32", // Not a uint32
			common.NextHopKey:     67890,          // Not a string
			common.MetricKey:      "not-a-uint32", // Not a uint32
		},
	}
	typeCastExpected := []*v1.Interface_Route{
		{
			Destination: "",
			Netmask:     net.IP(net.CIDRMask(0, 32)).String(),
			NextHop:     "",
			Metric:      0,
		},
	}
	typeCastRoute := parseRoutes(typeCastErrorArray)
	assert.Equal(t, typeCastExpected, typeCastRoute, "Parsed Route with type cast errors should match the expected Route v4")
}

func Test_ListConnections_ReturnsCorrectConnections(t *testing.T) {
	mockDevice := new(mockgnm.MockDeviceWired)
	mockConnection1 := new(mockgnm.MockConnection)
	mockConnection2 := new(mockgnm.MockConnection)
	mockSettings := new(mockgnm.MockSettings)
	mockActiveConnection := &mockgnm.MockActiveConnection{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	mockDevice.On("GetPropertyInterface").Return("eth0", nil)
	mockDevice.On("GetPropertyAvailableConnections").Return([]nm.Connection{mockConnection1, mockConnection2}, nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConnection, nil)
	mockActiveConnection.On("GetPropertyUUID").Return("f8bdcc0b-e999-44f4-9643-a5034edce2c4", nil)
	mockActiveConnection.On("GetPropertyID").Return("Wired connection 1", nil)
	mockSettings.On("ListConnections").Return([]nm.Connection{mockConnection1, mockConnection2}, nil)
	mockConnection1.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.InterfaceNameKey: "eth0",
			common.TypeKey:          common.EthernetType,
		},
	}, nil)

	mockConnection2.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.InterfaceNameKey: "eth0",
			common.TypeKey:          common.EthernetType,
		},
	}, nil)

	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	result := listConnections(mockDevice)

	assert.Equal(t, 2, len(result), "ListConnections should return a list with two connections")
	assert.Equal(t, mockConnection1, result[0], "The first connection in the result should match the mock connection")
	assert.Equal(t, mockConnection2, result[1], "The second connection in the result should match the mock connection")
}

func TestIsValidConnection_TypeNotEthernet(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.TypeKey: "wifi",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.False(t, result, "Expected false when connection type is not Ethernet")
}

func TestIsValidConnection_InterfaceNameMatches(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.TypeKey:          common.EthernetType,
			common.InterfaceNameKey: "eth0",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.True(t, result, "Expected true when interface name matches")
}

func TestIsValidConnection_InterfaceNameDoesNotMatch(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.TypeKey:          common.EthernetType,
			common.InterfaceNameKey: "eth1",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.False(t, result, "Expected false when interface name does not match")
}

func TestIsValidConnection_UUIDAndIDMatch(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.TypeKey: common.EthernetType,
			common.UUIDKey: "uuid",
			common.IDKey:   "id",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.True(t, result, "Expected true when UUID and ID match")
}

func TestIsValidConnection_UUIDAndIDDoNotMatch(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			common.TypeKey: common.EthernetType,
			common.UUIDKey: "different-uuid",
			common.IDKey:   "different-id",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.False(t, result, "Expected false when UUID and ID do not match")
}

func Test_DBusToProto_ReturnsNilWhenDeviceWiredIsNil(t *testing.T) {
	// Call the function
	result := DBusToProto(nil)
	// Assertions
	assert.Nil(t, result, "DBusToProto should return nil when the input is nil")
}

func Test_DBusToProto_ReturnsBasicInterfaceInformationWhenNoConnectionFound(t *testing.T) {
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockActiveConnection := &mockgnm.MockActiveConnection{}
	testMac := "00:0A:95:9D:68:16"
	testInterface := "eth0"
	expectedLabel := "testLabel"
	expectedL2Conf := getMockInterfaceL2Config()

	//Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(toDeviceWired, func(objectPath dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})

	mockDeviceWired.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockDeviceWired.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))

	mockDeviceWired.On("GetPropertyActiveConnection").Return(mockActiveConnection, errors.New("error from GetPropertyActiveConnection"))
	mockDeviceWired.On("GetPropertyHwAddress").Return(testMac, nil)
	mockDeviceWired.On("GetPropertyInterface").Return(testInterface, nil)

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{}
	})
	patches.ApplyFunc(dockerNetworkGetMacvlanConnection, func(_ string) *v1.Interface_L2 {
		return expectedL2Conf
	})
	patches.ApplyFunc(getLabelForInterface, func(interfaceName string) (string, error) {
		return expectedLabel, nil
	})
	// Call the function
	result := DBusToProto(mockDeviceWired)
	log.Printf("Result %v\n", result)

	assert.Equal(t, testInterface, result.InterfaceName, "DBusToProto should return an Interface with the correct interface name")
	assert.Equal(t, testMac, result.MacAddress, "DBusToProto should return an Interface with the correct MAC address")
	assert.Equal(t, expectedL2Conf, result.L2Conf, "DBusToProto should return an Interface with the correct L2 config")
	assert.Equal(t, expectedLabel, result.Label, "DBusToProto should return an Interface with the correct label")
}

func Test_DBusToProto_ReturnsInterfaceWithFirstConnectionWhenActiveConnectionNotAvailable(t *testing.T) {
	mockDevice := &mockgnm.MockDevice{}
	mockDeviceWired := mockgnm.MockDeviceWired{}
	mockConnection1 := &mockgnm.MockConnection{}
	mockConnection2 := &mockgnm.MockConnection{}
	mockActiveConnection := &mockgnm.MockActiveConnection{}
	testInterface := "eth0"
	expectedInterface := &v1.Interface{
		MacAddress:       "00:0A:95:9D:68:16",
		Label:            "testLabel",
		DHCP:             common.Enabled,
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        nil,
		GatewayInterface: false,
	}
	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	// Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(toDeviceWired, func(objectPath dbus.ObjectPath) (nm.DeviceWired, error) {
		log.Println("newWired patched function called")
		return &mockDeviceWired, nil
	})

	mockDevice.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))
	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConnection, errors.New("error from GetPropertyActiveConnection"))
	mockDeviceWired.On("GetPropertyHwAddress").Return(expectedInterface.MacAddress, nil)
	mockDevice.On("GetPropertyInterface").Return(testInterface, nil)
	mockConnection1.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			"id": "connection1",
		},
	}, nil)
	mockConnection2.On("GetSettings").Return(nm.ConnectionSettings{
		common.ConnectionKey: map[string]interface{}{
			"id": "connection2",
		},
	}, nil)

	// Patch listConnections function to return a non-empty list
	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{mockConnection1, mockConnection2}
	})

	// Patch convertToProto function
	patches.ApplyFunc(convertToProto, func(settings nm.ConnectionSettings, ipv4conf nm.IP4Config, mac string) *v1.Interface {
		return expectedInterface
	})

	patches.ApplyFunc(getLabelForInterface, func(_ string) (string, error) {
		return expectedInterface.Label, nil
	})

	// Call the function
	result := DBusToProto(mockDevice)
	log.Printf("Result %v\n", result)

	// Assertions
	assert.NotNil(t, result, "DBusToProto should return a non-nil Interface instance")
	assert.Equal(t, expectedInterface.Label, result.Label, "DBusToProto should return an Interface with the correct label")
	assert.Equal(t, expectedInterface.MacAddress, result.MacAddress, "DBusToProto should return an Interface with the correct MAC address")
	assert.Equal(t, testInterface, result.InterfaceName, "DBusToProto should return an Interface with the correct interface name")
	assert.Equal(t, expectedInterface.DHCP, result.DHCP, "DBusToProto should return an Interface with the correct DHCP setting")
	assert.Equal(t, expectedInterface.Static, result.Static, "DBusToProto should return an Interface with the correct static IP configuration")
	assert.Equal(t, expectedInterface.DNSConfig, result.DNSConfig, "DBusToProto should return an Interface with the correct DNS configuration")
	assert.Equal(t, expectedInterface.GatewayInterface, result.GatewayInterface, "DBusToProto should return an Interface with the correct gateway interface setting")
}

func Test_DBusToProto_ReturnsFullInterfaceInfoWhenActiveConnectionIsAvailable(t *testing.T) {
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockConnection := &mockgnm.MockConnection{}
	mockActiveConnection := &mockgnm.MockActiveConnection{}
	mockIP4Config := &mockgnm.MockIP4Config{}
	expectedInterface := &v1.Interface{
		MacAddress:       "00:0A:95:9D:68:16",
		Label:            "testLabel",
		DHCP:             common.Enabled,
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        getMockInterfaceDNSConfig(),
		GatewayInterface: true,
	}
	testInterface := "eth0"

	// Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(toDeviceWired, func(objectPath dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})

	mockDeviceWired.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockDeviceWired.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))
	mockDeviceWired.On("GetPropertyActiveConnection").Return(mockActiveConnection, nil)
	mockDeviceWired.On("GetPropertyHwAddress").Return(expectedInterface.MacAddress, nil)
	mockDeviceWired.On("GetPropertyInterface").Return(testInterface, nil)
	mockActiveConnection.On("GetPropertyIP4Config").Return(mockIP4Config, nil)
	mockActiveConnection.On("GetPropertyConnection").Return(mockConnection, nil)
	mockConnection.On("GetSettings").Return(nm.ConnectionSettings{}, nil)

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{}
	})

	// Patch convertToProto function
	patches.ApplyFunc(convertToProto, func(settings nm.ConnectionSettings, ipv4conf nm.IP4Config, mac string) *v1.Interface {
		return expectedInterface
	})

	// Call the function
	result := DBusToProto(mockDeviceWired)
	log.Printf("Result %v\n", result)

	// Assertions
	assert.NotNil(t, result, "DBusToProto should return a non-nil Interface instance")

	assert.Equal(t, expectedInterface.Label, result.Label, "DBusToProto should return an Interface with the correct label")
	assert.Equal(t, expectedInterface.MacAddress, result.MacAddress, "DBusToProto should return an Interface with the correct MAC address")
	assert.Equal(t, testInterface, result.InterfaceName, "DBusToProto should return an Interface with the correct interface name")
	assert.Equal(t, expectedInterface.DHCP, result.DHCP, "DBusToProto should return an Interface with the correct DHCP setting")
	assert.Equal(t, expectedInterface.Static, result.Static, "DBusToProto should return an Interface with the correct static IP configuration")
	assert.Equal(t, expectedInterface.DNSConfig, result.DNSConfig, "DBusToProto should return an Interface with the correct DNS configuration")
	assert.Equal(t, expectedInterface.GatewayInterface, result.GatewayInterface, "DBusToProto should return an Interface with the correct gateway interface setting")

}

func Test_ConvertToProto_ReturnsCorrectProtoWhenDHCPEnabled(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		common.IPV4Key: map[string]interface{}{
			common.MethodKey:      common.Auto,
			common.RouteMetricKey: int64(1),
		},
	}

	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)
	mockIPV4Config.On("GetPropertyRouteData").Return([]nm.IP4RouteData{}, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(parseDHCPIPv4Config, func(ipv4conf nm.IP4Config) *v1.Interface_StaticConf {
		return getMockInterfaceStaticConf()
	})

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, common.Enabled, result.DHCP, "DHCP should be enabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
}

func Test_ConvertToProto_ReturnsCorrectProtoWhenStaticIP(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		common.IPV4Key: map[string]interface{}{
			common.AddressDataKey: []map[string]interface{}{
				{
					common.AddressKey: "192.168.1.1",
					common.PrefixKey:  uint32(24),
				},
			},
			common.GatewayKey:     "192.168.1.254",
			common.RouteMetricKey: int64(1),
		},
	}

	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)
	mockIPV4Config.On("GetPropertyRouteData").Return([]nm.IP4RouteData{}, nil)

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, common.Disabled, result.DHCP, "DHCP should be disabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
}

func Test_ConvertToProto_ReturnsCorrectProtoWithDHCPIPConfig(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		common.IPV4Key: map[string]interface{}{
			common.MethodKey: common.Auto,
			common.AddressDataKey: []map[string]interface{}{
				{
					common.AddressKey: "192.168.1.1",
					common.PrefixKey:  uint32(24),
				},
			},
			common.GatewayKey:     "192.168.1.254",
			common.RouteMetricKey: int64(1),
		},
	}

	mockIPV4Config.On("GetPropertyAddressData").Return([]nm.IP4AddressData{{Address: "192.168.1.1", Prefix: uint8(24)}}, nil)
	mockIPV4Config.On("GetPropertyGateway").Return("192.168.1.254", nil)
	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)
	mockIPV4Config.On("GetPropertyRouteData").Return([]nm.IP4RouteData{}, nil)

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, common.Enabled, result.DHCP, "DHCP should be enabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
}

func Test_ConvertToProto_ReturnsCorrectProtoWithRouteConfig(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		common.IPV4Key: map[string]interface{}{
			common.MethodKey: common.Auto,
			common.AddressDataKey: []map[string]interface{}{
				{
					common.AddressKey: "192.168.1.1",
					common.PrefixKey:  uint32(24),
				},
			},
			common.GatewayKey:     "192.168.1.254",
			common.RouteMetricKey: int64(1),
			common.RouteDataKey: []map[string]interface{}{
				{
					common.DestinationKey: "1.2.3.1",
					common.PrefixKey:      uint32(23),
					common.NextHopKey:     "1.2.3.2",
					common.MetricKey:      uint32(5),
				},
			},
		},
	}

	mockIPV4Config.On("GetPropertyAddressData").Return([]nm.IP4AddressData{{Address: "192.168.1.1", Prefix: uint8(24)}}, nil)
	mockIPV4Config.On("GetPropertyGateway").Return("192.168.1.254", nil)
	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)

	expectedRoutes := []*v1.Interface_Route{
		{
			Destination: "1.2.3.1",
			Netmask:     "255.255.254.0",
			NextHop:     "1.2.3.2",
			Metric:      5,
		},
	}

	assert.Equal(t, expectedRoutes, result.GetRoutes())
}

func Test_NewSettingsFromProto_ReturnsSettingsWhenDHCPEnabled(t *testing.T) {
	deviceName := "eth0"
	protoData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       "20:87:56:b5:ed:e0",
		DHCP:             "enabled",
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "8.4.4.4"},
		Routes: []*v1.Interface_Route{
			{
				Destination: "1.5.6.7",
				Netmask:     "255.255.254.0",
				NextHop:     "1.2.3.4",
				Metric:      2,
			},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(common.IPToUInt32LI, func(ip string) uint32 {
		return 1234567890
	})

	patches.ApplyFunc(net.ParseMAC, func(mac string) (net.HardwareAddr, error) {
		return net.HardwareAddr{}, nil
	})

	settings := newSettingsFromProto(protoData, deviceName)

	assert.NotNil(t, settings, "newSettingsFromProto should return non-nil result")

	assert.NotNil(t, settings[common.ConnectionKey][common.UUIDKey], "UUID should be set in Connection settings")

	gotUUID := settings[common.ConnectionKey][common.UUIDKey].(string)
	_, err := uuid.Parse(gotUUID)
	assert.NoError(t, err, "UUID should be a valid UUID string")

	assert.NotNil(t, settings[common.ConnectionKey][common.TimeStampKey], "TimeStamp should be set in Connection settings")

	assert.NotEmpty(t, settings[common.ConnectionKey][common.IDKey], "ID should be set in Connection settings")
	assert.Equal(t, "802-3-ethernet", settings[common.ConnectionKey][common.TypeKey])
	assert.Equal(t, "eth0", settings[common.ConnectionKey][common.InterfaceNameKey])

	assert.Equal(t, 1, settings[common.IPV4Key][common.RouteMetricKey])
	assert.Equal(t, "auto", settings[common.IPV4Key][common.MethodKey])
	assert.Equal(t, []uint32{1234567890, 1234567890}, settings[common.IPV4Key][common.DNSKey])
	assert.Equal(t, 1, settings[common.IPV4Key][common.DNSIgnoreAutoKey])
}

func Test_ConvertToProto_RawRouteDataCastFails(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		common.IPV4Key: map[string]any{
			common.MethodKey:      common.Auto,
			common.RouteMetricKey: int64(1),
			common.RouteDataKey:   "not-a-slice", // This will fail type assertion
		},
	}
	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)
	mockIPV4Config.On("GetPropertyRouteData").Return([]nm.IP4RouteData{}, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(parseDHCPIPv4Config, func(ipv4conf nm.IP4Config) *v1.Interface_StaticConf {
		return getMockInterfaceStaticConf()
	})

	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, common.Enabled, result.DHCP, "DHCP should be enabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
	assert.Nil(t, result.Routes, "Routes should be nil when rawRouteData cast fails")
	assert.Contains(t, logOutput.String(), "failed to cast routes", "Should log cast failure")
}

// func Test_DBusToProto_GSMDevice_ReturnsGsmConfigWithoutMacAndL2(t *testing.T) {
// 	mockDevice := &mockgnm.MockDevice{}
// 	mockConnection := &mockgnm.MockConnection{}
// 	mockActiveConnection := &mockgnm.MockActiveConnection{}
// 	mockIP4 := &mockgnm.MockIP4Config{}

// 	testInterface := "wwan0"

// 	expectedGsm := &v1.Interface_GsmConf{
// 		Apn:      "internet",
// 		Pin:      "****",
// 		Username: "****",
// 		Password: "****",
// 	}

// 	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil)

// 	mockDevice.On("GetPropertyInterface").Return(testInterface, nil)

// 	mockDevice.On("GetPropertyActiveConnection").Return(mockActiveConnection, nil)

// 	mockActiveConnection.On("GetPropertyIP4Config").Return(mockIP4, nil)

// 	mockActiveConnection.On("GetPropertyConnection").Return(mockConnection, nil)

// 	mockConnection.On("GetSettings").
// 		Return(nm.ConnectionSettings{
// 			constants.GSMSetting: map[string]interface{}{
// 				constants.APNKey:      "internet",
// 				constants.PINKey:      "1234",
// 				constants.UsernameKey: "user",
// 				constants.PasswordKey: "pass",
// 			},
// 		}, nil)

// 	// --- Monkey patches ---
// 	patches := gomonkey.NewPatches()
// 	defer patches.Reset()

// 	// No Ethernet → listConnections not required
// 	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
// 		return nil
// 	})

// 	// convertToProto should NOT override GSM fields
// 	patches.ApplyFunc(convertToProto, func(
// 		settings nm.ConnectionSettings,
// 		ipv4 nm.IP4Config,
// 		mac string,
// 	) *v1.Interface {
// 		return &v1.Interface{}
// 	})

// 	patches.ApplyFunc(getLabelForInterface, func(_ string) (string, error) {
// 		return "gsm-label", nil
// 	})

// 	result := DBusToProto(mockDevice)
// 	log.Println("result:", result)
// 	assert.NotNil(t, result)
// 	// GSM behavior
// 	assert.Equal(t, expectedGsm.Apn, result.GsmConfiguration.Apn)
// 	assert.Equal(t, expectedGsm.Password, result.GsmConfiguration.Password)
// 	assert.Equal(t, expectedGsm.Username, result.GsmConfiguration.Username)
// 	assert.Equal(t, expectedGsm.Pin, result.GsmConfiguration.Pin)
// 	// No MAC for GSM
// 	assert.Empty(t, result.MacAddress)
// 	// No L2 for GSM
// 	assert.Nil(t, result.L2Conf)
// 	assert.Equal(t, testInterface, result.InterfaceName)
// 	assert.Equal(t, "gsm-label", result.Label)
// }

func Test_NewSettingsFromProto_ReturnsSettingsWithDefaultsWhenPutMacAddressFails(t *testing.T) {
	deviceName := "eth0"
	protoData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       "invalid mac address",
		DHCP:             "enabled",
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "8.4.4.4"},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(common.IPToUInt32LI, func(ip string) uint32 {
		return 1234567890
	})

	patches.ApplyFunc(net.ParseMAC, func(mac string) (net.HardwareAddr, error) {
		return net.HardwareAddr{}, errors.New("error from ParseMAC")
	})

	settings := newSettingsFromProto(protoData, deviceName)

	assert.NotNil(t, settings, "newSettingsFromProto should return non-nil result")
	assert.NotNil(t, settings[common.ConnectionKey][common.UUIDKey], "UUID should be set in Connection settings")

	gotUUID := settings[common.ConnectionKey][common.UUIDKey].(string)
	_, err := uuid.Parse(gotUUID)
	assert.NoError(t, err, "UUID should be a valid UUID string")

	assert.NotNil(t, settings[common.ConnectionKey][common.TimeStampKey], "TimeStamp should be set in Connection settings")

	assert.Equal(t, "invalid mac address_dhcp", settings[common.ConnectionKey][common.IDKey])
	assert.Equal(t, "802-3-ethernet", settings[common.ConnectionKey][common.TypeKey])
	assert.Equal(t, "eth0", settings[common.ConnectionKey][common.InterfaceNameKey])

	assert.Equal(t, 1, settings[common.IPV4Key][common.RouteMetricKey])
	assert.Equal(t, "auto", settings[common.IPV4Key][common.MethodKey])
	assert.Equal(t, []uint32{1234567890, 1234567890}, settings[common.IPV4Key][common.DNSKey])
	assert.Equal(t, 1, settings[common.IPV4Key][common.DNSIgnoreAutoKey])
}

func Test_NewSettingsFromProto_ReturnsSettingsWithStaticIPWhenDHCPDisabled(t *testing.T) {
	deviceName := "eth0"
	protoData := &v1.Interface{
		GatewayInterface: true,
		Label:            "eth0",
		DHCP:             "disabled",
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "8.8.4.4"},
		Routes: []*v1.Interface_Route{
			{
				Destination: "1.5.6.7",
				Netmask:     "255.255.254.0",
				NextHop:     "1.2.3.4",
				Metric:      2,
			},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(common.IPToUInt32LI, func(ip string) uint32 {
		return 1234567890
	})

	patches.ApplyFunc(net.ParseMAC, func(mac string) (net.HardwareAddr, error) {
		return net.HardwareAddr{}, nil
	})

	patches.ApplyFunc(dbus.MakeVariantWithSignature, func(value interface{}, signature dbus.Signature) dbus.Variant {
		return dbus.Variant{}
	})

	settings := newSettingsFromProto(protoData, deviceName)

	assert.NotNil(t, settings, "newSettingsFromProto should return non-nil result")

	assert.NotNil(t, settings, "newSettingsFromProto should return non-nil result")
	assert.NotNil(t, settings[common.ConnectionKey][common.UUIDKey], "UUID should be set in Connection settings")

	gotUUID := settings[common.ConnectionKey][common.UUIDKey].(string)
	_, err := uuid.Parse(gotUUID)
	assert.NoError(t, err, "UUID should be a valid UUID string")

	assert.NotNil(t, settings[common.ConnectionKey][common.TimeStampKey], "TimeStamp should be set in Connection settings")

	assert.NotEmpty(t, settings[common.ConnectionKey][common.IDKey], "ID should be set in Connection settings")
	assert.Equal(t, "802-3-ethernet", settings[common.ConnectionKey][common.TypeKey], "unexpected Type value")
	assert.Equal(t, "eth0", settings[common.ConnectionKey][common.InterfaceNameKey], "unexpected InterfaceName value")

	assert.Equal(t, 1, settings[common.IPV4Key][common.RouteMetricKey], "unexpected RouteMetric value")
	assert.Equal(t, "manual", settings[common.IPV4Key][common.MethodKey], "unexpected Method value")

	assert.Equal(t, []uint32{1234567890, 1234567890}, settings[common.IPV4Key][common.DNSKey], "unexpected DNS value")

	assert.Len(t, settings[common.IPV4Key][common.AddressDataKey].([]DBusDict), 1, "unexpected AddressData value")

	wantGateway := "192.168.1.254"
	assert.Equal(t, wantGateway, settings[common.IPV4Key][common.GatewayKey], "unexpected Gateway value")

	assert.Len(t, settings[common.IPV4Key][common.RouteDataKey].([]DBusDict), 1, "unexpected RouteData value")
}

func Test_GetMapWithUppercase_ConvertsKeysAndValuesToUppercase(t *testing.T) {
	inputMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	expectedOutput := map[string]string{
		"KEY1": "VALUE1",
		"KEY2": "VALUE2",
	}

	result := GetMapWithUppercase(inputMap)

	assert.Equal(t, expectedOutput, result, "GetMapWithUppercase() result should match the expected output")
}

func Test_WriteMapToFile_CreatesFileWithUppercaseMap(t *testing.T) {
	expectedContent := `{"KEY1":"VALUE1","KEY2":"VALUE2"}`
	inputMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	tempFile, err := os.CreateTemp("", "testfile.json")
	assert.NoError(t, err, "CreateTemp() should not return an error")
	defer os.Remove(tempFile.Name())

	err = WriteMapToFile(inputMap, tempFile.Name())
	assert.NoError(t, err, "WriteMapToFile() should not return an error")

	content, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err, "ReadFile() should not return an error")

	assert.Equal(t, expectedContent, string(content), "File content should match the expected content")
}

func Test_ReadMapFromFile_ReturnsCorrectMap(t *testing.T) {
	fileContent := `{"KEY1":"VALUE1","KEY2":"VALUE2"}`
	expectedMap := map[string]string{
		"KEY1": "VALUE1",
		"KEY2": "VALUE2",
	}

	tempFile, err := os.CreateTemp("", "testfile.json")
	assert.NoError(t, err, "CreateTemp() should not return an error")
	defer os.Remove(tempFile.Name())

	err = os.WriteFile(tempFile.Name(), []byte(fileContent), 0666)
	assert.NoError(t, err, "WriteFile() should not return an error")

	parsedMap, err := readMapFromFile(tempFile.Name())
	assert.NoError(t, err, "WriteFile() should not return an error")

	assert.Equal(t, expectedMap, parsedMap, "Parsed map should match the expected map")
}

func Test_GetInterfaceForLabel_ReturnsErrorWhenReadMapFromFileFail(t *testing.T) {
	label := "LABEL1"
	expectedLogContents := "error from readMapFromFile"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(readMapFromFile, func(fileName string) (map[string]string, error) {
		return nil, errors.New("error from readMapFromFile")
	})

	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	_ = getInterfaceForLabel(label)
	logContents := logOutput.String()

	assert.Contains(t, logContents, expectedLogContents, "Log contents do not match")
}

func Test_GetInterfaceForLabel_ReturnsCorrectInterface(t *testing.T) {
	label := "LABEL1"
	expectedInterface := "eth0"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(readMapFromFile, func(fileName string) (map[string]string, error) {
		return map[string]string{"LABEL1": "eth0", "LABEL2": "eth1"}, nil
	})

	result := getInterfaceForLabel(label)

	assert.Equal(t, expectedInterface, result, "For label '%s', expected interface '%s'", label, expectedInterface)
}

func Test_GetLabelForInterface_ReturnsCorrectLabel(t *testing.T) {
	interfaceName := "eth0"
	expectedLabel := "LABEL1"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(readMapFromFile, func(fileName string) (map[string]string, error) {
		return map[string]string{"LABEL1": "ETH0", "LABEL2": "eth1"}, nil
	})

	label, err := getLabelForInterface(interfaceName)

	assert.NoError(t, err, "Expected no error while reading label from valid interface")
	assert.Equal(t, expectedLabel, label, "Expected label for interface '%s' to be '%s'", interfaceName, expectedLabel)
}

func Test_GetLabelForInterface_ReturnsErrorForInvalidInterface(t *testing.T) {
	interfaceName := "eth0"
	expectedError := "interface not found: eth0"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(readMapFromFile, func(fileName string) (map[string]string, error) {
		return map[string]string{"LABEL1": "ETH2", "LABEL2": "eth1"}, nil
	})

	label, err := getLabelForInterface(interfaceName)

	assert.Error(t, err, "Expected no error while reading label from valid interface")
	assert.Equal(t, expectedError, err.Error())
	assert.Equal(t, "", label)
}

func Test_GetLabelForInterface_ReturnsErrorWhenReadMapFromFileFails(t *testing.T) {
	interfaceName := "invalid"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(readMapFromFile, func(fileName string) (map[string]string, error) {
		return nil, errors.New("error from readMapFromFile")
	})

	label, err := getLabelForInterface(interfaceName)

	assert.Equal(t, "", label)
	assert.Error(t, err, "Expected error from while trying to read label from inaccessible interface")
}

func Test_ConfigureExistingGatewayInterfacesExceptProtoData_GatewayInterfaceFalse(t *testing.T) {
	protoData := &v1.Interface{GatewayInterface: false}
	networkConfigurator := NetworkConfigurator{}

	err := ConfigureExistingGatewayInterfacesExceptProtoData(protoData, networkConfigurator)
	assert.NoError(t, err, "Expected no error when GatewayInterface is false")
}

func Test_ConfigureExistingGatewayInterfacesExceptProtoData_setGatewayInterfaceForDeviceConnectionsError(t *testing.T) {
	protoData := &v1.Interface{GatewayInterface: true}
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return []nm.DeviceWired{mockDevice}
	})

	patches.ApplyFuncReturn(setGatewayInterfaceForDeviceConnections, errors.New("error from setGatewayInterfaceForDeviceConnections"))

	err := ConfigureExistingGatewayInterfacesExceptProtoData(protoData, *nc)
	assert.Error(t, err, "Expected error from setGatewayInterfaceForDeviceConnections")
	assert.Equal(t, "error from setGatewayInterfaceForDeviceConnections", err.Error())
}

func Test_ConfigureExistingGatewayInterfacesExceptProtoData_Success(t *testing.T) {
	protoData := &v1.Interface{GatewayInterface: true}
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return []nm.DeviceWired{mockDevice}
	})

	var callCount int
	patches.ApplyFunc(setGatewayInterfaceForDeviceConnections, func(device nm.DeviceWired, protoData *v1.Interface) error {
		callCount++
		assert.Equal(t, mockDevice, device, "Expected device to be mockDevice")
		assert.Equal(t, protoData, protoData, "Expected protoData to be passed correctly")
		return nil
	})

	err := ConfigureExistingGatewayInterfacesExceptProtoData(protoData, *nc)
	assert.NoError(t, err, "Expected no error when setGatewayInterfaceForDeviceConnections succeeds")
	assert.Equal(t, 1, callCount, "Expected setGatewayInterfaceForDeviceConnections to be called twice")
}

func Test_setGatewayInterfaceForDeviceConnections_checkAndUpdateGatewayInterfaceForConnectionError(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	protoData := &v1.Interface{}
	networkConfigurator := NetworkConfigurator{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{&mockgnm.MockConnection{}}
	})
	patches.ApplyFuncReturn(checkAndUpdateGatewayInterfaceForConnection, errors.New("error from checkAndUpdateGatewayInterfaceForConnection"))

	err := setGatewayInterfaceForDeviceConnections(ethernetDevice, protoData, networkConfigurator)
	assert.Error(t, err, "Expected error from checkAndUpdateGatewayInterfaceForConnection")
}

func Test_setGatewayInterfaceForDeviceConnections_Success(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	protoData := &v1.Interface{}
	networkConfigurator := NetworkConfigurator{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	var callCount int
	// Mock listConnections to return a list of mock connections
	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		callCount++
		return []nm.Connection{&mockgnm.MockConnection{}}
	})

	// Mock checkAndUpdateGatewayInterfaceForConnection and deactivateAndActivateConnection to return nil
	patches.ApplyFuncReturn(checkAndUpdateGatewayInterfaceForConnection, nil)
	patches.ApplyFuncReturn(deactivateAndActivateConnection, nil)

	// Call the function under test
	err := setGatewayInterfaceForDeviceConnections(ethernetDevice, protoData, networkConfigurator)

	// Assertions
	assert.NoError(t, err, "Expected no error when all connections are processed successfully")
	assert.Equal(t, 1, callCount, "Expected listConnection to be called once")
}

func Test_deactivateAndActivateConnection_GetPropertyActiveConnectionError(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	connection := &mockgnm.MockConnection{}
	networkConfigurator := NetworkConfigurator{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyActiveConnection", nil, errors.New("error from GetPropertyActiveConnection"))

	err := deactivateAndActivateConnection(ethernetDevice, connection, networkConfigurator)
	assert.Error(t, err, "Expected error from GetPropertyActiveConnection")
}

func Test_deactivateAndActivateConnection_GetPropertyActiveConnectionWhenReturnNilActiveConnection(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	connection := &mockgnm.MockConnection{}
	networkConfigurator := NetworkConfigurator{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyActiveConnection", nil, nil)
	ethernetDevice.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)

	err := deactivateAndActivateConnection(ethernetDevice, connection, networkConfigurator)
	assert.Nil(t, err, "Expected no error when the active connection is nil")
}

func Test_deactivateAndActivateConnection_DeactivateConnectionError(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	connection := &mockgnm.MockConnection{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyActiveConnection", &mockgnm.MockActiveConnection{}, nil)
	patches.ApplyMethodReturn(networkConfigurator.gnm, "DeactivateConnection", errors.New("error from DeactivateConnection"))

	err := deactivateAndActivateConnection(ethernetDevice, connection, networkConfigurator)
	assert.Error(t, err, "Expected error from DeactivateConnection")
}

func Test_deactivateAndActivateConnection_ActivateConnectionError(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	connection := &mockgnm.MockConnection{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyActiveConnection", &mockgnm.MockActiveConnection{}, nil)
	patches.ApplyMethodReturn(networkConfigurator.gnm, "DeactivateConnection", nil)
	patches.ApplyMethodReturn(networkConfigurator.gnm, "ActivateConnection", nil, errors.New("error from ActivateConnection"))

	err := deactivateAndActivateConnection(ethernetDevice, connection, networkConfigurator)
	assert.Error(t, err, "Expected error from ActivateConnection")
}

func Test_deactivateAndActivateConnection_Success(t *testing.T) {
	ethernetDevice := &mockgnm.MockDeviceWired{}
	connection := &mockgnm.MockConnection{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyActiveConnection", &mockgnm.MockActiveConnection{}, nil)
	patches.ApplyMethodReturn(networkConfigurator.gnm, "DeactivateConnection", nil)
	patches.ApplyMethodReturn(networkConfigurator.gnm, "ActivateConnection", nil, nil)

	err := deactivateAndActivateConnection(ethernetDevice, connection, networkConfigurator)
	assert.NoError(t, err, "Expected no error when the connection is switched successfully")
}

func Test_checkAndUpdateGatewayInterfaceForConnection_GetSettingsError(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	protoData := &v1.Interface{}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "GetSettings", nil, errors.New("error from GetSettings"))

	err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator)
	assert.Error(t, err, "Expected error from GetSettings")
}

func Test_checkAndUpdateGatewayInterfaceForConnection_SetMacAddressInSettingsError(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	protoData := &v1.Interface{}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: nil}}, nil)
	patches.ApplyFuncReturn(setMacAddressInSettings, errors.New("error from setMacAddressInSettings"))

	err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator)
	assert.Error(t, err, "Expected error from setMacAddressInSettings")
}

func Test_checkAndUpdateGatewayInterfaceForConnection_changePriorityOfGatewayInterfaceError(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	protoData := &v1.Interface{}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}, nil)
	patches.ApplyFuncReturn(willGatewayInterfaceBeUpdated, true)
	patches.ApplyFuncReturn(changePriorityOfGatewayInterface, errors.New("error from changePriorityOfGatewayInterface"))

	err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator)
	assert.Error(t, err, "Expected error from changePriorityOfGatewayInterface")
}

func Test_checkAndUpdateGatewayInterfaceForConnection_deactivateAndActivateConnectionError(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	protoData := &v1.Interface{}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}, nil)
	patches.ApplyFuncReturn(willGatewayInterfaceBeUpdated, true)
	patches.ApplyFuncReturn(changePriorityOfGatewayInterface, nil)
	patches.ApplyFuncReturn(deactivateAndActivateConnection, errors.New("error from deactivateAndActivateConnection"))

	err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator)
	assert.Error(t, err, "Expected error from deactivateAndActivateConnection")
}

func Test_checkAndUpdateGatewayInterfaceForConnection_Success(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	protoData := &v1.Interface{}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	gnm := &mockgnm.MockNetworkManager{}
	networkConfigurator := NetworkConfigurator{gnm: gnm}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}, nil)
	patches.ApplyFuncReturn(willGatewayInterfaceBeUpdated, false)

	err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator)
	assert.NoError(t, err, "Expected no error when the connection is processed successfully")
}

func Test_setMacAddressInSettings_InvalidMacAddress(t *testing.T) {
	settings := nm.ConnectionSettings{common.EthernetType: map[string]interface{}{}}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyPermHwAddress", "invalid-mac-address", nil)
	patches.ApplyFuncReturn(net.ParseMAC, nil, errors.New("error from ParseMAC"))

	err := setMacAddressInSettings(settings, ethernetDevice)
	assert.Error(t, err, "Expected error from ParseMAC")
}

func Test_setMacAddressInSettings_FallbackToHwAddress(t *testing.T) {
	settings := nm.ConnectionSettings{common.EthernetType: map[string]interface{}{}}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Make GetPropertyPermHwAddress return empty string (simulate missing permanent MAC)
	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyPermHwAddress", "", nil)

	// Make GetPropertyHwAddress return a valid MAC instead
	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyHwAddress", "00:11:22:33:44:55", nil)

	// Patch ParseMAC to return the expected parsed hardware address
	expectedMac := net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	patches.ApplyFuncReturn(net.ParseMAC, expectedMac, nil)

	// Call the function
	err := setMacAddressInSettings(settings, ethernetDevice)

	// Assertions
	assert.NoError(t, err, "Expected no error when fallback MAC address is used")
	assert.Equal(
		t,
		expectedMac,
		net.HardwareAddr(settings[common.EthernetType][common.MACAddressKey].([]uint8)),
		"Expected fallback MAC address to be set in settings",
	)
}

func Test_setMacAddressInSettings_Success(t *testing.T) {
	settings := nm.ConnectionSettings{common.EthernetType: map[string]interface{}{}}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyPermHwAddress", "00:0A:95:9D:68:16", nil)
	patches.ApplyFuncReturn(net.ParseMAC, net.HardwareAddr{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}, nil)

	err := setMacAddressInSettings(settings, ethernetDevice)
	assert.NoError(t, err, "Expected no error when the MAC address is set successfully")
	assert.Equal(t, net.HardwareAddr{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}, net.HardwareAddr(settings[common.EthernetType][common.MACAddressKey].([]uint8)), "Expected MAC address to be set in the backup")
}

func Test_willGatewayInterfaceBeUpdated_MacAddressDifferent(t *testing.T) {
	protoData := &v1.Interface{MacAddress: "00:0A:95:9D:68:16"}
	settings := nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x17}}}

	result := willGatewayInterfaceBeUpdated(protoData, "00:0A:95:9D:68:17", settings)
	assert.True(t, result, "Expected true when MAC addresses are different")
}

func Test_willGatewayInterfaceBeUpdated_LabelDifferent(t *testing.T) {
	protoData := &v1.Interface{Label: "label1"}
	settings := nm.ConnectionSettings{common.ConnectionKey: map[string]interface{}{common.InterfaceNameKey: "eth0"}}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFuncReturn(getInterfaceForLabel, "eth1")

	result := willGatewayInterfaceBeUpdated(protoData, "", settings)
	assert.True(t, result, "Expected true when labels are different")
}

func Test_willGatewayInterfaceBeUpdated_NoUpdateNeeded(t *testing.T) {
	protoData := &v1.Interface{}
	settings := nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}

	result := willGatewayInterfaceBeUpdated(protoData, "00:0A:95:9D:68:16", settings)
	assert.False(t, result, "Expected false when no update is needed")
}

func Test_getMacAddressFromSettings_Success(t *testing.T) {
	settings := nm.ConnectionSettings{common.EthernetType: map[string]interface{}{common.MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}

	result := getMacAddressFromSettings(settings)
	assert.Equal(t, "00:0A:95:9D:68:16", result, "Expected MAC address to be retrieved successfully")
}

func Test_changePriorityOfGatewayInterface_UpdateError(t *testing.T) {
	settings := nm.ConnectionSettings{common.IPV4Key: map[string]interface{}{}}
	connection := &mockgnm.MockConnection{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "Update", errors.New("error from Update"))

	err := changePriorityOfGatewayInterface(settings, connection)
	assert.Error(t, err, "Expected error from Update")
}

func Test_changePriorityOfGatewayInterface_Success(t *testing.T) {
	settings := nm.ConnectionSettings{common.IPV4Key: map[string]interface{}{}}
	connection := &mockgnm.MockConnection{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "Update", nil)

	err := changePriorityOfGatewayInterface(settings, connection)
	assert.NoError(t, err, "Expected no error when the route metric is updated successfully")
	assert.Equal(t, int32(-1), settings[common.IPV4Key][common.RouteMetricKey], "Expected Route Metric to be set to -1")
}

func Test_putExtraRoutes(t *testing.T) {
	intf := &v1.Interface{
		Routes: []*v1.Interface_Route{
			{
				Destination: "1.5.6.7",
				Netmask:     "255.255.254.0",
				NextHop:     "1.2.3.4",
				Metric:      1,
			},
			{
				Destination: "1.5.6.0",
				Netmask:     "255.255.255.0",
				NextHop:     "1.2.3.6",
			},
		},
	}

	out := nm.ConnectionSettings{}

	putExtraRoutes(intf, out)

	dbusDicts, ok := out[common.IPV4Key][common.RouteDataKey].([]DBusDict)

	assert.Equal(t, ok, true, "expected a DBusDict")

	dbusDict := dbusDicts[0]

	assert.Equal(t, intf.Routes[0].Destination, dbusDict["dest"].Value(), "destination match")
	assert.Equal(t, uint32(23), dbusDict["prefix"].Value(), "netmask match")
	assert.Equal(t, intf.Routes[0].NextHop, dbusDict["next-hop"].Value(), "next hop match")
	assert.Equal(t, intf.Routes[0].Metric, dbusDict["metric"].Value(), "metric match")

	dbusDict = dbusDicts[1]

	assert.Equal(t, intf.Routes[1].Destination, dbusDict["dest"].Value(), "destination match 2")
	assert.Equal(t, uint32(24), dbusDict["prefix"].Value(), "netmask match 2")
	assert.Equal(t, intf.Routes[1].NextHop, dbusDict["next-hop"].Value(), "next hop match 2")
	assert.Equal(t, intf.Routes[1].Metric, dbusDict["metric"].Value(), "metric match 2")
}

func Test_isInterfaceConfigured(t *testing.T) {
	gsmType := v1.Interface_GSM

	tests := []struct {
		name     string
		iface    *v1.Interface
		expected bool
	}{
		{
			name:     "nil interface",
			iface:    nil,
			expected: false,
		},
		{
			name: "GSM interface with config",
			iface: &v1.Interface{
				InterfaceName:    "gsm0",
				GsmConfiguration: &v1.Interface_GsmConf{Apn: "internet"},
			},
			expected: true,
		},
		{
			name: "GSM interface by type only",
			iface: &v1.Interface{
				InterfaceName: "gsm0",
				InterfaceType: &gsmType,
			},
			expected: true,
		},
		{
			name: "Ethernet with DHCP enabled",
			iface: &v1.Interface{
				InterfaceName: "eth0",
				MacAddress:    "00:0A:95:9D:68:16",
				DHCP:          common.Enabled,
			},
			expected: true,
		},
		{
			name: "Ethernet with valid static IP",
			iface: &v1.Interface{
				InterfaceName: "eth0",
				MacAddress:    "00:0A:95:9D:68:16",
				DHCP:          common.Disabled,
				Static:        &v1.Interface_StaticConf{IPv4: "192.168.1.100", NetMask: "255.255.255.0", Gateway: "192.168.1.1"},
			},
			expected: true,
		},
		{
			name: "Unconfigured - empty DHCP, nil static",
			iface: &v1.Interface{
				InterfaceName: "eth2",
				MacAddress:    "00:15:5D:38:01:60",
			},
			expected: false,
		},
		{
			name: "Unconfigured - disabled DHCP, nil static",
			iface: &v1.Interface{
				InterfaceName: "eth2",
				MacAddress:    "00:15:5D:38:01:60",
				DHCP:          common.Disabled,
			},
			expected: false,
		},
		{
			name: "Unconfigured - empty DHCP, empty static IPv4",
			iface: &v1.Interface{
				InterfaceName: "eth2",
				MacAddress:    "00:15:5D:38:01:60",
				Static:        &v1.Interface_StaticConf{IPv4: ""},
			},
			expected: false,
		},
		{
			name: "Unconfigured - disabled DHCP, empty static IPv4",
			iface: &v1.Interface{
				InterfaceName: "eth2",
				MacAddress:    "00:15:5D:38:01:60",
				DHCP:          common.Disabled,
				Static:        &v1.Interface_StaticConf{IPv4: ""},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isInterfaceConfigured(tc.iface)
			assert.Equal(t, tc.expected, result)
		})
	}
}
