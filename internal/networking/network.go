/*
 * Copyright © Siemens 2020 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	"fmt"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/factory"
	"reflect"
	"strings"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

// Network interface that can perform
type Network interface {
	GetEthernetInterfaces() []*v1.Interface
	ArePreconditionsOk(newSettings *v1.NetworkSettings) (bool, error)
	Apply(newSettings []*v1.Interface) error
	GetInterfaceWithMac(mac string) *v1.Interface
	GetInterfaceWithLabel(Label string) *v1.Interface

	getDeviceWithMac(mac string) nm.DeviceWired
	getDeviceWithLabel(label string) nm.DeviceWired
	getMacWithInterfaceName(InterfaceName string) string
}

// NetworkConfigurator implements Network Interface.
type NetworkConfigurator struct {
	gnm            nm.NetworkManager
	gatewayManager GatewayHelper
}

// NewNetworkConfiguratorWithNM creates new NetworkConfigurator instance
func NewNetworkConfiguratorWithNM(networkManager nm.NetworkManager) *NetworkConfigurator {
	return &NetworkConfigurator{
		gnm:            networkManager,
		gatewayManager: NewGatewayManager(networkManager),
	}
}

//### PUBLIC FUNCTIONS

// GetInterfaceWithMac Returns All Ethernet typed interfaces on a device
func (nc *NetworkConfigurator) GetInterfaceWithMac(mac string) *v1.Interface {
	device := nc.getDeviceWithMac(mac)
	if device == nil {
		log.Println("Device is not found: ", mac)
	}
	return DBusToProto(device)
}

// GetInterfaceWithLabel Returns All Ethernet typed interfaces on a device
func (nc *NetworkConfigurator) GetInterfaceWithLabel(Label string) *v1.Interface {
	device := nc.getDeviceWithLabel(Label)
	if device == nil {
		log.Println("Device is not found: ", Label)
	}
	return DBusToProto(device)
}

// IsGatewayInterface checks if the interface with the given MAC address is the gateway interface.
func (nc *NetworkConfigurator) IsGatewayInterface(mac string) bool {
	devices := nc.getAllEthernetDevices()

	gatewayMAC := nc.findGatewayMAC(devices)
	log.Printf("Identified gateway MAC: %v\n", gatewayMAC)

	// Compare the gateway MAC with the input MAC.
	isGateway := strings.EqualFold(gatewayMAC, mac)
	log.Printf("Device with MAC %s has gateway interface value: %t\n", mac, isGateway)

	return isGateway
}

// findGatewayMAC identifies the MAC address of the gateway interface with the lowest metric.
func (nc *NetworkConfigurator) findGatewayMAC(devices []nm.DeviceWired) string {
	log.Println("Starting findGatewayMAC: Identifying the gateway MAC with the lowest metric.")

	var lowestMetric uint8 = common.MaxMetricValue // Initialize with the highest possible metric value
	var gatewayMAC string

	for _, device := range devices {
		hwAddr, _ := device.GetPropertyHwAddress()
		log.Printf("Processing device: MAC=%s\n", hwAddr)

		mac, metric, err := nc.getDeviceGatewayIDAndMetric(device)
		if err != nil {
			log.Printf("Error fetching gateway MAC and metric for device: MAC=%s: %v\n", hwAddr, err)
			continue
		}
		log.Printf("Device with MAC %s has metric value: %d\n", mac, metric)
		if metric < lowestMetric {
			log.Printf("New lowest metric found: %d (previous: %d). Updating gateway MAC to %v.\n", metric, lowestMetric, mac)
			lowestMetric = metric
			gatewayMAC = mac
		}
	}

	log.Printf("Identified gateway MAC: %v\n", gatewayMAC)
	return gatewayMAC
}

func (nc *NetworkConfigurator) getDeviceGatewayIDAndMetric(device nm.Device) (string, uint8, error) {
	if device == nil {
		log.Printf("getDeviceGatewayIDAndMetric: device is nil")
		return "", 0, fmt.Errorf("device is nil")
	}

	log.Printf("getDeviceGatewayIDAndMetric: fetching active connection")
	conn, err := device.GetPropertyActiveConnection()
	if err != nil || conn == nil {
		log.Printf("getDeviceGatewayIDAndMetric: no active connection err=%v", err)
		return "", 0, fmt.Errorf("no active connection")
	}

	ipv4, err := conn.GetPropertyIP4Config()
	if err != nil || ipv4 == nil || reflect.ValueOf(ipv4).IsNil() {
		log.Printf("getDeviceGatewayIDAndMetric: failed to get IPv4 config err=%v", err)
		return "", 0, fmt.Errorf("failed to get IPv4 configuration")
	}

	routeData, err := ipv4.GetPropertyRouteData()
	if err != nil || routeData == nil {
		log.Printf("getDeviceGatewayIDAndMetric: failed to get route data err=%v", err)
		return "", 0, fmt.Errorf("failed to get route data")
	}
	return nc.inspectRoutesForGateway(device, routeData)
}

func (nc *NetworkConfigurator) inspectRoutesForGateway(device nm.Device, routeData []nm.IP4RouteData) (string, uint8, error) {
	for _, route := range routeData {
		log.Printf("Inspecting route destination=%v prefix=%v metric=%v", route.Destination, route.Prefix, route.Metric)
		if route.Destination == common.OutgoingRouteDestination && route.Prefix == common.OutgoingRoutePrefix {
			deviceType, _ := device.GetPropertyDeviceType()
			if deviceType == nm.NmDeviceTypeEthernet {
				wired, err := nm.NewDeviceWired(device.GetPath())
				if err != nil {
					return "", 0, err
				}
				mac, err := wired.GetPropertyHwAddress()
				if err != nil {
					return "", 0, err
				}
				return mac, route.Metric, nil
			}
			log.Printf("Gateway identified via non-Ethernet device, no MAC available, metric=%d", route.Metric)
			return "", route.Metric, nil
		}
	}
	log.Printf("getDeviceGatewayIDAndMetric: no gateway route found")
	return "", 0, fmt.Errorf("no matching gateway route found")
}

// GetNetworkInterfaces returns all network interfaces (Ethernet and Modem) on the device.
func (nc *NetworkConfigurator) GetNetworkInterfaces() []*v1.Interface {

	log.Println("GetNetworkInterfaces: collecting all network interfaces.")
	devices := nc.getAllNetworkDevices()
	log.Printf("Fetched %d devices: %v\n", len(devices), devices)

	// Collect all interfaces into a slice.
	var interfaces []*v1.Interface
	for _, device := range devices {
		deviceType, _ := device.GetPropertyDeviceType()
		log.Printf("Processing device with type: %v, path: %v", deviceType, device.GetPath())
		proto := DBusToProto(device)
		interfaces = append(interfaces, proto)
	}
	// Identify the gateway interface.
	gatewayInterface := nc.findGatewayInterface(devices, interfaces)
	if gatewayInterface != nil {
		log.Printf("Gateway interface identified: %v\n", gatewayInterface)
		gatewayInterface.GatewayInterface = true
	}
	log.Printf("Returning %d interfaces: %v\n", len(interfaces), interfaces)
	return interfaces
}

func (nc *NetworkConfigurator) findGatewayInterface(devices []nm.Device, interfaces []*v1.Interface) *v1.Interface {
	log.Println("Starting findGatewayInterface: Identifying gateway interface.")

	var lowestMetric uint8 = common.MaxMetricValue // Initialize with the highest possible metric value
	var gatewayInterface *v1.Interface

	for i, device := range devices {
		deviceInterface, _ := device.GetPropertyInterface()
		log.Printf("Processing device at index %d: device's interface name=%s\n", i, deviceInterface)

		_, metric, err := nc.getDeviceGatewayIDAndMetric(device)
		if err != nil {
			log.Printf("Error fetching gateway MAC and metric for device with device's interface name=%s: %v\n", deviceInterface, err)
			continue
		}

		log.Printf("Device with Interface %s has metric %d and Interface %v\n", deviceInterface, metric, deviceInterface)

		if metric < lowestMetric {
			log.Printf("New lowest metric found: %d (previous: %d). Updating gateway interface.\n", metric, lowestMetric)
			lowestMetric = metric
			gatewayInterface = interfaces[i]
		}
	}

	log.Printf("Identified Gateway interface: %v\n", gatewayInterface)

	return gatewayInterface
}

// ArePreconditionsOk Checks all preconditions before applying any settings.
func (nc *NetworkConfigurator) ArePreconditionsOk(newSettings *v1.NetworkSettings) (bool, error) {
	return verify(newSettings, nc)
}

// Apply Applies given settings, if any error occures all Interfaces in system will be restored to original states.
func (nc *NetworkConfigurator) Apply(newSettings *v1.NetworkSettings) error {
	log.Println("Request to apply new network settings...")
	log.Printf("Received network settings payload: %v", newSettings)

	var hasError bool

	existingIfaces := nc.GetNetworkInterfaces()

	// Create a new interface state handler for this Apply call.
	// A fresh handler is instantiated on each invocation to ensure
	// backup and restore operations are isolated per Apply execution.
	ifaceHandler := NewInterfaceStateHandler(nc.gnm, existingIfaces)

	if err := ifaceHandler.Backup(); err != nil {
		log.Println("Failed to backup existing interface states: ", err)
		return err
	}

	defer func() {
		if hasError {
			ifaceHandler.Restore()
		}
	}()

	newIfaces := newSettings.GetInterfaces()
	for _, iface := range newIfaces {

		// Skip unconfigured interfaces (e.g., disconnected Ethernet with no DHCP and no static IP).
		if !isInterfaceConfigured(iface) {
			log.Printf("Skipping unconfigured interface: %s (no DHCP and no static IP)", iface.GetInterfaceName())
			continue
		}

		// For each interface supplied to this Apply call, create a new configurator instance
		// based on its type (Ethernet or GSM). Each configurator is created fresh for this
		// Apply invocation and is responsible for configuring its respective interface independently.
		configurator, err := NewInterfaceConfigurator(nc, iface)
		if err != nil {
			log.Println("Failed to init configurator: ", err)
			hasError = true
			return err
		}

		if err := configurator.Configure(); err != nil {
			log.Println("Failed to apply new settings:", err)
			hasError = true
			return err
		}
	}

	if err := nc.resetGateway(existingIfaces, newIfaces); err != nil {
		log.Println("Failed to reset gateway settings:", err)
		hasError = true
		return err
	}

	log.Println("New network settings applied successfully.")
	return nil
}

func (nc *NetworkConfigurator) resetGateway(existingIfaces, newIfaces []*v1.Interface) error {
	deviceName := nc.findGatewayToReset(existingIfaces, newIfaces)
	if len(deviceName) == 0 {
		return nil
	}
	return nc.gatewayManager.Reset(deviceName)
}

func (nc *NetworkConfigurator) findGatewayToReset(existingIfaces, incomingIfaces []*v1.Interface) string {
	incomingGateway := nc.findGateway(incomingIfaces)
	if incomingGateway == nil {
		log.Println("No new gateway interface specified. No need to reset gateway settings.")
		return ""
	}

	existingGateway := nc.findGateway(existingIfaces)
	if existingGateway == nil {
		log.Println("No previous gateway interface found. No need to reset gateway settings.")
		return ""
	}

	if nc.isGatewayChanged(existingGateway, incomingGateway) {
		log.Println("Gateway interface changed. Need to reset gateway settings.")
		return existingGateway.GetInterfaceName()
	}

	log.Println("No changes to gateway interface detected. No need to reset gateway settings.")
	return ""
}

// For existing interfaces, assuming that there is maximum one gateway interface.
// Although unlikely, if multiple gateway interfaces do exist, only the first one found will be returned.
// For incoming interfaces, if multiple interfaces are marked as gateway, the validation should catch it before.
func (nc *NetworkConfigurator) findGateway(ifaces []*v1.Interface) *v1.Interface {
	for _, iface := range ifaces {
		if iface.GetGatewayInterface() {
			return iface
		}
	}
	return nil
}

func (nc *NetworkConfigurator) isGatewayChanged(existingGateway, incomingGateway *v1.Interface) bool {
	existingGatewayIsGSM := isGSMInterface(existingGateway)
	incomingGatewayIsGSM := isGSMInterface(incomingGateway)

	if existingGatewayIsGSM && incomingGatewayIsGSM {
		log.Println("Both existing and incoming gateway interfaces are GSM. Comparing GSM settings to determine if gateway has changed...")
		return !gsmInterfaceMatches(existingGateway, incomingGateway)
	} else if !existingGatewayIsGSM && !incomingGatewayIsGSM {
		log.Println("Both existing and incoming gateway interfaces are Ethernet. Comparing Ethernet settings to determine if gateway has changed...")
		return !ethernetInterfaceMatches(existingGateway, incomingGateway)
	} else {
		log.Println("Existing and incoming gateway interfaces are of different types. Gateway has changed.")
		return true
	}
}

func (nc *NetworkConfigurator) restoreEthernetConnection(backup nm.ConnectionSettings) error {
	var mac net.HardwareAddr
	mac = backup[common.EthernetType][common.MACAddressKey].([]byte)

	err := nc.addConnection(mac.String(), backup)
	if err != nil {
		log.Printf("restoreEthernetConnection failed for mac: %v", mac)
		return err
	} else {
		log.Printf("restoreEthernetConnection success for mac: %v", mac)
		return nil
	}
}

func (nc *NetworkConfigurator) CreateConnection(newSettings *v1.ConnectionSettings) error {
	connType := newSettings.ConnectionType
	conn := factory.ConnectionFactory(connType, nc.gnm)
	if conn == nil {
		return fmt.Errorf("unsupported connection type: %v", connType)
	}
	return conn.CreateConnection(newSettings)
}

func (nc *NetworkConfigurator) RemoveConnection(newSettings *v1.ConnectionSettings) error {
	connType := newSettings.ConnectionType
	conn := factory.ConnectionFactory(connType, nc.gnm)
	if conn == nil {
		return fmt.Errorf("unsupported connection type: %v", connType)
	}
	return conn.RemoveConnection(newSettings)
}

//### PRIVATE functions
//#####################

func (nc *NetworkConfigurator) getDeviceWithMac(mac string) nm.DeviceWired {
	var retVal nm.DeviceWired
	for _, device := range nc.getAllEthernetDevices() {
		hw, _ := device.GetPropertyHwAddress()
		if strings.ToUpper(hw) == strings.ToUpper(mac) {
			retVal = device

		}
	}
	if retVal == nil {
		log.Println("getDeviceWithMac Device does not exist: ", mac)
	}
	return retVal
}

// getDeviceWithLabel returns a device which has a label match with input parameter
func (nc *NetworkConfigurator) getDeviceWithLabel(label string) nm.DeviceWired {
	expectedInterface := getInterfaceForLabel(label)

	for _, device := range nc.getAllEthernetDevices() {
		interfaceName, _ := device.GetPropertyInterface()

		if strings.ToUpper(expectedInterface) == strings.ToUpper(interfaceName) {
			log.Println("getDeviceWithLabel Device Found for the label: ", label)
			return device
		}
	}

	log.Println("getDeviceWithLabel Device does not exist for: ", label)
	return nil
}

// getAllEthernetDevices retrieves all Ethernet devices available on the system
// by querying the NetworkManager and filtering the devices by type.
func (nc *NetworkConfigurator) getAllEthernetDevices() []nm.DeviceWired {

	var foundEthernetDevices []nm.DeviceWired

	list, _ := nc.gnm.GetDevices()

	for _, device := range list {
		if device == nil {
			log.Println("Skipping nil device from NetworkManager")
			continue
		}
		deviceType, err := device.GetPropertyDeviceType()
		if err != nil {
			log.Printf("Error getting device type: %v\n", err)
			continue
		}
		if deviceType == nm.NmDeviceTypeEthernet {
			wired, err := nm.NewDeviceWired(device.GetPath())
			if err != nil {
				log.Printf("Error creating DeviceWired: %v\n", err)
				continue
			}
			foundEthernetDevices = append(foundEthernetDevices, wired)
		}
	}
	return foundEthernetDevices
}

// getAllNetworkDevices retrieves all supported network devices from NetworkManager.
// It returns Ethernet devices as wired devices and Modem (GSM/LTE) devices as generic devices
func (nc *NetworkConfigurator) getAllNetworkDevices() []nm.Device {
	var foundNetworkDevices []nm.Device
	list, err := nc.gnm.GetDevices()
	if err != nil {
		log.Println("Failed to get devices: ", err)
		return nil
	}
	for _, device := range list {
		if device == nil {
			log.Println("Skipping nil device from NetworkManager")
			continue
		}
		deviceType, err := device.GetPropertyDeviceType()
		if err != nil {
			log.Printf("Error getting device type: %v\n", err)
			continue
		}
		if deviceType == nm.NmDeviceTypeEthernet || deviceType == nm.NmDeviceTypeModem {
			log.Printf("Found supported device type: %v", deviceType)
			foundNetworkDevices = append(foundNetworkDevices, device)
		}
	}
	return foundNetworkDevices
}

// getDeviceBy retrieves the Ethernet device based on the provided protoData,
// either by MAC address or by label.
func (nc *NetworkConfigurator) getDeviceBy(protoData *v1.Interface) (nm.DeviceWired, error) {
	if protoData.MacAddress != "" {
		return nc.getDeviceWithMac(strings.ToUpper(protoData.MacAddress)), nil
	} else if protoData.Label != "" {
		return nc.getDeviceWithLabel(protoData.Label), nil
	}
	return nil, errors.New("error, Mac address or Interface name should be entered")
}

// prepareSettings prepares the connection settings based on the provided protoData
// and the Ethernet device.
func (nc *NetworkConfigurator) prepareSettings(protoData *v1.Interface, device nm.DeviceWired) (nm.ConnectionSettings, error) {
	deviceName, err := device.GetPropertyInterface()
	if err != nil {
		log.Println("Failed to get device interface name: ", err)
		return nil, err
	}
	return newSettingsFromProto(protoData, deviceName), nil
}

// getAllConnections retrieves all saved NetworkManager connections.
func (nc *NetworkConfigurator) getAllConnections() ([]nm.Connection, error) {
	settings, err := nm.NewSettings()
	if err != nil {
		log.Println("Failed to create settings manager: ", err)
		return nil, err
	}

	// List all saved NM connections
	connections, err := settings.ListConnections()
	if err != nil {
		log.Println("Failed to list connections: ", err)
		return nil, err
	}

	return connections, err
}

// updateConnections updates the connections for the given Ethernet device
// by deleting old connections and adding the new settings.
func (nc *NetworkConfigurator) updateConnections(device nm.DeviceWired, settings nm.ConnectionSettings) error {
	connections := listConnections(device)

	if err := nc.deleteOldConnections(connections); err != nil {
		return err
	}

	mac, err := device.GetPropertyHwAddress()
	if err != nil {
		log.Println("Failed to get MAC address of device: ", err)
		return err
	}

	return nc.addConnection(strings.ToUpper(mac), settings)
}

// setMACAddressInBackup ensures that the MAC address is set in the backup.
// If the MAC address is not already present in the backup, it retrieves the permanent hardware address
// from the wired device, parses it into a MAC address, and sets it in the backup.
// This is necessary to correctly restore the connection settings if needed.
func setMACAddressInBackup(backup nm.ConnectionSettings, wired nm.DeviceWired) error {
	if backup[common.EthernetType][common.MACAddressKey] == nil {
		retValue, _ := wired.GetPropertyPermHwAddress()
		macAddr, err := net.ParseMAC(retValue)
		if err == nil {
			backup[common.EthernetType][common.MACAddressKey] = []uint8(macAddr)
			return nil
		} else {
			return err
		}
	}
	return nil
}

// addConnection adds a new connection with the provided settings to the Ethernet device
// identified by the given MAC address.
func (nc *NetworkConfigurator) addConnection(mac string, settings nm.ConnectionSettings) error {
	device := nc.getDeviceWithMac(mac)
	settingsM, _ := nm.NewSettings()
	conn, err := settingsM.AddConnection(settings)
	if err == nil {
		log.Printf("settings applied for device %v successfully", mac)
		log.Printf("connection with Path: %v has been successfully added", conn.GetPath())
		_, aErr := nc.gnm.ActivateConnection(conn, device, nil)
		if aErr != nil {
			log.Println("configuration applied,but could not activated since: ", aErr)
		}
	} else {
		log.Printf("Failed to add connection for device %v: %v", mac, err)
	}
	return err
}

// deleteOldConnections deletes all old connections from the provided list of connections.
func (nc *NetworkConfigurator) deleteOldConnections(connections []nm.Connection) error {
	for _, connection := range connections {
		if err := connection.Delete(); err != nil {
			log.Printf("Failed to delete connection with Path: %v, error: %v", connection.GetPath(), err)
			return err
		}
		log.Printf("Connection with Path: %v has been successfully deleted", connection.GetPath())
	}
	return nil
}
