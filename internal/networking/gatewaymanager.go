/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"log"

	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/devicemanager"
	"networkservice/internal/networking/interfaces"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

type GatewayHelper interface {
	Reset(string) error
}

type GatewayManager struct {
	nm            nm.NetworkManager
	deviceName    string
	connManager   interfaces.ConnectionHelper
	connUtils     interfaces.ConnectionUtils
	deviceManager interfaces.DeviceHelper
}

func NewGatewayManager(nm nm.NetworkManager) GatewayHelper {
	return &GatewayManager{
		nm:            nm,
		connManager:   connectionmanager.NewConnectionManager(nm),
		connUtils:     connectionmanager.NewConnectionUtils(),
		deviceManager: devicemanager.NewDeviceManager(nm),
	}
}

func (g *GatewayManager) Reset(interfaceName string) error {
	log.Println("Resetting gateway for device: ", interfaceName)

	devicesMap, err := g.deviceManager.GetAvailableDevicesMap()
	if err != nil {
		return err
	}

	device, exists := devicesMap[interfaceName]
	if !exists {
		log.Println("Device not found: ", interfaceName)
		return nil
	}

	conn, activated, err := g.connManager.FindConnectionWithStatus(device)
	if err != nil {
		return err
	}
	if !activated {
		if err := g.connManager.ActivateConnectionWithRetry(conn, device); err != nil {
			return err
		}
	}

	connSettings, err := g.connManager.GetConnectionSettings(conn)
	if err != nil {
		return err
	}

	log.Println("Resetting route metric for previous gateway...")
	g.connUtils.ConfigureGateway(connSettings, common.RouteMetricLowestPriority)

	log.Println("Updating connection settings for previous gateway")
	if err := g.connManager.UpdateConnection(conn, connSettings); err != nil {
		return err
	}

	// Reactivate the connection to apply the new settings,
	// specifically the updated route metric for the gateway
	log.Println("Activating connection for previous gateway")
	if err := g.connManager.ActivateConnectionWithRetry(conn, device); err != nil {
		return err
	}

	log.Println("Successfully reset gateway for previous gateway")
	return nil
}
