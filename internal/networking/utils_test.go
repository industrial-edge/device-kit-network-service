/*
 * Copyright © Siemens 2024 - 2025. ALL RIGHTS RESERVED.
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
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"os"
	"reflect"
	"testing"
	"time"

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

func Test_RetrieveSettingsFromBackup_ReturnsCorrectSettings(t *testing.T) {
	backup := nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			IDKey:            "connection-id",
			TypeKey:          "connection-type",
			InterfaceNameKey: "eth0",
			UUIDKey:          "backup-uuid",
			TimeStampKey:     int64(1234567890),
		},
		EthernetType: map[string]interface{}{
			MACAddressKey: "00:11:22:33:44:55",
		},
		IPV4Key: map[string]interface{}{
			AddressDataKey: []map[string]interface{}{
				{
					AddressKey: "192.168.1.1",
					PrefixKey:  uint32(24),
				},
			},
		},
	}

	connection := retrieveSettingsFromBackup(backup)

	// Verify that the input and output objects are different instances
	assert.NotEqual(t, &backup, &connection, "Input and output objects should be different instances")

	// Verify that the entire object is correctly transformed
	assert.Equal(t, backup[ConnectionKey][IDKey], connection[ConnectionKey][IDKey], "ID should be the same")
	assert.Equal(t, backup[ConnectionKey][TypeKey], connection[ConnectionKey][TypeKey], "Type should be the same")
	assert.Equal(t, backup[ConnectionKey][InterfaceNameKey], connection[ConnectionKey][InterfaceNameKey], "InterfaceName should be the same")
	assert.NotEqual(t, backup[ConnectionKey][UUIDKey], connection[ConnectionKey][UUIDKey], "UUID should be different")
	assert.Equal(t, backup[EthernetType][MACAddressKey], connection[EthernetType][MACAddressKey], "MACAddress should be the same")
	assert.Equal(t, backup[IPV4Key][AddressDataKey], connection[IPV4Key][AddressDataKey], "AddressData should be the same")

	expectedTimeStamp := connection[ConnectionKey][TimeStampKey].(int64)
	currentTime := time.Now().UnixNano()
	assert.True(t, expectedTimeStamp >= 1234567890 && expectedTimeStamp <= currentTime, "Timestamp should be within a valid range")
}

func Test_ParseStaticIPConfig_ReturnsCorrectConfig(t *testing.T) {
	expected := getMockInterfaceStaticConf()
	connection := nm.ConnectionSettings{
		IPV4Key: map[string]interface{}{
			AddressDataKey: []map[string]interface{}{
				{
					AddressKey: "192.168.1.1",
					PrefixKey:  uint32(24),
				},
			},
			GatewayKey: "192.168.1.254",
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
		ConnectionKey: map[string]interface{}{
			InterfaceNameKey: "eth0",
			TypeKey:          EthernetType,
		},
	}, nil)

	mockConnection2.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			InterfaceNameKey: "eth0",
			TypeKey:          EthernetType,
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
		ConnectionKey: map[string]interface{}{
			TypeKey: "wifi",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.False(t, result, "Expected false when connection type is not Ethernet")
}

func TestIsValidConnection_InterfaceNameMatches(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			TypeKey:          EthernetType,
			InterfaceNameKey: "eth0",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.True(t, result, "Expected true when interface name matches")
}

func TestIsValidConnection_InterfaceNameDoesNotMatch(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			TypeKey:          EthernetType,
			InterfaceNameKey: "eth1",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.False(t, result, "Expected false when interface name does not match")
}

func TestIsValidConnection_UUIDAndIDMatch(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			TypeKey: EthernetType,
			UUIDKey: "uuid",
			IDKey:   "id",
		},
	}, nil)

	result := isValidConnection(connection, "eth0", "uuid", "id")
	assert.True(t, result, "Expected true when UUID and ID match")
}

func TestIsValidConnection_UUIDAndIDDoNotMatch(t *testing.T) {
	connection := &mockgnm.MockConnection{}
	connection.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			TypeKey: EthernetType,
			UUIDKey: "different-uuid",
			IDKey:   "different-id",
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

	mockDeviceWired.On("GetPropertyActiveConnection").Return(mockActiveConnection, errors.New("error from GetPropertyActiveConnection"))
	mockDeviceWired.On("GetPropertyHwAddress").Return(testMac, nil)
	mockDeviceWired.On("GetPropertyInterface").Return(testInterface, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{}
	})

	patches.ApplyFunc(dockerNetworkGetMacvlanConnection, func(_ string) *v1.Interface_L2 {
		return expectedL2Conf
	})

	patches.ApplyFunc(getLabelForInterface, func(interfaceName string) (string, error) {
		if interfaceName == testInterface {
			return expectedLabel, nil
		}
		return "", nil
	})

	// Call the function
	result := DBusToProto(mockDeviceWired)
	// Assertions
	assert.Equal(t, testInterface, result.InterfaceName, "DBusToProto should return an Interface with the correct interface name")
	assert.Equal(t, testMac, result.MacAddress, "DBusToProto should return an Interface with the correct MAC address")
	assert.Equal(t, expectedL2Conf, result.L2Conf, "DBusToProto should return an Interface with the correct L2 config")
	assert.Equal(t, expectedLabel, result.Label, "DBusToProto should return an Interface with the correct label")
}

func Test_DBusToProto_ReturnsInterfaceWithFirstConnectionWhenActiveConnectionNotAvailable(t *testing.T) {
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockConnection1 := &mockgnm.MockConnection{}
	mockConnection2 := &mockgnm.MockConnection{}
	mockActiveConnection := &mockgnm.MockActiveConnection{}
	testInterface := "eth0"
	expectedInterface := &v1.Interface{
		MacAddress:       "00:0A:95:9D:68:16",
		Label:            "testLabel",
		DHCP:             Enabled,
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        nil,
		GatewayInterface: false,
	}

	mockDeviceWired.On("GetPropertyActiveConnection").Return(mockActiveConnection, errors.New("error from GetPropertyActiveConnection"))
	mockDeviceWired.On("GetPropertyHwAddress").Return(expectedInterface.MacAddress, nil)
	mockDeviceWired.On("GetPropertyInterface").Return(testInterface, nil)
	mockConnection1.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			"id": "connection1",
		},
	}, nil)
	mockConnection2.On("GetSettings").Return(nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			"id": "connection2",
		},
	}, nil)

	// Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Patch listConnections function to return a non-empty list
	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{mockConnection1, mockConnection2}
	})

	// Patch convertToProto function
	patches.ApplyFunc(convertToProto, func(settings nm.ConnectionSettings, ipv4conf nm.IP4Config, mac string) *v1.Interface {
		if settings[ConnectionKey]["id"] == "connection1" && mac == expectedInterface.MacAddress && ipv4conf == nil {
			return expectedInterface
		}
		return nil
	})

	patches.ApplyFunc(getLabelForInterface, func(_ string) (string, error) {
		return expectedInterface.Label, nil
	})

	// Call the function
	result := DBusToProto(mockDeviceWired)

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
		DHCP:             Enabled,
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        getMockInterfaceDNSConfig(),
		GatewayInterface: true,
	}
	testInterface := "eth0"

	mockDeviceWired.On("GetPropertyActiveConnection").Return(mockActiveConnection, nil)
	mockDeviceWired.On("GetPropertyHwAddress").Return(expectedInterface.MacAddress, nil)
	mockDeviceWired.On("GetPropertyInterface").Return(testInterface, nil)
	mockActiveConnection.On("GetPropertyIP4Config").Return(mockIP4Config, nil)
	mockActiveConnection.On("GetPropertyConnection").Return(mockConnection, nil)
	mockConnection.On("GetSettings").Return(nm.ConnectionSettings{}, nil)

	// Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{}
	})

	// Patch convertToProto function
	patches.ApplyFunc(convertToProto, func(settings nm.ConnectionSettings, ipv4conf nm.IP4Config, mac string) *v1.Interface {
		if mac == expectedInterface.MacAddress && ipv4conf == mockIP4Config {
			return expectedInterface
		}
		return nil
	})

	// Call the function
	result := DBusToProto(mockDeviceWired)

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
		IPV4Key: map[string]interface{}{
			MethodKey:      Auto,
			RouteMetricKey: int64(1),
		},
	}

	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(parseDHCPIPv4Config, func(ipv4conf nm.IP4Config) *v1.Interface_StaticConf {
		return getMockInterfaceStaticConf()
	})

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, Enabled, result.DHCP, "DHCP should be enabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
}

func Test_ConvertToProto_ReturnsCorrectProtoWhenStaticIP(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		IPV4Key: map[string]interface{}{
			AddressDataKey: []map[string]interface{}{
				{
					AddressKey: "192.168.1.1",
					PrefixKey:  uint32(24),
				},
			},
			GatewayKey:     "192.168.1.254",
			RouteMetricKey: int64(1),
		},
	}

	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, Disabled, result.DHCP, "DHCP should be disabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
}

func Test_ConvertToProto_ReturnsCorrectProtoWithDHCPIPConfig(t *testing.T) {
	mac := "f7:2b:a1:d5:97:4e"
	mockNsData := getMockIP4NsData()
	mockIPV4Config := new(mockgnm.MockIP4Config)
	connection := nm.ConnectionSettings{
		IPV4Key: map[string]interface{}{
			MethodKey: Auto,
			AddressDataKey: []map[string]interface{}{
				{
					AddressKey: "192.168.1.1",
					PrefixKey:  uint32(24),
				},
			},
			GatewayKey:     "192.168.1.254",
			RouteMetricKey: int64(1),
		},
	}

	mockIPV4Config.On("GetPropertyAddressData").Return([]nm.IP4AddressData{{Address: "192.168.1.1", Prefix: uint8(24)}}, nil)
	mockIPV4Config.On("GetPropertyGateway").Return("192.168.1.254", nil)
	mockIPV4Config.On("GetPropertyNameserverData").Return(mockNsData, nil)

	result := convertToProto(connection, mockIPV4Config, mac)

	assert.Equal(t, Enabled, result.DHCP, "DHCP should be enabled")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", result.MacAddress)
}

func Test_NewSettingsFromProto_ReturnsSettingsWhenDHCPEnabled(t *testing.T) {
	deviceName := "eth0"
	protoData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       "20:87:56:b5:ed:e0",
		DHCP:             "enabled",
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "8.4.4.4"},
	}

	expectedTimestamp := int64(1727351248)
	expectedUUID := "7f9d7444-0dd7-431f-b67a-f9eb78f12af7"

	expectedSettings := nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			IDKey:            "20:87:56:b5:ed:e0_dhcp",
			UUIDKey:          expectedUUID,
			TimeStampKey:     expectedTimestamp,
			TypeKey:          EthernetType,
			InterfaceNameKey: deviceName,
		},
		IPV4Key: map[string]interface{}{
			MethodKey:        Auto,
			DNSKey:           []uint32{1234567890, 1234567890},
			DNSIgnoreAutoKey: Yes,
			RouteMetricKey:   1,
		},
		EthernetType: map[string]interface{}{
			MACAddressKey: net.HardwareAddr{},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(time.Now, func() time.Time {
		return time.Unix(1727351248, 0)
	})

	patches.ApplyFunc(uuid.New, func() uuid.UUID {
		return uuid.Must(uuid.Parse(expectedUUID))
	})

	patches.ApplyFunc(IPToUInt32LI, func(ip string) uint32 {
		return 1234567890
	})

	patches.ApplyFunc(net.ParseMAC, func(mac string) (net.HardwareAddr, error) {
		return net.HardwareAddr{}, nil
	})

	settings := newSettingsFromProto(protoData, deviceName)

	assert.NotNil(t, settings, "newSettingsFromProto should return non-nil result")
	assert.Equal(t, expectedSettings, settings)
}

func Test_NewSettingsFromProto_ReturnsSettingsWithDefaultsWhenPutMacAddressFails(t *testing.T) {
	deviceName := "eth0"
	protoData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       "invalid mac address",
		DHCP:             "enabled",
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "8.4.4.4"},
	}

	expectedTimestamp := int64(1727351248)
	expectedUUID := "7f9d7444-0dd7-431f-b67a-f9eb78f12af7"
	expectedSettings := nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			IDKey:            "invalid mac address_dhcp",
			UUIDKey:          expectedUUID,
			TimeStampKey:     expectedTimestamp,
			TypeKey:          EthernetType,
			InterfaceNameKey: "eth0",
		},
		IPV4Key: map[string]interface{}{
			MethodKey:        Auto,
			DNSKey:           []uint32{1234567890, 1234567890},
			DNSIgnoreAutoKey: Yes,
			RouteMetricKey:   1,
		},
		EthernetType: map[string]interface{}{
			MACAddressKey: net.HardwareAddr{},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(time.Now, func() time.Time {
		return time.Unix(1727351248, 0)
	})

	patches.ApplyFunc(uuid.New, func() uuid.UUID {
		return uuid.Must(uuid.Parse("7f9d7444-0dd7-431f-b67a-f9eb78f12af7"))
	})

	patches.ApplyFunc(IPToUInt32LI, func(ip string) uint32 {
		return 1234567890
	})

	patches.ApplyFunc(net.ParseMAC, func(mac string) (net.HardwareAddr, error) {
		return net.HardwareAddr{}, errors.New("error from ParseMAC")
	})

	settings := newSettingsFromProto(protoData, deviceName)

	assert.Equal(t, expectedSettings, settings)
	assert.NotNil(t, settings, "newSettingsFromProto should return non-nil result")

}

func Test_NewSettingsFromProto_ReturnsSettingsWithStaticIPWhenDHCPDisabled(t *testing.T) {
	deviceName := "eth0"
	protoData := &v1.Interface{
		GatewayInterface: true,
		Label:            "eth0",
		DHCP:             "disabled",
		Static:           getMockInterfaceStaticConf(),
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "8.8.4.4"},
	}

	expectedTimestamp := int64(1727351248)
	expectedUUID := "7f9d7444-0dd7-431f-b67a-f9eb78f12af7"

	expectedSettings := nm.ConnectionSettings{
		ConnectionKey: map[string]interface{}{
			IDKey:            "eth0_static",
			UUIDKey:          expectedUUID,
			TimeStampKey:     expectedTimestamp,
			TypeKey:          EthernetType,
			InterfaceNameKey: deviceName,
		},
		IPV4Key: map[string]interface{}{
			MethodKey:      Manual,
			DNSKey:         []uint32{1234567890, 1234567890},
			RouteMetricKey: 1,
			AddressDataKey: []DBusDict{
				{
					AddressKey: dbus.MakeVariantWithSignature(nil, dbus.ParseSignatureMust("")),
					PrefixKey:  dbus.MakeVariantWithSignature(nil, dbus.ParseSignatureMust("")),
				},
			},
			GatewayKey: "192.168.1.254",
		},
		EthernetType: map[string]interface{}{
			MACAddressKey: net.HardwareAddr{},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(time.Now, func() time.Time {
		return time.Unix(1727351248, 0)
	})

	patches.ApplyFunc(uuid.New, func() uuid.UUID {
		return uuid.Must(uuid.Parse(expectedUUID))
	})

	patches.ApplyFunc(IPToUInt32LI, func(ip string) uint32 {
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
	assert.Equal(t, expectedSettings, settings)
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

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: nil}}, nil)
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

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}, nil)
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

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}, nil)
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

	patches.ApplyMethodReturn(connection, "GetSettings", nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}, nil)
	patches.ApplyFuncReturn(willGatewayInterfaceBeUpdated, false)

	err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator)
	assert.NoError(t, err, "Expected no error when the connection is processed successfully")
}

func Test_setMacAddressInSettings_InvalidMacAddress(t *testing.T) {
	settings := nm.ConnectionSettings{EthernetType: map[string]interface{}{}}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyPermHwAddress", "invalid-mac-address", nil)
	patches.ApplyFuncReturn(net.ParseMAC, nil, errors.New("error from ParseMAC"))

	err := setMacAddressInSettings(settings, ethernetDevice)
	assert.Error(t, err, "Expected error from ParseMAC")
}

func Test_setMacAddressInSettings_Success(t *testing.T) {
	settings := nm.ConnectionSettings{EthernetType: map[string]interface{}{}}
	ethernetDevice := &mockgnm.MockDeviceWired{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(ethernetDevice, "GetPropertyPermHwAddress", "00:0A:95:9D:68:16", nil)
	patches.ApplyFuncReturn(net.ParseMAC, net.HardwareAddr{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}, nil)

	err := setMacAddressInSettings(settings, ethernetDevice)
	assert.NoError(t, err, "Expected no error when the MAC address is set successfully")
	assert.Equal(t, net.HardwareAddr{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}, net.HardwareAddr(settings[EthernetType][MACAddressKey].([]uint8)), "Expected MAC address to be set in the backup")
}

func Test_willGatewayInterfaceBeUpdated_MacAddressDifferent(t *testing.T) {
	protoData := &v1.Interface{MacAddress: "00:0A:95:9D:68:16"}
	settings := nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x17}}}

	result := willGatewayInterfaceBeUpdated(protoData, "00:0A:95:9D:68:17", settings)
	assert.True(t, result, "Expected true when MAC addresses are different")
}

func Test_willGatewayInterfaceBeUpdated_LabelDifferent(t *testing.T) {
	protoData := &v1.Interface{Label: "label1"}
	settings := nm.ConnectionSettings{ConnectionKey: map[string]interface{}{InterfaceNameKey: "eth0"}}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFuncReturn(getInterfaceForLabel, "eth1")

	result := willGatewayInterfaceBeUpdated(protoData, "", settings)
	assert.True(t, result, "Expected true when labels are different")
}

func Test_willGatewayInterfaceBeUpdated_NoUpdateNeeded(t *testing.T) {
	protoData := &v1.Interface{}
	settings := nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}

	result := willGatewayInterfaceBeUpdated(protoData, "00:0A:95:9D:68:16", settings)
	assert.False(t, result, "Expected false when no update is needed")
}

func Test_getMacAddressFromSettings_Success(t *testing.T) {
	settings := nm.ConnectionSettings{EthernetType: map[string]interface{}{MACAddressKey: []byte{0x00, 0x0A, 0x95, 0x9D, 0x68, 0x16}}}

	result := getMacAddressFromSettings(settings)
	assert.Equal(t, "00:0A:95:9D:68:16", result, "Expected MAC address to be retrieved successfully")
}

func Test_changePriorityOfGatewayInterface_UpdateError(t *testing.T) {
	settings := nm.ConnectionSettings{IPV4Key: map[string]interface{}{}}
	connection := &mockgnm.MockConnection{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "Update", errors.New("error from Update"))

	err := changePriorityOfGatewayInterface(settings, connection)
	assert.Error(t, err, "Expected error from Update")
}

func Test_changePriorityOfGatewayInterface_Success(t *testing.T) {
	settings := nm.ConnectionSettings{IPV4Key: map[string]interface{}{}}
	connection := &mockgnm.MockConnection{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(connection, "Update", nil)

	err := changePriorityOfGatewayInterface(settings, connection)
	assert.NoError(t, err, "Expected no error when the route metric is updated successfully")
	assert.Equal(t, int32(-1), settings[IPV4Key][RouteMetricKey], "Expected Route Metric to be set to -1")
}
