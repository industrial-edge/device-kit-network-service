/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package devicemanager

import (
	"errors"
	"log"

	"networkservice/internal/networking/interfaces"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

// Only Ethernet and GSM (Modem) device types are supported
var supportedDeviceTypes = map[nm.NmDeviceType]bool{
	nm.NmDeviceTypeEthernet: true,
	nm.NmDeviceTypeModem:    true,
}

type DeviceManager struct {
	nm nm.NetworkManager
}

func NewDeviceManager(nm nm.NetworkManager) interfaces.DeviceHelper {
	return &DeviceManager{
		nm: nm,
	}
}

func (d *DeviceManager) GetAvailableDevicesMap() (map[string]nm.Device, error) {
	log.Println("Retrieving available devices from NetworkManager...")

	devicesMap := make(map[string]nm.Device)

	devices, err := d.nm.GetDevices()
	if err != nil {
		log.Println("Failed to get devices from NetworkManager: ", err)
		return nil, err
	}

	for _, device := range devices {
		ifaceName, err := device.GetPropertyInterface()
		if err != nil {
			log.Println("Failed to get device interface name: ", err)
			return nil, err
		}

		supported, err := d.isSupportedDeviceType(device)
		if err != nil {
			return nil, err
		}

		if supported {
			devicesMap[ifaceName] = device
		}
	}

	log.Println("Successfully retrieved available devices")
	return devicesMap, nil
}

func (d *DeviceManager) isSupportedDeviceType(device nm.Device) (bool, error) {
	deviceType, err := device.GetPropertyDeviceType()
	if err != nil {
		log.Println("Failed to get device type: ", err)
		return false, err
	}

	if _, ok := supportedDeviceTypes[deviceType]; ok {
		return true, nil
	}

	return false, nil
}

func (d *DeviceManager) GetGSMDevice(name string) (nm.Device, error) {
	log.Printf("Retrieving GSM device: %s...\n", name)

	devices, err := d.nm.GetDevices()
	if err != nil {
		log.Println("Failed to get devices: ", err)
		return nil, err
	}

	for _, device := range devices {
		deviceType, err := device.GetPropertyDeviceType()
		if err != nil {
			log.Println("Failed to get device type: ", err)
			continue
		}

		// In gonetworkmanager, GSM devices are represented as Modem type
		if deviceType == nm.NmDeviceTypeModem {
			deviceName, err := device.GetPropertyInterface()
			if err != nil {
				log.Println("Failed to get device interface name: ", err)
				return nil, err
			}

			if deviceName == name {
				log.Println("Successfully retrieved GSM device: ", name)
				return device, nil
			}
		}
	}

	return nil, errors.New("Failed to retrieve GSM device: " + name)
}
