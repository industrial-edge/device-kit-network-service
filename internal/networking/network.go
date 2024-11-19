/*
 * Copyright (c) 2021 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	nm "github.com/Wifx/gonetworkmanager/v2"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"strings"
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
	gnm nm.NetworkManager
}

// NewNetworkConfiguratorWithNM creates new NetworkConfigurator instance
func NewNetworkConfiguratorWithNM(wifxNetworkManager nm.NetworkManager) *NetworkConfigurator {
	return &NetworkConfigurator{gnm: wifxNetworkManager}
}

// NewNetworkConfigurator creates new NetworkConfigurator instance
func NewNetworkConfigurator() *NetworkConfigurator {
	val, _ := nm.NewNetworkManager()
	return &NetworkConfigurator{gnm: val}
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

// GetEthernetInterfaces Returns All Ethernet typed interfaces on a device
func (nc *NetworkConfigurator) GetEthernetInterfaces() []*v1.Interface {

	devices := nc.getAllEthernetDevices()
	var interfaces []*v1.Interface

	for _, device := range devices {
		name, _ := device.GetPropertyInterface()
		mac, _ := device.GetPropertyHwAddress()
		log.Printf("interface found name: %v MAC: %v", name, mac)
		proto := DBusToProto(device)
		interfaces = append(interfaces, proto)
	}
	return interfaces
}

// ArePreconditionsOk Checks all preconditions before applying any settings.
func (nc *NetworkConfigurator) ArePreconditionsOk(newSettings *v1.NetworkSettings) (bool, error) {
	return verify(newSettings, nc)
}

// Apply Applies given settings, if any error occures all Interfaces in system will be restored to original states.
func (nc *NetworkConfigurator) Apply(newSettings *v1.NetworkSettings) error {
	log.Println("new settings request -- ", newSettings)
	var backups []nm.ConnectionSettings

	//iterate through all interfaces in given new Settings
	for _, element := range newSettings.Interfaces {

		//try APPLY new settings to each network interface
		backup, err := nc.applyAndBackupSettings(element)

		//add backup if any active connections exists before
		if backup != nil {
			backups = append(backups, backup)
			log.Println("backup  : > ", backup)
		}
		//if any error occurs, all interfaces will be RESTOREd to original
		if err != nil {
			log.Println("applying new settings failed for:", err)
			log.Println("Restoring all settings:")
			for _, data := range backups {
				if data != nil {
					_ = nc.restoreConnection(data)

				}
			}
			//return error to caller since new settings could not apply,but restored.
			return err
		}
	}
	log.Println("all interface(s) configured successfully")
	return nil
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
		deviceType, _ := device.GetPropertyDeviceType()
		if deviceType == nm.NmDeviceTypeEthernet {
			wired, _ := nm.NewDeviceWired(device.GetPath())
			foundEthernetDevices = append(foundEthernetDevices, wired)
		}
	}
	return foundEthernetDevices
}

// applyAndBackupSettings applies the provided network settings to the device
// and creates a backup of the existing settings before applying the new ones.
func (nc *NetworkConfigurator) applyAndBackupSettings(protoData *v1.Interface) (nm.ConnectionSettings, error) {
	device, err := nc.getDeviceBy(protoData)
	if err != nil {
		return nil, err
	}

	backup := nc.createBackupFromExisting(device)
	settings, err := nc.prepareSettings(protoData, device)
	if err != nil {
		return backup, err
	}

	if err := nc.updateConnections(device, settings); err != nil {
		return backup, err
	}

	if err := ConfigureExistingGatewayInterfacesExceptProtoData(protoData, *nc); err != nil {
		return backup, err
	}

	return backup, nil
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
		return nil, err
	}
	return newSettingsFromProto(protoData, deviceName), nil
}

// updateConnections updates the connections for the given Ethernet device
// by deleting old connections and adding the new settings.
func (nc *NetworkConfigurator) updateConnections(device nm.DeviceWired, settings nm.ConnectionSettings) error {
	connections := listConnections(device)
	if err := nc.deleteOldConnections(connections); err != nil {
		log.Println("could not delete connection: ", err)
		return err
	}

	mac, err := device.GetPropertyHwAddress()
	if err != nil {
		return err
	}

	return nc.addConnection(strings.ToUpper(mac), settings)
}

// setMACAddressInBackup ensures that the MAC address is set in the backup.
// If the MAC address is not already present in the backup, it retrieves the permanent hardware address
// from the wired device, parses it into a MAC address, and sets it in the backup.
// This is necessary to correctly restore the connection settings if needed.
func setMACAddressInBackup(backup nm.ConnectionSettings, wired nm.DeviceWired) error {
	if backup[EthernetType][MACAddressKey] == nil {
		retValue, _ := wired.GetPropertyPermHwAddress()
		macAddr, err := net.ParseMAC(retValue)
		if err == nil {
			backup[EthernetType][MACAddressKey] = []uint8(macAddr)
			return nil
		} else {
			return err
		}
	}
	return nil
}

// createBackupFromExisting creates a backup of the existing connection settings
// for the given Ethernet device.
func (nc *NetworkConfigurator) createBackupFromExisting(wired nm.DeviceWired) nm.ConnectionSettings {

	list := listConnections(wired)

	if list == nil || len(list) < 1 {
		log.Printf("there is not any connection found on device to create backup ")
		return nil
	} else {
		backup, _ := list[0].GetSettings()
		err := setMACAddressInBackup(backup, wired)
		if err != nil {
			log.Println("Cannot create backup, parse mac address error: ", err)
		}
		log.Println("created backup for existing connection")
		// new settings instance needed to be ready for applying backup
		log.Println(backup)
		return retrieveSettingsFromBackup(backup)
	}
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
	}
	return err
}

// restoreConnection restores the connection settings from the provided backup.
func (nc *NetworkConfigurator) restoreConnection(backup nm.ConnectionSettings) error {
	var mac net.HardwareAddr
	mac = backup[EthernetType][MACAddressKey].([]byte)

	err := nc.addConnection(mac.String(), backup)
	if err != nil {
		log.Printf("rostoreConnection failed for mac: %v", mac)
		return err
	} else {
		log.Printf("rostoreConnection success for mac: %v", mac)
		return nil
	}
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
