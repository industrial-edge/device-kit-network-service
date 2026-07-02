/*
 * Copyright © Siemens 2020 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	"fmt"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/mocks/gonetworkmanager"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"networkservice/internal/networking/mocks/networking"
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
	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceGatewayIDAndMetric", func(_ *NetworkConfigurator, device nm.Device) (string, uint8, error) {
		if callCount == 0 {
			callCount++
			return "11:22:33:44:55:66", 50, nil
		}
		return expectedMac, 10, nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(&mockgnm.MockDeviceWired{}), "GetPropertyHwAddress", func(device nm.DeviceWired) (string, error) {
		if callCount == 0 {
			callCount++
			return "11:22:33:44:55:66", nil
		}
		return expectedMac, nil
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

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceGatewayIDAndMetric", func(_ *NetworkConfigurator, device nm.DeviceWired) (string, uint8, error) {
		return "", 0, fmt.Errorf("no gateway found")
	})

	patches.ApplyMethod(reflect.TypeOf(&mockgnm.MockDeviceWired{}), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "00:0a:95:9d:68:16", nil
	})

	result := nc.findGatewayMAC(mockDevices)

	assert.Equal(t, "", result, "findGatewayMAC should return an empty string when no gateway MAC is found")
}
func Test_getDeviceGatewayIDAndMetric_ReturnsCorrectMACAndMetric(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDevice{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}
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

	// Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(nm.NewDeviceWired, func(_ dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})
	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "inspectRoutesForGateway", func(_ *NetworkConfigurator, device nm.DeviceWired) (string, uint8, error) {
		return "F7:2B:A1:D5:97:4E", uint8(10), nil
	})
	mockDevice.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))
	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockDeviceWired.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)
	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

	// Test the function
	mac, metric, err := nc.getDeviceGatewayIDAndMetric(mockDevice)

	// Assertions
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "F7:2B:A1:D5:97:4E", mac, "Expected correct MAC address")
	assert.Equal(t, uint8(10), metric, "Expected correct metric value")
}

func Test_getDeviceGatewayIDAndMetric_NewDeviceWiredError(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockActiveConnection{}
	mockIPv4Config := &mockgnm.MockIP4Config{}

	mockRouteData := []nm.IP4RouteData{
		{
			Destination: common.OutgoingRouteDestination,
			Prefix:      common.OutgoingRoutePrefix,
			Metric:      20,
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Force nm.NewDeviceWired to fail
	expectedErr := fmt.Errorf("failed to create wired device")
	patches.ApplyFunc(nm.NewDeviceWired, func(_ dbus.ObjectPath) (nm.DeviceWired, error) {
		return nil, expectedErr
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "inspectRoutesForGateway", func(_ *NetworkConfigurator, device nm.DeviceWired) (string, uint8, error) {
		return "", 0, expectedErr
	})

	mockDevice.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))
	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)
	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

	// Execute
	mac, metric, err := nc.getDeviceGatewayIDAndMetric(mockDevice)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", mac)
	assert.Equal(t, uint8(0), metric)
}

func Test_getDeviceGatewayIDAndMetric_GetHwAddressError(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDevice{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockConn := &mockgnm.MockActiveConnection{}
	mockIPv4Config := &mockgnm.MockIP4Config{}

	mockRouteData := []nm.IP4RouteData{
		{
			Destination: common.OutgoingRouteDestination,
			Prefix:      common.OutgoingRoutePrefix,
			Metric:      15,
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// nm.NewDeviceWired succeeds
	patches.ApplyFunc(nm.NewDeviceWired, func(_ dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})

	expectedErr := fmt.Errorf("failed to get MAC address")
	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "inspectRoutesForGateway", func(_ *NetworkConfigurator, device nm.DeviceWired) (string, uint8, error) {
		return "", 0, expectedErr
	})
	mockDeviceWired.On("GetPropertyHwAddress").Return("", expectedErr)

	mockDevice.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))
	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)
	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

	// Execute
	mac, metric, err := nc.getDeviceGatewayIDAndMetric(mockDevice)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", mac)
	assert.Equal(t, uint8(0), metric)
}

func Test_getDeviceGatewayIDAndMetric_ReturnsEmptyMACForModem(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDevice{}
	mockConn := &mockgnm.MockActiveConnection{}
	mockIPv4Config := &mockgnm.MockIP4Config{}

	mockRouteData := []nm.IP4RouteData{
		{
			Destination: "0.0.0.0",
			Prefix:      0,
			NextHop:     "10.0.0.1",
			Metric:      20,
		},
	}

	// Create patches
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Device is GSM / Modem
	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)

	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

	// Execute
	id, metric, err := nc.getDeviceGatewayIDAndMetric(mockDevice)

	// Assertions
	assert.NoError(t, err, "Expected no error for GSM device")
	assert.Equal(t, "", id, "Expected empty MAC for GSM device")
	assert.Equal(t, uint8(20), metric, "Expected correct metric value")
}

func Test_getDeviceGatewayIDAndMetric_DeviceIsNil(t *testing.T) {
	nc := &NetworkConfigurator{}

	mac, metric, err := nc.getDeviceGatewayIDAndMetric(nil)

	assert.Error(t, err, "Expected an error")
	assert.Equal(t, "", mac, "Expected empty MAC address")
	assert.Equal(t, uint8(0), metric, "Expected metric value to be 0")
	assert.Equal(t, "device is nil", err.Error(), "Expected error message to be 'device is nil'")
}

type mockIP4Config struct {
	nm.IP4Config
}

func Test_getDeviceGatewayIDAndMetric_IPv4WrapperIsNil(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	mockConn := &mockgnm.MockActiveConnection{}
	anyHwAddress := "F7:2B:A1:D5:97:4E"

	mockDevice.On("GetPropertyHwAddress").Return(anyHwAddress, nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)
	mockConn.On("GetPropertyIP4Config").Return((*mockIP4Config)(nil), nil)

	mac, metric, err := nc.getDeviceGatewayIDAndMetric(mockDevice)

	assert.Error(t, err, "Expected an error")
	assert.Equal(t, "", mac, "Expected empty MAC address")
	assert.Equal(t, uint8(0), metric, "Expected metric value to be 0")
	assert.Equal(t, "failed to get IPv4 configuration", err.Error(), "Expected error message to be 'failed to get IPv4 configuration'")
}

func Test_getDeviceGatewayIDAndMetric_RouteDataIsNil(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDeviceWired{}
	mockConn := &mockgnm.MockActiveConnection{}
	mockIPv4Config := &mockgnm.MockIP4Config{}
	anyHwAddress := "F7:2B:A1:D5:97:4E"

	mockDevice.On("GetPropertyHwAddress").Return(anyHwAddress, nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)
	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return([]nm.IP4RouteData{}, fmt.Errorf("failed to get route data"))

	mac, metric, err := nc.getDeviceGatewayIDAndMetric(mockDevice)

	assert.Error(t, err, "Expected an error")
	assert.Equal(t, "", mac, "Expected empty MAC address")
	assert.Equal(t, uint8(0), metric, "Expected metric value to be 0")
	assert.Equal(t, "failed to get route data", err.Error(), "Expected error message to be 'failed to get route data'")
}

func Test_GetInterfaceWithMac_ReturnsCorrectInterface(t *testing.T) {
	var actualDeviceType reflect.Type
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDevice{}
	testMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.Device {
		return mockDevice
	})

	patches.ApplyFunc(DBusToProto, func(device nm.Device) *v1.Interface {
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
	mockDevice := &mockgnm.MockDevice{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithLabel",
		func(_ *NetworkConfigurator, label string) nm.Device { return mockDevice })

	patches.ApplyFunc(DBusToProto, func(device nm.Device) *v1.Interface {
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
		func(_ *NetworkConfigurator, label string) nm.Device { return nil })

	result := nc.GetInterfaceWithLabel(testLabel)

	assert.Nil(t, result, "GetInterfaceWithLabel should return a nil Interface instance when device is not found")
}

func Test_GetEthernetInterfaces_ReturnsAllAvailableNetworkDevices(t *testing.T) {
	var actualDeviceType reflect.Type
	nc := &NetworkConfigurator{}
	mockDevice := &mockgnm.MockDevice{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}

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

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(nm.NewDeviceWired, func(_ dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})
	mockDevice.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))
	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)

	mockDevice.On("GetPropertyInterface").Return("eth0", nil)
	mockDeviceWired.On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)
	mockDevice.On("GetPropertyActiveConnection").Return(mockConn, nil)

	mockConn.On("GetPropertyIP4Config").Return(mockIPv4Config, nil)
	mockIPv4Config.On("GetPropertyRouteData").Return(mockRouteData, nil)

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllNetworkDevices", func(_ *NetworkConfigurator) []nm.Device {
		return []nm.Device{mockDevice, mockDevice}
	})

	patches.ApplyFunc(DBusToProto, func(device nm.Device) *v1.Interface {
		actualDeviceType = reflect.TypeOf(device)
		return &v1.Interface{
			InterfaceName: "eth0",
			MacAddress:    "00:0a:95:9d:68:16",
		}
	})

	result := nc.GetNetworkInterfaces()

	assert.Equal(t, reflect.TypeOf(mockDevice), actualDeviceType, "Expected device type should be the same as mockDevice")
	assert.NotNil(t, result, "GetEthernetInterfaces should return non-nil Interface instances")
	assert.Equal(t, 2, len(result), "GetEthernetInterfaces should return the correct number of interfaces")
	assert.Equal(t, "eth0", result[0].InterfaceName, "First interface should have the correct name")
	assert.Equal(t, "eth0", result[1].InterfaceName, "Second interface should have the correct name")
}

func Test_GetEthernetInterfaces_EnsuresGatewayInterfaceIsMarked(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDevice1 := &mockgnm.MockDevice{}
	mockDevice2 := &mockgnm.MockDevice{}

	mockDevicesWired := make([]mockgnm.MockDeviceWired, 2)

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

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(nm.NewDeviceWired, func(objectPath dbus.ObjectPath) (nm.DeviceWired, error) {
		switch objectPath {
		case "mockEno0Path":
			return &mockDevicesWired[0], nil
		case "mockEno1Path":
			return &mockDevicesWired[1], nil
		}
		panic("invalid object path")
	})

	mockDevice1.On("GetPath").Return(dbus.ObjectPath("mockEno0Path"))

	// Mock behaviors for Device 1
	mockDevice1.On("GetPropertyInterface").Return("eth0", nil)
	mockDevice1.On("GetPropertyActiveConnection").Return(mockConn1, nil)
	mockConn1.On("GetPropertyIP4Config").Return(mockIPv4Config1, nil)
	mockIPv4Config1.On("GetPropertyRouteData").Return(mockRouteData1, nil)
	mockDevice1.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)

	// Mock behaviors for Device 2
	mockDevice2.On("GetPropertyInterface").Return("eth1", nil)
	mockDevice2.On("GetPropertyActiveConnection").Return(mockConn2, nil)
	mockConn2.On("GetPropertyIP4Config").Return(mockIPv4Config2, nil)
	mockIPv4Config2.On("GetPropertyRouteData").Return(mockRouteData2, nil)
	mockDevice2.On("GetPath").Return(dbus.ObjectPath("mockEno1Path"))
	mockDevice2.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)

	mockDevicesWired[0].On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4E", nil)
	mockDevicesWired[1].On("GetPropertyHwAddress").Return("F7:2B:A1:D5:97:4", nil)

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllNetworkDevices", func(_ *NetworkConfigurator) []nm.Device {
		return []nm.Device{mockDevice1, mockDevice2}
	})

	patches.ApplyFunc(DBusToProto, func(device nm.Device) *v1.Interface {
		name, _ := device.GetPropertyInterface()
		return &v1.Interface{
			InterfaceName: name,
			MacAddress:    "00:0a:95:9d:68:16",
		}
	})

	result := nc.GetNetworkInterfaces()

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

func Test_GetDeviceWithMac_ReturnsNilWhenNoMatchingDevice(t *testing.T) {
	mockNetworkManager := new(mockgnm.MockNetworkManager)
	nc := &NetworkConfigurator{
		gnm: mockNetworkManager,
	}
	testMac := "00:0a:95:9d:68:16"

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock getAllEthernetDevices method to return an empty list

	mockNetworkManager.On("GetDevices").Return([]nm.Device{}, nil)

	result := nc.getDeviceWithMac(testMac)

	assert.Nil(t, result, "getDeviceWithMac should return a nil DeviceWired instance when no device with the matching MAC address is found")
}

func Test_GetDeviceWithLabel_ReturnsDeviceWithCorrectLabel(t *testing.T) {
	testLabel := "eth0"
	otherLabel := "eth1"
	mockNetworkManager := new(mockgnm.MockNetworkManager)
	nc := &NetworkConfigurator{
		gnm: mockNetworkManager,
	}

	mockDevice := &mockgnm.MockDeviceWired{}
	otherMockDevice := &mockgnm.MockDeviceWired{}

	mockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	otherMockDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)

	mockDevice.On("GetPropertyInterface").Return(testLabel, nil)
	otherMockDevice.On("GetPropertyInterface").Return(otherLabel, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	mockNetworkManager.On("GetDevices").Return([]nm.Device{mockDevice, otherMockDevice}, nil)

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return []nm.DeviceWired{mockDevice}
	})

	patches.ApplyFunc(getInterfaceForLabel, func(_ string) string { return testLabel })

	mockDevice.On("GetPath").Return(dbus.ObjectPath("/path/to/connection1"))

	result := nc.getDeviceWithLabel(testLabel)
	interfaceName, _ := result.GetPropertyInterface()

	assert.NotNil(t, result, "getDeviceWithLabel should return a non-nil DeviceWired instance")
	assert.Equal(t, testLabel, interfaceName, "getDeviceWithLabel should return a DeviceWired with the correct label")
	assert.Equal(t, mockDevice, result, "getDeviceWithLabel should return the correct DeviceWired instance")
}

func Test_GetDeviceWithLabel_ReturnsNilWhenNoMatchingDevice(t *testing.T) {
	testLabel := "eth0"
	mockNetworkManager := new(mockgnm.MockNetworkManager)
	nc := &NetworkConfigurator{
		gnm: mockNetworkManager,
	}

	mockDevice1 := &mockgnm.MockDeviceWired{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	mockDevice1.On("GetPropertyInterface").Return("eth1", nil)
	mockDevice2.On("GetPropertyInterface").Return("eth2", nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	mockDevice1.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockDevice2.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)

	mockNetworkManager.On("GetDevices").Return([]nm.Device{mockDevice1, mockDevice2}, nil)

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getAllEthernetDevices", func(_ *NetworkConfigurator) []nm.DeviceWired {
		return []nm.DeviceWired{mockDevice1, mockDevice2}
	})

	patches.ApplyFunc(getInterfaceForLabel, func(_ string) string { return testLabel })

	mockDevice1.On("GetPath").Return(dbus.ObjectPath("/path/to/connection1"))
	mockDevice2.On("GetPath").Return(dbus.ObjectPath("/path/to/connection2"))

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
			return &mockgnm.MockDeviceWired{}, nil
		}
		return &mockgnm.MockDeviceWired{}, nil
	})

	// Call the function
	result := nc.getAllEthernetDevices()

	// Assertions
	assert.Equal(t, 2, len(result), "Expected two Ethernet devices")
	assert.Equal(t, mockDeviceWiredForEthernet1, result[0], "Expected the first mock Ethernet device")
	assert.Equal(t, mockDeviceWiredForEthernet2, result[1], "Expected the second mock Ethernet device")
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

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithLabel", func(_ *NetworkConfigurator, mac string) nm.Device {
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
	mockDeviceWired := &mockgnm.MockDeviceWired{}
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

	patches.ApplyMethod(reflect.TypeOf(mockDeviceWired), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "00:0a:95:9d:68:16", nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, _ string, _ nm.ConnectionSettings) error {
		return nil
	})

	err := nc.updateConnections(mockDeviceWired, mockSettings)

	assert.Nil(t, err, "updateConnections should not return an error on success")
}

func Test_UpdateConnections_DeleteOldConnectionsFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConn1 := &mockgnm.MockConnection{}
	mockConn2 := &mockgnm.MockConnection{}
	mockConnections := []nm.Connection{mockConn1, mockConn2}
	expectedError := errors.New("deleteOldConnections error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return expectedError
	})

	err := nc.updateConnections(mockDeviceWired, mockSettings)

	assert.NotNil(t, err, "updateConnections should return an error when deleteOldConnections fails")
	assert.Equal(t, expectedError, err, "updateConnections should return the correct error message")
}

func Test_UpdateConnections_GetPropertyHwAddressFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConn1 := &mockgnm.MockConnection{}
	mockConn2 := &mockgnm.MockConnection{}
	mockConnections := []nm.Connection{mockConn1, mockConn2}
	expectedError := errors.New("GetPropertyHwAddress error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockDeviceWired), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "", expectedError
	})

	err := nc.updateConnections(mockDeviceWired, mockSettings)

	assert.NotNil(t, err, "updateConnections should return an error when GetPropertyHwAddress fails")
	assert.Equal(t, expectedError, err, "updateConnections should return the correct error message")
}

func Test_UpdateConnections_AddConnectionFails(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}
	mockSettings := nm.ConnectionSettings{}
	mockConn1 := &mockgnm.MockConnection{}
	mockConn2 := &mockgnm.MockConnection{}
	mockConnections := []nm.Connection{mockConn1, mockConn2}
	expectedError := errors.New("addConnection error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(listConnections, func(_ nm.Device) []nm.Connection {
		return mockConnections
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "deleteOldConnections", func(_ *NetworkConfigurator, _ []nm.Connection) error {
		return nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockDeviceWired), "GetPropertyHwAddress", func(_ nm.DeviceWired) (string, error) {
		return "00:0a:95:9d:68:16", nil
	})

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, _ string, _ nm.ConnectionSettings) error {
		return expectedError
	})

	err := nc.updateConnections(mockDeviceWired, mockSettings)

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

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetPropertyDeviceType to return Ethernet for both devices
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWired), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})

	// Test case where MAC address is not present in the backup
	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: nil,
		},
	}

	err := setMACAddressInBackup(backup, mockDeviceWired)

	// Assert that the MAC address was set in the backup
	assert.Equal(t, []uint8(parsedMac), backup[common.EthernetType][common.MACAddressKey], "MAC address should be set in the backup")
	assert.Nil(t, err, "setMACAddressInBackup should not return an error when MAC address is set in the backup")
}

func Test_SetMACAddressInBackup_SkipsForModemDevice(t *testing.T) {
	// Create a mock generic device (not DeviceWired)
	mockDeviceWired := &mockgnm.MockDeviceWired{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock device type as Modem (GSM)
	mockDeviceWired.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil)

	// Prepare backup with empty ethernet section
	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: "nil",
		},
	}

	// Execute
	err := setMACAddressInBackup(backup, mockDeviceWired)

	// Assertions
	assert.NoError(t, err, "Expected no error for GSM device")
	assert.Equal(t, "nil", backup[common.EthernetType][common.MACAddressKey],
		"nil")
}

func Test_SetMACAddressInBackup_DoesNotReturnErrorWhenBackupIsNotNil(t *testing.T) {
	mockDevice := &mockgnm.MockDeviceWired{}
	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeOf(mockDevice), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})
	mockDevice.On("GetPropertyPermHwAddress").Return("testMac", nil)
	err := setMACAddressInBackup(backup, mockDevice)

	// Assert that the MAC address was set in the backup
	assert.Nil(t, err, "setMACAddressInBackup should not return an error when MAC address is set in the backup")
}

func Test_SetMACAddressInBackup_ReturnsErrorWhenParseMACFails(t *testing.T) {
	testMac := "00:0A:95:9D:68:16"
	expectedError := "invalid MAC address"
	mockDevice := &mockgnm.MockDeviceWired{}

	mockDevice.On("GetPropertyPermHwAddress").Return(testMac, nil)

	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: nil,
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(net.ParseMAC, func(s string) (net.HardwareAddr, error) {
		return nil, errors.New("invalid MAC address")
	})

	// Mock GetPropertyDeviceType to return Ethernet for both devices
	patches.ApplyMethod(reflect.TypeOf(mockDevice), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})

	err := setMACAddressInBackup(backup, mockDevice)

	assert.NotNil(t, err, "setMACAddressInBackup should return an error when ParseMAC fails")
	assert.Equal(t, expectedError, err.Error(), "setMACAddressInBackup should return the correct error message")

}

func Test_AddConnection_SuccessfullyActivatesConnection(t *testing.T) {
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	nc := &NetworkConfigurator{gnm: mockNetworkManager}
	testMac := "00:0a:95:9d:68:16"
	mockDevice := &mockgnm.MockDeviceWired{}
	mockSettings := &mockgnm.MockSettings{}
	mockConnection := &mockgnm.MockConnection{}
	// mockActiveConnection := &mockgnm.MockActiveConnection{}

	testSettings := nm.ConnectionSettings{
		"connection": {
			"id":   "Test Connection",
			"uuid": "123e4567-e89b-12d3-a456-426614174000",
			"type": "802-3-ethernet",
		},
	}

	// mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, (*dbus.Object)(nil)).Return(mockActiveConnection, nil).Once()

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	mockSettings.On("AddConnection", testSettings).Return(mockConnection, nil).Once()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "getDeviceWithMac", func(_ *NetworkConfigurator, mac string) nm.DeviceWired {
		return mockDevice
	})

	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockConnection), "GetPath", func() dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/1"
	})

	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "ActivateConnection", func(_ nm.NetworkManager, connection nm.Connection, device nm.Device, specificObject *dbus.Object) (nm.ActiveConnection, error) {
		return nil, errors.New("could not activate connection")
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
	// mockNetworkManager.On("ActivateConnection", mockConnection, mockDevice, (*dbus.Object)(nil)).Return(mockActiveConnection, errors.New("ActivateConnection error")).Once()

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
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "ActivateConnection", func(_ nm.NetworkManager, connection nm.Connection, device nm.Device, specificObject *dbus.Object) (nm.ActiveConnection, error) {
		return mockActiveConnection, nil
	})

	err := nc.addConnection(testMac, testSettings)

	mockNetworkManager.AssertExpectations(t)
	mockConnection.AssertExpectations(t)
	assert.Nil(t, err, "addConnection should not return an error when ActivateConnection returns an error")
}

func Test_RestoreConnection_ReturnsNilWhenSuccessful(t *testing.T) {
	nc := &NetworkConfigurator{}
	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, mac string, settings nm.ConnectionSettings) error {
		return nil
	})

	err := nc.restoreEthernetConnection(backup)
	assert.Nil(t, err, "restoreEthernetConnection should not return an error when addConnection succeeds")
}

func Test_RestoreConnection_ReturnsErrorWhenAddConnectionFails(t *testing.T) {
	expectedError := "addConnection error"
	nc := &NetworkConfigurator{}
	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
		},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyPrivateMethod(reflect.TypeOf(nc), "addConnection", func(_ *NetworkConfigurator, mac string, settings nm.ConnectionSettings) error {
		return errors.New(expectedError)
	})

	err := nc.restoreEthernetConnection(backup)

	assert.NotNil(t, err, "restoreEthernetConnection should return an error when addConnection fails")
	assert.Equal(t, expectedError, err.Error(), "restoreEthernetConnection should return the correct error message")
}

func Test_RestoreConnection_CallCount(t *testing.T) {
	nc := &NetworkConfigurator{}
	backup := nm.ConnectionSettings{
		common.EthernetType: map[string]interface{}{
			common.MACAddressKey: []byte{0x00, 0x0a, 0x95, 0x9d, 0x68, 0x16},
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

	err := nc.restoreEthernetConnection(backup)

	// Assertions
	assert.Nil(t, err, "restoreEthernetConnection should not return an error")
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

func Test_CreateConnectionGSM_WhenAddGSMconnection_Succeeds(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	newSetting := &v1.ConnectionSettings{Name: "testgsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	mockDevice1 := &mockgnm.MockDeviceWireless{}
	mockDevice2 := &mockgnm.MockDeviceWired{}

	nc.gnm = mockNetworkManager
	mockSettings := &mockgnm.MockSettings{}
	mockConnection := &mockgnm.MockConnection{}
	mockConn1 := &mockgnm.MockConnection{}
	connections := []nm.Connection{mockConn1}
	mockActiveConnection := &mockgnm.MockActiveConnection{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list of devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockDevice1, mockDevice2}, nil
	})
	// Mock GetPropertyDeviceType to return Ethernet for both devices
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeModem, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice2), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPath", func(_ nm.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/2"
	})
	patches.ApplyMethod(reflect.TypeOf(mockDevice1), "GetPropertyInterface", func(_ nm.Device) (string, error) {
		return "interfacetest", nil
	})
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockConnection), "GetPath", func(_ nm.Connection) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Connection/2"
	})
	patches.ApplyMethod(reflect.TypeOf(mockSettings), "ListConnections", func(_ nm.Settings) ([]nm.Connection, error) {
		return connections, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockConn1), "GetSettings", func(_ nm.Connection) (nm.ConnectionSettings, error) {
		connection := nm.ConnectionSettings{
			"connection": map[string]interface{}{
				"id":   "testgsm_dhcp",
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"type": "gsm",
			},
		}
		return connection, nil

	})

	patches.ApplyMethod(reflect.TypeOf(mockConn1), "Delete", func(_ nm.Connection) error {
		return nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockSettings), "AddConnection", func(_ nm.Settings, settings nm.ConnectionSettings) (nm.Connection, error) {
		return mockConnection, nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "ActivateConnection", func(_ nm.NetworkManager, connection nm.Connection, device nm.Device, specificObject *dbus.Object) (nm.ActiveConnection, error) {
		return mockActiveConnection, nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockActiveConnection), "GetPropertyState", func(_ nm.ActiveConnection) (nm.NmActiveConnectionState, error) {
		return nm.NmActiveConnectionStateActivated, nil
	})
	err := nc.CreateConnection(newSetting)
	assert.Nil(t, err, nil)
}

func Test_RemoveConnection_SuccessInDeleteConnection(t *testing.T) {

	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}

	newSetting := &v1.ConnectionSettings{Name: "testgsm", ConnectionType: v1.ConnectionSettings_GSM,
		ConnectionConfig: &v1.ConnectionSettings_Gsm{Gsm: &v1.Interface_GsmConf{Apn: "my-internet", Pin: "1234"}}}

	mockConn1 := &mockgnm.MockConnection{}
	mockConnections := []nm.Connection{mockConn1}

	nc.gnm = mockNetworkManager
	mockSettings := &mockgnm.MockSettings{}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(nm.NewSettings, func() (nm.Settings, error) {
		return mockSettings, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockSettings), "ListConnections", func(_ nm.Settings) ([]nm.Connection, error) {
		return mockConnections, nil
	})
	patches.ApplyMethod(reflect.TypeOf(mockConn1), "GetSettings", func(_ nm.Connection) (nm.ConnectionSettings, error) {
		connection := nm.ConnectionSettings{
			"connection": map[string]interface{}{
				"id":   "testgsm_dhcp",
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"type": "gsm",
			},
		}
		return connection, nil

	})
	patches.ApplyMethod(reflect.TypeOf(mockConn1), "Delete", func(_ nm.Connection) error {
		return nil
	})
	err := nc.RemoveConnection(newSetting)
	assert.Nil(t, err, nil)
}

func Test_getAllNetworkDevices_EthernetAndModemDevices(t *testing.T) {
	nc := &NetworkConfigurator{}

	mockGnm := &mockgnm.MockNetworkManager{}
	mockEthernetDevice := &mockgnm.MockDevice{}
	mockModemDevice := &mockgnm.MockDevice{}
	mockWiredDevice := &mockgnm.MockDeviceWired{}
	mockGenericModem := &mockgnm.MockDevice{}

	nc.gnm = mockGnm

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// gnm.GetDevices returns Ethernet + Modem
	mockGnm.On("GetDevices").Return([]nm.Device{
		mockEthernetDevice,
		mockModemDevice,
	}, nil)

	// Device types
	mockEthernetDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeEthernet, nil)
	mockModemDevice.On("GetPropertyDeviceType").Return(nm.NmDeviceTypeModem, nil)

	// Device paths
	mockEthernetDevice.On("GetPath").Return(dbus.ObjectPath("ethernet-path"))
	mockModemDevice.On("GetPath").Return(dbus.ObjectPath("modem-path"))

	// Patch constructors
	patches.ApplyFunc(nm.NewDeviceWired, func(_ dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockWiredDevice, nil
	})

	patches.ApplyFunc(nm.NewDevice, func(_ dbus.ObjectPath) (nm.Device, error) {
		return mockGenericModem, nil
	})

	patches.ApplyMethod(reflect.TypeOf(mockGnm), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockEthernetDevice, mockModemDevice}, nil
	})

	// Execute
	result := nc.getAllNetworkDevices()

	// Assertions
	assert.Len(t, result, 2)
}

func TestNetworkConfigurator_Apply_Success(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	mockEthernetConfigurator := new(networking.MockEthernetConfigurator)
	mockGSMConfigurator := new(networking.MockGSMConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*NetworkConfigurator](), "GetNetworkInterfaces",
		func() []*v1.Interface {
			return existingIfaces
		})

	mockInterfaceStateManager := new(networking.MockInterfaceStateManager)

	patches.ApplyFunc(NewInterfaceStateHandler,
		func(_ nm.NetworkManager, _ []*v1.Interface) InterfaceStateManager {
			return mockInterfaceStateManager
		})

	mockInterfaceStateManager.On("Backup").Return(nil).Once()

	patches.ApplyFunc(NewInterfaceConfigurator,
		func(nc *NetworkConfigurator, iface *v1.Interface) (InterfaceConfigurator, error) {
			assert.Equal(t, nc, nc, "NetworkConfigurator instance should be passed to NewConnectionHandler")
			switch iface.InterfaceName {
			case newIfaceOne.InterfaceName:
				return mockEthernetConfigurator, nil
			default:
				return mockGSMConfigurator, nil
			}
		})

	mockEthernetConfigurator.On("Configure").Return(nil).Once()
	mockGSMConfigurator.On("Configure").Return(nil).Once()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "resetGateway",
		func(_, _ []*v1.Interface) error {
			return nil
		})

	// Act
	err := nc.Apply(&v1.NetworkSettings{Interfaces: newIfaces})

	// Assertions
	assert.NoError(t, err, "Apply should not return an error")

	mockEthernetConfigurator.AssertExpectations(t)
	mockGSMConfigurator.AssertExpectations(t)
	mockInterfaceStateManager.AssertExpectations(t)
}

func TestNetworkConfigurator_Apply_ResetGateway_Error(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	mockEthernetConfigurator := new(networking.MockEthernetConfigurator)
	mockGSMConfigurator := new(networking.MockGSMConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*NetworkConfigurator](), "GetNetworkInterfaces",
		func() []*v1.Interface {
			return existingIfaces
		})

	mockInterfaceStateManager := new(networking.MockInterfaceStateManager)

	patches.ApplyFunc(NewInterfaceStateHandler,
		func(_ nm.NetworkManager, _ []*v1.Interface) InterfaceStateManager {
			return mockInterfaceStateManager
		})

	mockInterfaceStateManager.On("Backup").Return(nil).Once()

	patches.ApplyFunc(NewInterfaceConfigurator,
		func(nc *NetworkConfigurator, iface *v1.Interface) (InterfaceConfigurator, error) {
			assert.Equal(t, nc, nc, "NetworkConfigurator instance should be passed to NewConnectionHandler")
			switch iface.InterfaceName {
			case newIfaceOne.InterfaceName:
				return mockEthernetConfigurator, nil
			default:
				return mockGSMConfigurator, nil
			}
		})

	mockEthernetConfigurator.On("Configure").Return(nil).Once()
	mockGSMConfigurator.On("Configure").Return(nil).Once()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "resetGateway",
		func(_, _ []*v1.Interface) error {
			return fmt.Errorf("mock error")
		})

	mockInterfaceStateManager.On("Restore").Return(nil).Once()

	// Act
	err := nc.Apply(&v1.NetworkSettings{Interfaces: newIfaces})

	// Assertions
	assert.Error(t, err, "Apply should return an error")

	mockEthernetConfigurator.AssertExpectations(t)
	mockGSMConfigurator.AssertExpectations(t)
	mockInterfaceStateManager.AssertExpectations(t)
}

func TestNetworkConfigurator_Apply_Configure_Error(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	mockEthernetConfigurator := new(networking.MockEthernetConfigurator)
	mockGSMConfigurator := new(networking.MockGSMConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*NetworkConfigurator](), "GetNetworkInterfaces",
		func() []*v1.Interface {
			return existingIfaces
		})

	mockInterfaceStateManager := new(networking.MockInterfaceStateManager)

	patches.ApplyFunc(NewInterfaceStateHandler,
		func(_ nm.NetworkManager, _ []*v1.Interface) InterfaceStateManager {
			return mockInterfaceStateManager
		})

	mockInterfaceStateManager.On("Backup").Return(nil).Once()

	patches.ApplyFunc(NewInterfaceConfigurator,
		func(nc *NetworkConfigurator, iface *v1.Interface) (InterfaceConfigurator, error) {
			assert.Equal(t, nc, nc, "NetworkConfigurator instance should be passed to NewConnectionHandler")
			switch iface.InterfaceName {
			case newIfaceOne.InterfaceName:
				return mockEthernetConfigurator, nil
			default:
				return mockGSMConfigurator, nil
			}
		})

	mockEthernetConfigurator.On("Configure").Return(nil).Once()
	mockGSMConfigurator.On("Configure").Return(errors.New("mock error")).Once()

	mockInterfaceStateManager.On("Restore").Return(nil).Once()

	// Act
	err := nc.Apply(&v1.NetworkSettings{Interfaces: newIfaces})

	// Assertions
	assert.Error(t, err, "Apply should return an error")

	mockEthernetConfigurator.AssertExpectations(t)
	mockGSMConfigurator.AssertExpectations(t)
}

func TestNetworkConfigurator_Apply_NewInterfaceConfigurator_Error(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	mockEthernetConfigurator := new(networking.MockEthernetConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*NetworkConfigurator](), "GetNetworkInterfaces",
		func() []*v1.Interface {
			return existingIfaces
		})

	mockInterfaceStateManager := new(networking.MockInterfaceStateManager)

	patches.ApplyFunc(NewInterfaceStateHandler,
		func(_ nm.NetworkManager, _ []*v1.Interface) InterfaceStateManager {
			return mockInterfaceStateManager
		})

	mockInterfaceStateManager.On("Backup").Return(nil).Once()

	patches.ApplyFunc(NewInterfaceConfigurator,
		func(nc *NetworkConfigurator, iface *v1.Interface) (InterfaceConfigurator, error) {
			assert.Equal(t, nc, nc, "NetworkConfigurator instance should be passed to NewConnectionHandler")
			switch iface.InterfaceName {
			case newIfaceOne.InterfaceName:
				return mockEthernetConfigurator, nil
			default:
				return nil, errors.New("mock error")
			}
		})

	mockEthernetConfigurator.On("Configure").Return(nil).Once()

	mockInterfaceStateManager.On("Restore").Return(nil).Once()

	// Act
	err := nc.Apply(&v1.NetworkSettings{Interfaces: newIfaces})

	// Assertions
	assert.Error(t, err, "Apply should return an error")

	mockEthernetConfigurator.AssertExpectations(t)
	mockInterfaceStateManager.AssertExpectations(t)
}

func TestNetworkConfigurator_Apply_Backup_Error(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*NetworkConfigurator](), "GetNetworkInterfaces",
		func() []*v1.Interface {
			return existingIfaces
		})

	mockInterfaceStateManager := new(networking.MockInterfaceStateManager)

	patches.ApplyFunc(NewInterfaceStateHandler,
		func(_ nm.NetworkManager, _ []*v1.Interface) InterfaceStateManager {
			return mockInterfaceStateManager
		})

	mockInterfaceStateManager.On("Backup").Return(errors.New("mock backup error")).Once()

	// Act
	err := nc.Apply(&v1.NetworkSettings{Interfaces: newIfaces})

	// Assertions
	assert.Error(t, err, "Apply should return an error")
	mockInterfaceStateManager.AssertExpectations(t)
}

func TestNetworkConfigurator_Apply_SkipsUnconfiguredInterfaces(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	configuredIface := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaces := []*v1.Interface{configuredIface}

	// Unconfigured interface - no DHCP, no static IP, should be skipped
	unconfiguredIface := &v1.Interface{
		InterfaceName: "eth2",
		MacAddress:    "00:15:5D:38:01:60",
	}

	newIfaces := []*v1.Interface{configuredIface, unconfiguredIface}

	mockEthernetConfigurator := new(networking.MockEthernetConfigurator)

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeFor[*NetworkConfigurator](), "GetNetworkInterfaces",
		func() []*v1.Interface {
			return existingIfaces
		})

	mockInterfaceStateManager := new(networking.MockInterfaceStateManager)

	patches.ApplyFunc(NewInterfaceStateHandler,
		func(_ nm.NetworkManager, _ []*v1.Interface) InterfaceStateManager {
			return mockInterfaceStateManager
		})

	mockInterfaceStateManager.On("Backup").Return(nil).Once()

	// NewInterfaceConfigurator should only be called for the configured interface.
	// If it gets called for the unconfigured one, the test would fail because
	// there is no matching mock expectation for it.
	patches.ApplyFunc(NewInterfaceConfigurator,
		func(nc *NetworkConfigurator, iface *v1.Interface) (InterfaceConfigurator, error) {
			assert.Equal(t, configuredIface.InterfaceName, iface.InterfaceName,
				"NewInterfaceConfigurator should only be called for the configured interface")
			return mockEthernetConfigurator, nil
		})

	// Configure should only be called once - for the configured interface
	mockEthernetConfigurator.On("Configure").Return(nil).Once()

	patches.ApplyPrivateMethod(reflect.TypeFor[*NetworkConfigurator](), "resetGateway",
		func(_, _ []*v1.Interface) error {
			return nil
		})

	// Act
	err := nc.Apply(&v1.NetworkSettings{Interfaces: newIfaces})

	// Assertions
	assert.NoError(t, err, "Apply should succeed even with unconfigured interfaces in payload")
	mockEthernetConfigurator.AssertExpectations(t)
	mockInterfaceStateManager.AssertExpectations(t)
}

func TestNetworkConfigurator_resetGateway_WithInterfaceNames_GatewayChange_Success(t *testing.T) {
	// Arrange
	mockGatewayManager := new(networking.MockGatewayManager)

	nc := &NetworkConfigurator{
		gnm:            new(gonetworkmanager.MockNetworkManager),
		gatewayManager: mockGatewayManager,
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceTwo := &v1.Interface{
		InterfaceName: "eth1",
		MacAddress:    "00:0A:95:9D:68:17",
		DHCP:          "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceTwo, newIfaceThree}

	mockGatewayManager.On("Reset", existingIfaceOne.InterfaceName).Return(nil).Once()

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
	mockGatewayManager.AssertExpectations(t)
}

func TestNetworkConfigurator_resetGateway_GatewayChange_Success(t *testing.T) {
	// Arrange
	mockGatewayManager := new(networking.MockGatewayManager)

	nc := &NetworkConfigurator{
		gnm:            new(gonetworkmanager.MockNetworkManager),
		gatewayManager: mockGatewayManager,
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceTwo := &v1.Interface{
		MacAddress: "00:0A:95:9D:68:17",
		DHCP:       "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceTwo, newIfaceThree}

	mockGatewayManager.On("Reset", existingIfaceOne.InterfaceName).Return(nil).Once()

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
	mockGatewayManager.AssertExpectations(t)
}

func TestNetworkConfigurator_resetGateway_GatewayChange_EthernetGateways_MacComparison_Success(t *testing.T) {
	// Arrange
	mockGatewayManager := new(networking.MockGatewayManager)

	nc := &NetworkConfigurator{
		gnm:            new(gonetworkmanager.MockNetworkManager),
		gatewayManager: mockGatewayManager,
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		MacAddress:       "00:0A:95:9D:68:19",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		MacAddress: "00:0A:95:9D:68:17",
		DHCP:       "enabled",
	}

	newIfaceTwo := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       "00:0A:95:9D:68:19",
		DHCP:             "enabled",
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceTwo}

	mockGatewayManager.On("Reset", existingIfaceOne.InterfaceName).Return(nil).Once()

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
	mockGatewayManager.AssertExpectations(t)
}

func TestNetworkConfigurator_resetGateway_GatewayChange_EthernetGateways_LabelComparison_Success(t *testing.T) {
	// Arrange
	mockGatewayManager := new(networking.MockGatewayManager)

	nc := &NetworkConfigurator{
		gnm:            new(gonetworkmanager.MockNetworkManager),
		gatewayManager: mockGatewayManager,
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:17",
		Label:            "ethOne",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		MacAddress:       "00:0A:95:9D:68:19",
		DHCP:             "enabled",
		Label:            "ethTwo",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		Label: "ethOne",
		DHCP:  "enabled",
	}

	newIfaceTwo := &v1.Interface{
		GatewayInterface: true,
		Label:            "ethTwo",
		DHCP:             "enabled",
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceTwo}

	mockGatewayManager.On("Reset", existingIfaceOne.InterfaceName).Return(nil).Once()

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
	mockGatewayManager.AssertExpectations(t)
}

func TestNetworkConfigurator_resetGateway_EthernetGatewaysNoChange_WithInterfaceName_Success(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
}

func TestNetworkConfigurator_resetGateway_GSMGatewayNoChange_WithInterfaceName_Success(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		InterfaceName: "eth0",
		MacAddress:    "00:0A:95:9D:68:16",
		DHCP:          "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName:    "gsm0",
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		InterfaceName: "eth0",
		MacAddress:    "00:0A:95:9D:68:16",
		DHCP:          "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
}

func TestNetworkConfigurator_resetGateway_EthernetGatewaysNoChange_Success(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	newIfaceThree := &v1.Interface{
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
}

func TestNetworkConfigurator_resetGateway_GSMGatewayNoChange_Success(t *testing.T) {
	// Arrange
	nc := &NetworkConfigurator{
		gnm: new(gonetworkmanager.MockNetworkManager),
	}

	existingIfaceOne := &v1.Interface{
		InterfaceName: "eth0",
		MacAddress:    "00:0A:95:9D:68:16",
		DHCP:          "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName:    "gsm0",
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceOne := &v1.Interface{
		MacAddress: "00:0A:95:9D:68:16",
		DHCP:       "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceOne, newIfaceThree}

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.NoError(t, err, "resetGateway should not return an error")
}

func TestNetworkConfigurator_resetGateway_Error(t *testing.T) {
	// Arrange
	mockGatewayManager := new(networking.MockGatewayManager)

	nc := &NetworkConfigurator{
		gnm:            new(gonetworkmanager.MockNetworkManager),
		gatewayManager: mockGatewayManager,
	}

	existingIfaceOne := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "eth0",
		MacAddress:       "00:0A:95:9D:68:16",
		DHCP:             "enabled",
	}

	existingIfaceTwo := &v1.Interface{
		InterfaceName:    "eth1",
		MacAddress:       "00:0A:95:9D:68:17",
		DHCP:             "enabled",
		GatewayInterface: false,
	}

	existingIfaceThree := &v1.Interface{
		InterfaceName: "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	existingIfaces := []*v1.Interface{existingIfaceOne, existingIfaceTwo, existingIfaceThree}

	newIfaceTwo := &v1.Interface{
		InterfaceName: "eth1",
		MacAddress:    "00:0A:95:9D:68:17",
		DHCP:          "enabled",
	}

	newIfaceThree := &v1.Interface{
		GatewayInterface: true,
		InterfaceName:    "gsm0",
		GsmConfiguration: &v1.Interface_GsmConf{
			Apn:      "internet",
			Pin:      "1234",
			Username: "test",
			Password: "password",
		},
	}

	newIfaces := []*v1.Interface{newIfaceTwo, newIfaceThree}

	mockGatewayManager.On("Reset", existingIfaceOne.InterfaceName).Return(errors.New("mock error")).Once()

	// Act
	err := nc.resetGateway(existingIfaces, newIfaces)

	// Assert
	assert.Error(t, err, "resetGateway should return an error")
	mockGatewayManager.AssertExpectations(t)
}

func Test_GetAllEthernetDevices_SkipsNilDevices(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockDeviceWired := &mockgnm.MockDeviceWired{}

	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list with nil devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{nil, mockDeviceWired, nil}, nil
	})

	// Mock GetPropertyDeviceType to return Ethernet
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWired), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})

	// Mock GetPath
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWired), "GetPath", func(_ nm.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/1"
	})

	// Mock NewDeviceWired
	patches.ApplyFunc(nm.NewDeviceWired, func(path dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockDeviceWired, nil
	})

	// Call the function
	result := nc.getAllEthernetDevices()

	// Assertions - should only return 1 device, skipping the nil entries
	assert.Equal(t, 1, len(result), "Should skip nil devices and return only valid devices")
	assert.Equal(t, mockDeviceWired, result[0], "Should return the valid mock device")
}

func Test_GetAllEthernetDevices_SkipsDevicesWithTypeError(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockDeviceWithError := &mockgnm.MockDeviceWired{}
	mockValidDevice := &mockgnm.MockDeviceWired{}

	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list with devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockDeviceWithError, mockValidDevice}, nil
	})

	// Mock GetPropertyDeviceType - use call counter to differentiate between devices
	callCount := 0
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWithError), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		callCount++
		if callCount == 1 {
			return nm.NmDeviceTypeUnknown, errors.New("failed to get device type")
		}
		return nm.NmDeviceTypeEthernet, nil
	})

	// Mock GetPath for valid device
	patches.ApplyMethod(reflect.TypeOf(mockValidDevice), "GetPath", func(_ nm.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/2"
	})

	// Mock NewDeviceWired
	patches.ApplyFunc(nm.NewDeviceWired, func(path dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockValidDevice, nil
	})

	// Call the function
	result := nc.getAllEthernetDevices()

	// Assertions - should only return 1 device, skipping the one with error
	assert.Equal(t, 1, len(result), "Should skip devices with GetPropertyDeviceType errors")
	assert.Equal(t, mockValidDevice, result[0], "Should return only the valid device")
}

func Test_GetAllEthernetDevices_SkipsNonEthernetDeviceTypes(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockEthernetDevice := &mockgnm.MockDeviceWired{}
	mockWifiDevice := &mockgnm.MockDeviceWired{}
	mockBridgeDevice := &mockgnm.MockDeviceWired{}

	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list with mixed device types
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockEthernetDevice, mockWifiDevice, mockBridgeDevice}, nil
	})

	// Mock GetPropertyDeviceType to return different types
	callCount := 0
	patches.ApplyMethod(reflect.TypeOf(mockEthernetDevice), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		callCount++
		switch callCount {
		case 1:
			return nm.NmDeviceTypeEthernet, nil
		case 2:
			return nm.NmDeviceTypeWifi, nil
		case 3:
			return nm.NmDeviceTypeBridge, nil
		default:
			return nm.NmDeviceTypeUnknown, nil
		}
	})

	// Mock GetPath for Ethernet device
	patches.ApplyMethod(reflect.TypeOf(mockEthernetDevice), "GetPath", func(_ nm.Device) dbus.ObjectPath {
		return "/org/freedesktop/NetworkManager/Devices/1"
	})

	// Mock NewDeviceWired
	patches.ApplyFunc(nm.NewDeviceWired, func(path dbus.ObjectPath) (nm.DeviceWired, error) {
		return mockEthernetDevice, nil
	})

	// Call the function
	result := nc.getAllEthernetDevices()

	// Assertions - should only return 1 Ethernet device, skipping WiFi and Bridge
	assert.Equal(t, 1, len(result), "Should return only Ethernet devices")
	assert.Equal(t, mockEthernetDevice, result[0], "Should return the Ethernet device")
}

func Test_GetAllNetworkDevices_SkipsNilDevices(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockValidDevice := &mockgnm.MockDevice{}
	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list with a nil device
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{nil, mockValidDevice}, nil
	})

	// Mock GetPropertyDeviceType for valid device
	patches.ApplyMethod(reflect.TypeOf(mockValidDevice), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		return nm.NmDeviceTypeEthernet, nil
	})

	// Call the function
	result := nc.getAllNetworkDevices()

	// Assertions - should only return 1 device, skipping the nil device
	assert.Equal(t, 1, len(result), "Should skip nil devices")
	assert.Equal(t, mockValidDevice, result[0], "Should return only the valid device")
}

func Test_GetAllNetworkDevices_SkipsDevicesWithGetPropertyDeviceTypeError(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockDeviceWithError := &mockgnm.MockDevice{}
	mockValidDevice := &mockgnm.MockDevice{}
	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return a list with devices
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockDeviceWithError, mockValidDevice}, nil
	})

	// Mock GetPropertyDeviceType to return error for first device, success for second
	callCount := 0
	patches.ApplyMethod(reflect.TypeOf(mockDeviceWithError), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		callCount++
		if callCount == 1 {
			return nm.NmDeviceTypeUnknown, errors.New("device type error")
		}
		return nm.NmDeviceTypeModem, nil
	})

	// Call the function
	result := nc.getAllNetworkDevices()

	// Assertions - should only return 1 device, skipping the one with error
	assert.Equal(t, 1, len(result), "Should skip devices with GetPropertyDeviceType errors")
	assert.Equal(t, mockValidDevice, result[0], "Should return only the valid device")
}

func Test_GetAllNetworkDevices_SkipsUnsupportedDeviceTypes(t *testing.T) {
	nc := &NetworkConfigurator{}
	mockNetworkManager := &mockgnm.MockNetworkManager{}
	mockWifiDevice := &mockgnm.MockDevice{}
	mockBridgeDevice := &mockgnm.MockDevice{}
	mockEthernetDevice := &mockgnm.MockDevice{}
	nc.gnm = mockNetworkManager

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock GetDevices to return mixed device types
	patches.ApplyMethod(reflect.TypeOf(mockNetworkManager), "GetDevices", func(_ nm.NetworkManager) ([]nm.Device, error) {
		return []nm.Device{mockWifiDevice, mockBridgeDevice, mockEthernetDevice}, nil
	})

	// Mock GetPropertyDeviceType to return different types
	callCount := 0
	patches.ApplyMethod(reflect.TypeOf(mockWifiDevice), "GetPropertyDeviceType", func(_ nm.Device) (nm.NmDeviceType, error) {
		callCount++
		switch callCount {
		case 1:
			return nm.NmDeviceTypeWifi, nil
		case 2:
			return nm.NmDeviceTypeBridge, nil
		case 3:
			return nm.NmDeviceTypeEthernet, nil
		default:
			return nm.NmDeviceTypeUnknown, nil
		}
	})

	// Call the function
	result := nc.getAllNetworkDevices()

	// Assertions - should only return 1 Ethernet device, skipping WiFi and Bridge
	assert.Equal(t, 1, len(result), "Should return only supported device types")
	assert.Equal(t, mockEthernetDevice, result[0], "Should return only Ethernet device")
}
