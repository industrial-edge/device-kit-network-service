/*
 * Copyright © Siemens 2020 - 2025. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	"fmt"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"reflect"
	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func Test_NewNetworkConfiguratorWithNM_ReturnsNonNilInstance(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	nc := NewNetworkConfiguratorWithNM(mockNetworkManager)

	assert.NotNil(t, nc, "NewNetworkConfiguratorWithNM should return a non-nil NetworkConfigurator instance")
	assert.Equal(t, mockNetworkManager, nc.gnm, "NewNetworkConfiguratorWithNM should set the provided NetworkManager instance")
}

func Test_NewNetworkConfigurator_ReturnsNonNilInstance(t *testing.T) {
	nc := NewNetworkConfigurator()
	assert.NotNil(t, nc, "NewNetworkConfigurator should return a non-nil NetworkConfigurator instance")
}

func Test_IsGatewayInterface_ReturnsTrueForMatchingMac(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevices := []nm.DeviceWired{
		&mockgnm.MockDeviceWired{},
	}
	testMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return mockDevices
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "findGatewayMAC", func(_ *NetworkConfigurator, devices []nm.DeviceWired) string {
		return testMac
	})

	result := nc.IsGatewayInterface(testMac)

	assert.True(t, result, "IsGatewayInterface should return true for the gateway MAC")
}

func Test_IsGatewayInterface_ReturnsFalseForNonMatchingMac(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevices := []nm.DeviceWired{
		&mockgnm.MockDeviceWired{},
	}
	gatewayMac := "00:0a:95:9d:68:16"
	nonGatewayMac := "11:22:33:44:55:66"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return mockDevices
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "findGatewayMAC", func(_ *NetworkConfigurator, devices []nm.DeviceWired) string {
		return gatewayMac
	})

	result := nc.IsGatewayInterface(nonGatewayMac)

	assert.False(t, result, "IsGatewayInterface should return false for a non-gateway MAC")
}

func Test_findGatewayMAC_ReturnsGatewayMacWithLowestMetric(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevices := []nm.DeviceWired{
		&mockgnm.MockDeviceWired{},
		&mockgnm.MockDeviceWired{},
	}

	expectedMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	callCount := 0
	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceGatewayMACAndMetric", func(_ *NetworkConfigurator, device nm.DeviceWired) (string, uint8, error) {
		if callCount == 0 {
			callCount++
			return "11:22:33:44:55:66", 50, nil
		}
		return expectedMac, 10, nil
	})

	result := nc.findGatewayMAC(mockDevices)

	assert.Equal(t, expectedMac, result, "findGatewayMAC should return the MAC address with the lowest metric")
}

func Test_findGatewayMAC_ReturnsEmptyWhenNoDevicesHaveGateway(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevices := []nm.DeviceWired{
		&mockgnm.MockDeviceWired{},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceGatewayMACAndMetric", func(_ *NetworkConfigurator, device nm.DeviceWired) (string, uint8, error) {
		return "", 0, fmt.Errorf("no gateway found")
	})

	result := nc.findGatewayMAC(mockDevices)

	assert.Equal(t, "", result, "findGatewayMAC should return an empty string when no gateway MAC is found")
}
func Test_getDeviceGatewayMACAndMetric_ReturnsCorrectMACAndMetric(t *testing.T) {
    nc := &NetworkConfigurator{}
    mockDevice := &mockgnm.MockDeviceWired{}
    mockConn := &mockgnm.MockActiveConnection{}
    mockIPv4Config := &mockgnm.MockIP4Config{}

    mockRouteData := []nm.IP4RouteData{
        {
            Destination: "0.0.0.0",
            Prefix:      0,
            NextHop:     "192.168.1.1",
            Metric:      10,
        },
    }

    // Mock method returns
    mockDevice.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)
    mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)
    mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
    mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

    // Test the function
    mac, metric, err := nc.getDeviceGatewayMACAndMetric(mockDevice)

    // Assertions
    assert.NoError(t, err, "Expected no error")
    assert.Equal(t, "F7:2B:A1:D5:97:4E", mac, "Expected correct MAC address")
    assert.Equal(t, uint8(10), metric, "Expected correct metric value")
}

func Test_GetInterfaceWithMac_ReturnsCorrectInterface(t *testing.T) {
	var actualDeviceType reflect.Type
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	testMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return mockDevice
	})

	patches.ApplyFunc(DBusToProto, func(device nm.DeviceWired) *v1.Interface {
		actualDeviceType = reflect.TypeOf(device)
		return &v1.Interface{MacAddress: testMac}
	})

	result := nc.GetInterfaceWithMac(testMac)

	assert.Equal(t, reflect.TypeOf(mockDevice), actualDeviceType, "Expected device type should be the same as mockDevice")
	assert.NotNil(t, result, "GetInterfaceWithMac should return a non-nil Interface instance")
	assert.Equal(t, testMac, result.MacAddress, "GetInterfaceWithMac should return an Interface with the correct MAC address")
}

func Test_GetInterfaceWithMac_ReturnsNilWhenDeviceIsNil(t *testing.T) {
	nc := &NetworkConfigurator{}
	testMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return nil
	})

	result := nc.GetInterfaceWithMac(testMac)

	assert.Nil(t, result, "GetInterfaceWithMac should return a nil Interface instance when device is nil")
}

func Test_GetInterfaceWithLabel_ReturnsCorrectInterface(t *testing.T) {
	var actualDeviceType reflect.Type
	testLabel := "eth0"
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithLabel",
		func(_ *NetworkConfigurator, label string) nm.DeviceWired { return mockDevice })

	patches.ApplyFunc(DBusToProto, func(device nm.DeviceWired) *v1.Interface {
		actualDeviceType = reflect.TypeOf(device)
		return &v1.Interface{Label: testLabel}
	})

	result := nc.GetInterfaceWithLabel(testLabel)

	assert.Equal(t, reflect.TypeOf(mockDevice), actualDeviceType, "Expected device type should be the same as mockDevice")
	assert.NotNil(t, result, "GetInterfaceWithLabel should return a non-nil Interface instance")
	assert.Equal(t, testLabel, result.Label, "GetInterfaceWithLabel should return an Interface with the correct label")
}

func Test_GetInterfaceWithLabel_ReturnsNilWhenDeviceNotFound(t *testing.T) {
	testLabel := "eth0"
	nc := &NetworkConfigurator{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithLabel",
		func(_ *NetworkConfigurator, label string) nm.DeviceWired { return nil })

	result := nc.GetInterfaceWithLabel(testLabel)

	assert.Nil(t, result, "GetInterfaceWithLabel should return a nil Interface instance when device is not found")
}

func Test_GetEthernetInterfaces_ReturnsAllAvailableNetworkDevices(t *testing.T) {
	var actualDeviceType reflect.Type
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}

	mockConn := &mockgnm.MockActiveConnection{}
	mockIPv4Config := &mockgnm.MockIP4Config{}

	// Replace networking.RouteData with gonetworkmanager.IP4RouteData
	mockRouteData := []nm.IP4RouteData{
		{
			Destination: "0.0.0.0",
			Prefix:      0,
			NextHop:     "192.168.1.1",
			Metric:      10,
		},
	}

	mockDevice.On("GetPropertyInterface").Return("eth0", nil)
	mockDevice.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)

	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return []nm.DeviceWired{mockDevice, mockDevice}
	})

	patches.ApplyFunc(DBusToProto, func(device nm.DeviceWired) *v1.Interface {
		actualDeviceType = reflect.TypeOf(device)
		return &v1.Interface{
			InterfaceName: "eth0",
			MacAddress:    "00:0a:95:9d:68:16",
		}
	})

	result := nc.GetEthernetInterfaces()

	assert.Equal(t, reflect.TypeOf(mockDevice), actualDeviceType, "Expected device type should be the same as mockDevice")
	assert.NotNil(t, result, "GetEthernetInterfaces should return non-nil Interface instances")
	assert.Equal(t, 2, len(result), "GetEthernetInterfaces should return the correct number of interfaces")
	assert.Equal(t, "eth0", result[0].InterfaceName, "First interface should have the correct name")
	assert.Equal(t, "eth0", result[1].InterfaceName, "Second interface should have the correct name")
}

func Test_GetEthernetInterfaces_EnsuresGatewayInterfaceIsMarked(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice1 := &mockgnm.MockDeviceWired{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	mockConn1 := &mockgnm.MockActiveConnection{}
	mockConn2 := &mockgnm.MockActiveConnection{}

	mockIPv4Config1 := &mockgnm.MockIP4Config{}
	mockIPv4Config2 := &mockgnm.MockIP4Config{}

	mockRouteData1 := []nm.IP4RouteData{
		{
			Destination: "0.0.0.0",
			Prefix:      0,
			NextHop:     "192.168.1.1",
			Metric:      10, // Lower metric, this will be marked as the gateway interface
		},
	}
	mockRouteData2 := []nm.IP4RouteData{
		{
			Destination: "0.0.0.0",
			Prefix:      0,
			NextHop:     "192.168.2.1",
			Metric:      20,
		},
	}

	// Mock behaviors for Device 1
	mockDevice1.On("GetPropertyInterface").Return("eth0", nil)
	mockDevice1.On("GetPropertyActiveConnection").Return(mockConn1, nil)
	mockConn1.On("GetPropertyIP4Config").Return(mockIPv4Config1, nil)
	mockIPv4Config1.On("GetPropertyRouteData").Return(mockRouteData1, nil)
	
	// Mock behaviors for Device 2
	mockDevice2.On("GetPropertyInterface").Return("eth1", nil)
	mockDevice2.On("GetPropertyActiveConnection").Return(mockConn2, nil)
	mockConn2.On("GetPropertyIP4Config").Return(mockIPv4Config2, nil)
	mockIPv4Config2.On("GetPropertyRouteData").Return(mockRouteData2, nil)

	mockDevice1.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)
	mockDevice2.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4", nil)
	
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return []nm.DeviceWired{mockDevice1, mockDevice2}
	})

	patches.ApplyFunc(DBusToProto, func(device nm.DeviceWired) *v1.Interface {
		name, _ := device.GetPropertyInterface()
		return &v1.Interface{
			InterfaceName: name,
			MacAddress:    "00:0a:95:9d:68:16",
		}
	})

	result := nc.GetEthernetInterfaces()

	// Assert that at least one interface has GatewayInterface set to true
	gatewayInterfaceFound := false
	for _, iface := range result {
		if iface.GatewayInterface {
			gatewayInterfaceFound = true
			break
		}
	}

	assert.True(t, gatewayInterfaceFound, "At least one network interface should have GatewayInterface set to true")
	assert.Equal(t, 2, len(result), "Expected two interfaces")
	assert.Equal(t, "eth0", result[0].InterfaceName, "First interface should be eth0")
	assert.Equal(t, "eth1", result[1].InterfaceName, "Second interface should be eth1")
}

func Test_ArePreconditionsOk_ReturnsTrueWithValidSettings(t *testing.T) {
	nc := &NetworkConfigurator{}
	newSettings := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{MacAddress: "00:0a:95:9d:68:16"}},
	}

	var verifyCallCount int
	var verifyArg1 *v1.NetworkSettings
	var verifyArg2 *NetworkConfigurator

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(verify, func(arg1 *v1.NetworkSettings, arg2 *NetworkConfigurator) (bool, error) {
		verifyCallCount++
		verifyArg1 = arg1
		verifyArg2 = arg2
		return true, nil
	})

	result, err := nc.ArePreconditionsOk(newSettings)

	assert.True(t, result, "ArePreconditionsOk should return true")
	assert.Nil(t, err, "ArePreconditionsOk should not return an error")

	assert.Equal(t, 1, verifyCallCount, "verify should be called once")
	assert.Equal(t, newSettings, verifyArg1, "verify should be called with the correct NetworkSettings argument")
	assert.Equal(t, nc, verifyArg2, "verify should be called with the correct NetworkConfigurator argument")

}

func Test_Apply_ReturnsNilWhenNoBackup(t *testing.T) {
	nc := &NetworkConfigurator{}
	newSettings := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{MacAddress: "00:0a:95:9d:68:16"},
			{MacAddress: "00:0a:95:9d:68:17"},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "applyAndBackupSettings",
		func(_ *NetworkConfigurator, _ *v1.Interface) (nm.ConnectionSettings, error) {
			return nil, nil // return nil backup
		})

	err := nc.Apply(newSettings)

	assert.Nil(t, err, "Apply should not return an error when applyAndBackupSettings does not return an error and there is no backup")
}

func Test_Apply_ReturnsErrorWhenTryApplyFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	newSettings := &v1.NetworkSettings{
		Interfaces: []*v1.Interface{
			{MacAddress: "00:0a:95:9d:68:16"},
			{MacAddress: "00:0a:95:9d:68:17"},
		},
	}

	mockBackup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	expectedError := errors.New("test error from applyAndBackupSettings")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "applyAndBackupSettings",
		func(_ *NetworkConfigurator, _ *v1.Interface) (nm.ConnectionSettings, error) {
			return mockBackup, expectedError
		})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "restoreConnection", func(_ *NetworkConfigurator, backup nm.ConnectionSettings) error {
		return nil
	})

	err := nc.Apply(newSettings)

	assert.NotNil(t, err, "Apply should return an error when applyAndBackupSettings returns an error")
	assert.Equal(t, expectedError, err, "Apply should return the same error as applyAndBackupSettings")
}

func Test_GetDeviceWithMac_ReturnsDeviceWithCorrectMacAddress(t *testing.T) {
	nc := &NetworkConfigurator{}
	testMac := "00:0a:95:9d:68:16"
	mockDevice := &mockgnm.MockDeviceWired{}

	mockDevice.On("GetPropertyHwAddress").Return(testMac, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices",
		func(_ *NetworkConfigurator) []nm.DeviceWired { return []nm.DeviceWired{mockDevice} })

	result := nc.getDeviceWithMac(testMac)
	hwAddress, _ := result.GetPropertyHwAddress()

	assert.NotNil(t, result, "getDeviceWithMac should return a non-nil DeviceWired instance")
	assert.Equal(t, testMac, hwAddress, "getDeviceWithMac should return a DeviceWired with the correct MAC address")
	assert.Equal(t, mockDevice, result, "getDeviceWithMac should return the correct DeviceWired instance")
}

func Test_GetDeviceWithMac_ReturnsNilWhenNoMatchingDevice(t *testing.T) {
	nc := &NetworkConfigurator{}
	testMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock getAllEthernetDevices method to return an empty list
	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices",
		func(_ *NetworkConfigurator) []nm.DeviceWired { return []nm.DeviceWired{} })

	result := nc.getDeviceWithMac(testMac)

	assert.Nil(t, result, "getDeviceWithMac should return a nil DeviceWired instance when no device with the matching MAC address is found")
}

func Test_GetDeviceWithLabel_ReturnsDeviceWithCorrectLabel(t *testing.T) {
	testLabel := "eth0"
	otherLabel := "eth1"
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	otherMockDevice := &mockgnm.MockDeviceWired{}

	mockDevice.On("GetPropertyInterface").Return(testLabel, nil)
	otherMockDevice.On("GetPropertyInterface").Return(otherLabel, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices",
		func(_ *NetworkConfigurator) []nm.DeviceWired { return []nm.DeviceWired{mockDevice, otherMockDevice} })

	patches.ApplyFunc(getInterfaceForLabel, func(_ string) string { return testLabel })

	result := nc.getDeviceWithLabel(testLabel)
	interfaceName, _ := result.GetPropertyInterface()

	assert.NotNil(t, result, "getDeviceWithLabel should return a non-nil DeviceWired instance")
	assert.Equal(t, testLabel, interfaceName, "getDeviceWithLabel should return a DeviceWired with the correct label")
	assert.Equal(t, mockDevice, result, "getDeviceWithLabel should return the correct DeviceWired instance")
}

func Test_GetDeviceWithLabel_ReturnsNilWhenNoMatchingDevice(t *testing.T) {
	testLabel := "eth0"
	nc := &NetworkConfigurator{}

	mockDevice1 := &mockgnm.MockDeviceWired{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	mockDevice1.On("GetPropertyInterface").Return("eth1", nil)
	mockDevice2.On("GetPropertyInterface").Return("eth2", nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices",
		func(_ *NetworkConfigurator) []nm.DeviceWired { return []nm.DeviceWired{mockDevice1, mockDevice2} })

	patches.ApplyFunc(getInterfaceForLabel, func(_ string) string {
		return testLabel
	})

	result := nc.getDeviceWithLabel(testLabel)

	assert.Nil(t, result, "getDeviceWithLabel should return a nil DeviceWired instance when no device with the matching label is found")
}

func Test_GetAllEthernetDevices_ReturnsAllDevices(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockDeviceWiredForEthernet1 := &mockgnm.MockDeviceWired{}
	mockDeviceWiredForEthernet2 := &mockgnm.MockDeviceWired{}

	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list of devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockDeviceWiredForEthernet1, mockDeviceWiredForEthernet2}, nil
	})

	// Mock GetPropertyDeviceType to return Ethernet for both devices
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWiredForEthernet1), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWiredForEthernet2), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})

	// Mock GetPath to return different paths for each device
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWiredForEthernet1), "GetPath", func(_ nm.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/1"
	})
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWiredForEthernet2), "GetPath", func(_ nm.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/2"
	})

	// Mock NewDeviceWired to return the corresponding mock DeviceWired
	patches.ApplyFunc(nm.NewDeviceWired, func(path dbus.ObjectPath) (nm.DeviceWired, error) {
		if path == "/org/freedesktop/NetworkManager/Devices/1" {
			return mockDeviceWiredForEthernet1, nil
		}
		return mockDeviceWiredForEthernet2, nil
	})

	// Call the function
	result := nc.getAllEthernetDevices()

	// Assertions
	assert.Equal(t, 2, len(result), "Expected two Ethernet devices")
	assert.Equal(t, mockDeviceWiredForEthernet1, result[0], "Expected the first mock Ethernet device")
	assert.Equal(t, mockDeviceWiredForEthernet2, result[1], "Expected the second mock Ethernet device")
}

func Test_ApplyAndBackupSettings_ReturnsErrorWhenGetDeviceFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceBy", func(_ *NetworkConfigurator, _ *v1.Interface) (nm.DeviceWired, error) {
		return nil, errors.New("getDeviceBy error")
	})

	backup, err := nc.applyAndBackupSettings(protoData)

	assert.Nil(t, backup, "applyAndBackupSettings should return a nil backup when getDeviceBy fails")
	assert.NotNil(t, err, "applyAndBackupSettings should return an error when getDeviceBy fails")
	assert.Equal(t, "getDeviceBy error", err.Error(), "applyAndBackupSettings should return the correct error message")
}

func Test_ApplyAndBackupSettings_ReturnsErrorWhenPrepareSettingsFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}
	mockDevice := &mockgnm.MockDeviceWired{}
	expectedBackup := nm.ConnectionSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceBy", func(_ *NetworkConfigurator, _ *v1.Interface) (nm.DeviceWired, error) {
		return mockDevice, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "createBackupFromExisting", func(_ *NetworkConfigurator, _ nm.DeviceWired) nm.ConnectionSettings {
		return expectedBackup
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "prepareSettings", func(_ *NetworkConfigurator, _ *v1.Interface, _ nm.DeviceWired) (nm.ConnectionSettings, error) {
		return nil, errors.New("prepareSettings error")
	})

	backup, err := nc.applyAndBackupSettings(protoData)

	assert.Equal(t, expectedBackup, backup, "applyAndBackupSettings should return the correct backup when prepareSettings fails")
	assert.NotNil(t, err, "applyAndBackupSettings should return an error when prepareSettings fails")
	assert.Equal(t, "prepareSettings error", err.Error(), "applyAndBackupSettings should return the correct error message")
}

func Test_ApplyAndBackupSettings_ReturnsErrorWhenUpdateConnectionsFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}
	mockDevice := &mockgnm.MockDeviceWired{}
	expectedBackup := nm.ConnectionSettings{}
	mockSettings := nm.ConnectionSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceBy", func(_ *NetworkConfigurator, _ *v1.Interface) (nm.DeviceWired, error) {
		return mockDevice, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "createBackupFromExisting", func(_ *NetworkConfigurator, _ nm.DeviceWired) nm.ConnectionSettings {
		return expectedBackup
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "prepareSettings", func(_ *NetworkConfigurator, _ *v1.Interface, _ nm.DeviceWired) (nm.ConnectionSettings, error) {
		return mockSettings, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "updateConnections", func(_ *NetworkConfigurator, _ nm.DeviceWired, _ nm.ConnectionSettings) error {
		return errors.New("updateConnections error")
	})

	backup, err := nc.applyAndBackupSettings(protoData)

	assert.Equal(t, expectedBackup, backup, "applyAndBackupSettings should return the correct backup when updateConnections fails")
	assert.NotNil(t, err, "applyAndBackupSettings should return an error when updateConnections fails")
	assert.Equal(t, "updateConnections error", err.Error(), "applyAndBackupSettings should return the correct error message")
}

func Test_ApplyAndBackupSettings_ReturnsErrorWhenSetRouteMetricFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}
	mockDevice := &mockgnm.MockDeviceWired{}
	expectedBackup := nm.ConnectionSettings{}
	mockSettings := nm.ConnectionSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceBy", func(_ *NetworkConfigurator, _ *v1.Interface) (nm.DeviceWired, error) {
		return mockDevice, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "createBackupFromExisting", func(_ *NetworkConfigurator, _ nm.DeviceWired) nm.ConnectionSettings {
		return expectedBackup
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "prepareSettings", func(_ *NetworkConfigurator, _ *v1.Interface, _ nm.DeviceWired) (nm.ConnectionSettings, error) {
		return mockSettings, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "updateConnections", func(_ *NetworkConfigurator, _ nm.DeviceWired, _ nm.ConnectionSettings) error {
		return nil
	})

	patches.ApplyFunc(ConfigureExistingGatewayInterfacesExceptProtoData, func(_ *v1.Interface, _ NetworkConfigurator) error {
		return errors.New("ConfigureExistingGatewayInterfacesExceptProtoData error")
	})

	backup, err := nc.applyAndBackupSettings(protoData)

	assert.Equal(t, expectedBackup, backup, "applyAndBackupSettings should return the correct backup when ConfigureExistingGatewayInterfacesExceptProtoData fails")
	assert.NotNil(t, err, "applyAndBackupSettings should return an error when ConfigureExistingGatewayInterfacesExceptProtoData fails")
	assert.Equal(t, "ConfigureExistingGatewayInterfacesExceptProtoData error", err.Error(), "applyAndBackupSettings should return the correct error message")
}

func Test_ApplyAndBackupSettings_Success(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}
	mockDevice := &mockgnm.MockDeviceWired{}
	expectedBackup := nm.ConnectionSettings{}
	mockSettings := nm.ConnectionSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceBy", func(_ *NetworkConfigurator, _ *v1.Interface) (nm.DeviceWired, error) {
		return mockDevice, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "createBackupFromExisting", func(_ *NetworkConfigurator, _ nm.DeviceWired) nm.ConnectionSettings {
		return expectedBackup
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "prepareSettings", func(_ *NetworkConfigurator, _ *v1.Interface, _ nm.DeviceWired) (nm.ConnectionSettings, error) {
		return mockSettings, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "updateConnections", func(_ *NetworkConfigurator, _ nm.DeviceWired, _ nm.ConnectionSettings) error {
		return nil
	})

	patches.ApplyFunc(ConfigureExistingGatewayInterfacesExceptProtoData, func(_ *v1.Interface, _ NetworkConfigurator) error {
		return nil
	})

	backup, err := nc.applyAndBackupSettings(protoData)

	assert.Equal(t, expectedBackup, backup, "applyAndBackupSettings should return the correct backup on success")
	assert.Nil(t, err, "applyAndBackupSettings should not return an error on success")
}

func Test_getDeviceBy_WithMacAddress(t *testing.T) {
	nc := &NetworkConfigurator{}
	testMac := "00:0A:95:9D:68:16"
	protoData := &v1.Interface{MacAddress: testMac}

	expectedDevice := &mockgnm.MockDeviceWired{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return expectedDevice
	})

	device, err := nc.getDeviceBy(protoData)
	assert.NoError(t, err)
	assert.Equal(t, expectedDevice, device)
}

func Test_getDeviceBy_WithLabel(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{Label: "eth0"}
	expectedDevice := &mockgnm.MockDeviceWired{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithLabel", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return expectedDevice
	})

	device, err := nc.getDeviceBy(protoData)
	assert.NoError(t, err)
	assert.Equal(t, expectedDevice, device)
}

func Test_getDeviceBy_NoMacAddressOrLabel(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}

	device, err := nc.getDeviceBy(protoData)

	assert.Error(t, err, "Expected error when neither MacAddress nor Label is provided")
	assert.Nil(t, device, "Expected no device to be returned when neither MacAddress nor Label is provided")
	assert.Equal(t, "error, Mac address or Interface name should be entered", err.Error())
}

func Test_UpdateConnections_Success(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConnections := []nm.Connection{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockDevice), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "00:0a:95:9d:68:16", nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, _ string, _ nm.ConnectionSettings) error {
		return nil
	})

	err := nc.updateConnections(mockDevice, mockSettings)

	assert.Nil(t, err, "updateConnections should not return an error on success")
}

func Test_UpdateConnections_DeleteOldConnectionsFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConnections := []nm.Connection{}
	expectedError := errors.New("deleteOldConnections error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return expectedError
	})

	err := nc.updateConnections(mockDevice, mockSettings)

	assert.NotNil(t, err, "updateConnections should return an error when deleteOldConnections fails")
	assert.Equal(t, expectedError, err, "updateConnections should return the correct error message")
}

func Test_UpdateConnections_GetPropertyHwAddressFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConnections := []nm.Connection{}
	expectedError := errors.New("GetPropertyHwAddress error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockDevice), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "", expectedError
	})

	err := nc.updateConnections(mockDevice, mockSettings)

	assert.NotNil(t, err, "updateConnections should return an error when GetPropertyHwAddress fails")
	assert.Equal(t, expectedError, err, "updateConnections should return the correct error message")
}

func Test_UpdateConnections_AddConnectionFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConnections := []nm.Connection{}
	expectedError := errors.New("addConnection error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockDevice), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "00:0a:95:9d:68:16", nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, _ string, _ nm.ConnectionSettings) error {
		return expectedError
	})

	err := nc.updateConnections(mockDevice, mockSettings)

	assert.NotNil(t, err, "updateConnections should return an error when addConnection fails")
	assert.Equal(t, expectedError, err, "updateConnections should return the correct error message")
}

func TestPrepareSettings_Success(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}
	mockDevice := &mockgnm.MockDeviceWired{}
	expectedSettings := nm.ConnectionSettings{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(mockDevice, "GetPropertyInterface", "eth0", nil)
	patches.ApplyFuncReturn(newSettingsFromProto, expectedSettings)

	settings, err := nc.prepareSettings(protoData, mockDevice)

	assert.NoError(t, err, "Expected no error when GetPropertyInterface succeeds")
	assert.Equal(t, expectedSettings, settings, "Expected settings to match the expected settings")
}

func TestPrepareSettings_GetPropertyInterfaceError(t *testing.T) {
	nc := &NetworkConfigurator{}
	protoData := &v1.Interface{}
	mockDevice := &mockgnm.MockDeviceWired{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethodReturn(mockDevice, "GetPropertyInterface", "", errors.New("error from GetPropertyInterface"))

	settings, err := nc.prepareSettings(protoData, mockDevice)

	assert.Error(t, err, "Expected error when GetPropertyInterface fails")
	assert.Nil(t, settings, "Expected settings to be nil when GetPropertyInterface fails")
}

func Test_SetMACAddressInBackup_SetsMACAddress(t *testing.T) {
	// Create a mock DeviceWired
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	testMac := "00:0A:95:9D:68:16"
	parsedMac, _ := net.ParseMAC(testMac)

	// Set up the mock to return the test MAC address
	mockDeviceWired.On("GetPropertyPermHwAddress").Return(testMac, nil)

	// Test case where MAC address is not present in the backup
	backup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: nil,
		},
	}

	err := setMACAddressInBackup(backup, mockDeviceWired)

	// Assert that the MAC address was set in the backup
	assert.Equal(t, []uint8(parsedMac), backup[EthernetType][MACAddressKey], "MAC address should be set in the backup")
	assert.Nil(t, err, "setMACAddressInBackup should not return an error when MAC address is set in the backup")
}

func Test_SetMACAddressInBackup_DoesNotReturnErrorWhenBackupIsNotNil(t *testing.T) {

	backup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	err := setMACAddressInBackup(backup, nil)

	// Assert that the MAC address was set in the backup
	assert.Nil(t, err, "setMACAddressInBackup should not return an error when MAC address is set in the backup")
}

func Test_SetMACAddressInBackup_ReturnsErrorWhenParseMACFails(t *testing.T) {
	testMac := "00:0A:95:9D:68:16"
	expectedError := "invalid MAC address"
	mockDevice := &mockgnm.MockDeviceWired{}

	mockDevice.On("GetPropertyPermHwAddress").Return(testMac, nil)

	backup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: nil,
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(net.ParseMAC, func(s string) (net.HardwareAddr, error) {
		return nil, errors.New("invalid MAC address")
	})

	err := setMACAddressInBackup(backup, mockDevice)

	assert.NotNil(t, err, "setMACAddressInBackup should return an error when ParseMAC fails")
	assert.Equal(t, expectedError, err.Error(), "setMACAddressInBackup should return the correct error message")

}

func Test_CreateBackupFromExisting_ReturnsBackupWithValidConnection(t *testing.T) {
	nc := &NetworkConfigurator{}
	testDevice := new(mockgnm.MockDeviceWired)
	mockConnection := new(mockgnm.MockConnection)

	expectedBackup := nm.ConnectionSettings{
		"connection": {
			"id":   "Test Connection",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "802-3-ethernet",
		},
		EthernetType: map[string]interface{}{
			MACAddressKey: "F7:2B:A1:D5:97:4E",
		},
	}

	mockConnection.On("GetSettings").Return(nm.ConnectionSettings{
		"connection": {
			"id":             "Test Connection",
			"uuid":           "123e4567-e89b-12d3-a456-426614174000",
			"type":           "802-3-ethernet",
			"interface-name": nil,
			"timestamp":      1727244747362599705,
		},
		EthernetType: map[string]interface{}{
			MACAddressKey: "F7:2B:A1:D5:97:4E",
		},
		IPV4Key: nil,
	}, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock listConnections function to return a list of mock connections
	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return []nm.Connection{mockConnection}
	})

	patches.ApplyFunc(setMACAddressInBackup, func(backup nm.ConnectionSettings, device nm.DeviceWired) error {
		return errors.New("error from setMACAddressInBackup")
	})

	result := nc.createBackupFromExisting(testDevice)

	assert.Equal(t, expectedBackup["connection"]["id"], result["connection"]["id"], "ID should match")
	assert.Equal(t, expectedBackup["connection"]["type"], result["connection"]["type"], "Type should match")
	assert.Equal(t, expectedBackup["802-3-ethernet"]["mac-address"], result["802-3-ethernet"]["mac-address"], "MAC address should match")
}

func Test_CreateBackupFromExisting_ReturnsNilWhenNoConnections(t *testing.T) {
	nc := &NetworkConfigurator{}
	testDevice := new(mockgnm.MockDeviceWired)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock listConnections function to return an empty list
	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection { return []nm.Connection{} })

	result := nc.createBackupFromExisting(testDevice)

	assert.Nil(t, result, "createBackupFromExisting should return a nil ConnectionSettings instance when no connections are found")
}

func Test_AddConnection_SuccessfullyActivatesConnection(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	nc := &NetworkConfigurator{gnm: mockNetworkManager}
	testMac := "00:0a:95:9d:68:16"
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := &mockgnm.MockSettings{}
	mockConnection := &mockgnm.MockConnection{}
	mockActiveConnection := &mockgnm.MockActiveConnection{}

	testSettings := nm.ConnectionSettings{
		"connection": {
			"id":   "Test Connection",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "802-3-ethernet",
		},
	}

	mockSettings.On("AddConnection", testSettings).Return(mockConnection, nil).Once()
	mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, (*dbus.Object)(nil)).Return(mockActiveConnection, nil).Once()

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return mockDevice
	})

	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockConnection), "GetPath", func() dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/1"
	})

	err := nc.addConnection(testMac, testSettings)

	mockNetworkManager.AssertExpectations(t)
	mockConnection.AssertExpectations(t)
	assert.Nil(t, err, "addConnection should not return an error")
}

func Test_AddConnection_ReturnsNilWhenActivationFails(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	nc := &NetworkConfigurator{gnm: mockNetworkManager}
	testMac := "00:0a:95:9d:68:16"
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := &mockgnm.MockSettings{}
	mockConnection := &mockgnm.MockConnection{}
	mockActiveConnection := &mockgnm.MockActiveConnection{}

	testSettings := nm.ConnectionSettings{
		"connection": {
			"id":   "Test Connection",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "802-3-ethernet",
		},
	}

	mockSettings.On("AddConnection", testSettings).Return(mockConnection, nil).Once()
	mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, (*dbus.Object)(nil)).Return(mockActiveConnection, errors.New("ActivateConnection error")).Once()

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return mockDevice
	})

	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockConnection), "GetPath", func() dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/1"
	})

	err := nc.addConnection(testMac, testSettings)

	mockNetworkManager.AssertExpectations(t)
	mockConnection.AssertExpectations(t)
	assert.Nil(t, err, "addConnection should not return an error when ActivateConnection returns an error")
}

func Test_RestoreConnection_ReturnsNilWhenSuccessful(t *testing.T) {
	nc := &NetworkConfigurator{}
	backup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, mac string, settings nm.ConnectionSettings) error {
		return nil
	})

	err := nc.restoreConnection(backup)
	assert.Nil(t, err, "restoreConnection should not return an error when addConnection succeeds")
}

func Test_RestoreConnection_ReturnsErrorWhenAddConnectionFails(t *testing.T) {
	expectedError := "addConnection error"
	nc := &NetworkConfigurator{}
	backup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, mac string, settings nm.ConnectionSettings) error {
		return errors.New(expectedError)
	})

	err := nc.restoreConnection(backup)

	assert.NotNil(t, err, "restoreConnection should return an error when addConnection fails")
	assert.Equal(t, expectedError, err.Error(), "restoreConnection should return the correct error message")
}

func Test_RestoreConnection_CallCount(t *testing.T) {
	nc := &NetworkConfigurator{}
	backup := nm.ConnectionSettings{
		EthernetType: map[string]interface{}{
			MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	var addConnectionCalled bool
	var addConnectionMac string
	var addConnectionSettings nm.ConnectionSettings

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, mac string, settings nm.ConnectionSettings) error {
		addConnectionCalled = true
		addConnectionMac = mac
		addConnectionSettings = settings
		return nil
	})

	err := nc.restoreConnection(backup)

	// Assertions
	assert.Nil(t, err, "restoreConnection should not return an error")
	assert.True(t, addConnectionCalled, "addConnection should be called")
	assert.Equal(t, "00:0a:95:9d:68:16", addConnectionMac, "addConnection should be called with the correct MAC address")
	assert.Equal(t, backup, addConnectionSettings, "addConnection should be called with the correct settings")
}

func TestDeleteOldConnections_Failure(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockConn1 := &mockgnm.MockConnection{}
	mockConn2 := &mockgnm.MockConnection{}

	mockConn1.On("Delete").Return(nil)
	mockConn1.On("GetPath").Return(dbus.ObjectPath("/path/to/connection1"))

	mockConn2.On("Delete").Return(errors.New("delete error"))
	mockConn2.On("GetPath").Return(dbus.ObjectPath("/path/to/connection2"))

	connections := []nm.Connection{mockConn1, mockConn2}

	err := nc.deleteOldConnections(connections)
	assert.Error(t, err, "Expected error when a connection fails to delete")
	assert.Equal(t, "delete error", err.Error())

	mockConn1.AssertExpectations(t)
	mockConn2.AssertExpectations(t)
}

func Test_DeleteOldConnections_Success(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockConn1 := &mockgnm.MockConnection{}
	mockConn2 := &mockgnm.MockConnection{}

	// Set up the mock to return nil error on Delete
	mockConn1.On("Delete").Return(nil)
	mockConn1.On("GetPath").Return(dbus.ObjectPath("/path/to/connection1"))

	mockConn2.On("Delete").Return(nil)
	mockConn2.On("GetPath").Return(dbus.ObjectPath("/path/to/connection2"))

	connections := []nm.Connection{mockConn1, mockConn2}

	err := nc.deleteOldConnections(connections)

	// Assertions
	assert.Nil(t, err, "deleteOldConnections should not return an error")
	mockConn1.AssertNumberOfCalls(t, "Delete", 1)
	mockConn2.AssertNumberOfCalls(t, "Delete", 1)
}
