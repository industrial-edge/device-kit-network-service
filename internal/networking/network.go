/*
 * Copyright (c) 2021 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	nm "github.com/Wifx/gonetworkmanager"
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

	getDeviceWithMac(mac string) nm.DeviceWired
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

//GetInterfaceWithMac Returns All Ethernet typed interfaces on a device
func (nc *NetworkConfigurator) GetInterfaceWithMac(mac string) *v1.Interface {
	device := nc.getDeviceWithMac(mac)
	if device == nil {
		log.Println("Device is not found: ", mac)
	}
	return DBusToProto(device)
}

//GetEthernetInterfaces Returns All Ethernet typed interfaces on a device
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

// Checks all preconditions before applying any settings.
func (nc *NetworkConfigurator) ArePreconditionsOk(newSettings *v1.NetworkSettings) (bool, error) {
	return verify(newSettings, nc)
}

// Applies given settings, if any error occures all Interfaces in system will be restored to original states.
func (nc *NetworkConfigurator) Apply(newSettings *v1.NetworkSettings) error {
	log.Println("new settings request -- ", newSettings)
	var backups []nm.ConnectionSettings

	//iterate through all interfaces in given new Settings
	for _, element := range newSettings.Interfaces {

		//try APPLY new settings to each network interface
		backup, err := nc.tryApply(element)

		//add backup if any active connections exists before
		if backup != nil {
			backups = append(backups, backup)
		}
		//if any error occurs, all interfaces will be RESTOREd to original
		if err != nil {
			log.Println("applying new settings failed for:", err)
			log.Println("Restoring all settings:")
			for _, data := range backups {
				if data != nil {
					nc.restoreConnection(data)
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
		log.Println("Device does not exist: ", mac)
	}
	return retVal
}

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

func (nc *NetworkConfigurator) tryApply(protoData *v1.Interface) (nm.ConnectionSettings, error) {
	mac := strings.ToUpper(protoData.MacAddress)
	device := nc.getDeviceWithMac(mac)
	backup := nc.createBackupFromExisting(device)

	deviceName, _ := device.GetPropertyInterface()
	settings := newSettingsFromProto(protoData, deviceName)
	//delete all existing connection profiles for this device
	nc.deleteOldConnections(device)
	//add this
	err := nc.addConnection(mac, settings)

	return backup, err
}

func (nc *NetworkConfigurator) createBackupFromExisting(wired nm.DeviceWired) nm.ConnectionSettings {

	list := listConnections(wired)

	if list == nil || len(list) < 1 {
		log.Printf("there is not any connection found on device to create backup ")
		return nil
	} else {
		backup, _ := list[0].GetSettings()
		log.Println("created backup for existing connection")
		// new settings instance needed to be ready for applying backup
		return retrieveSettingsFromBackup(backup)
	}
}
func (nc *NetworkConfigurator) addConnection(mac string, settings nm.ConnectionSettings) error {
	device := nc.getDeviceWithMac(mac)
	settingsM, _ := nm.NewSettings()
	conn, err := settingsM.AddConnection(settings)
	if err == nil {
		log.Printf("settings applied for device %v successfully", mac)
		_, aErr := nc.gnm.ActivateConnection(conn, device)
		if aErr != nil {
			log.Println("configuration applied,but could not activated since: ", aErr)
		}
	}
	return err
}
func (nc *NetworkConfigurator) restoreConnection(backup nm.ConnectionSettings) {
	var mac net.HardwareAddr
	mac = backup[EthernetType][MACAddressKey].([]byte)

	err := nc.addConnection(mac.String(), backup)
	if err != nil {
		log.Println("Rare: Restoration Failed!")
	} else {
		log.Println("Restoration completed for ", mac.String())
	}

}

func (nc *NetworkConfigurator) deleteOldConnections(device nm.DeviceWired) {
	list := listConnections(device)
	for _, connection := range list {
		_ = connection.Delete()
	}
}
