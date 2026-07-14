/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"errors"
	"log"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/devicemanager"
	"networkservice/internal/networking/interfaces"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

type GSMConfigurator struct {
	isGateway     bool
	gsmConfig     *v1.Interface_GsmConf
	dnsConfig     *v1.Interface_Dns
	connManager   interfaces.ConnectionHelper
	connUtils     interfaces.ConnectionUtils
	deviceManager interfaces.DeviceHelper
}

func NewGSMConfigurator(networkManager nm.NetworkManager, iface *v1.Interface) *GSMConfigurator {
	log.Println("Initializing GSM configurator to update interface settings")

	gsmConfig := iface.GetGsmConfiguration()
	dnsConfig := iface.GetDNSConfig()
	isGateway := iface.GetGatewayInterface()

	return &GSMConfigurator{
		isGateway:     isGateway,
		gsmConfig:     gsmConfig,
		dnsConfig:     dnsConfig,
		connManager:   connectionmanager.NewConnectionManager(networkManager),
		connUtils:     connectionmanager.NewConnectionUtils(),
		deviceManager: devicemanager.NewDeviceManager(networkManager),
	}
}

func (c *GSMConfigurator) Configure() error {
	log.Println("Configuring GSM connection...")

	// Using the APN and Username (if provided), to identify the correct GSM connection
	// since during the configuration of a GSM during on-boarding, we do not have direct access to other unique identifiers,
	// like - connection name and device name i.e., connection ID and connection interface name respectively.
	// Although unlikely, if multiple connections exist with the same APN and Username,
	// only the first matching connection will be returned.
	conn, connSettings, err := c.retrievingConnectionDetails(c.gsmConfig.GetApn(), c.gsmConfig.GetUsername())
	if err != nil {
		return err
	}

	if conn == nil || connSettings == nil {
		log.Println("no GSM connection found")
		return errors.New("no GSM connection found; use CreateConnection RPC first")
	}

	deviceName := c.connUtils.RetrieveDeviceName(connSettings)

	device, err := c.deviceManager.GetGSMDevice(deviceName)
	if err != nil {
		return err
	}

	if c.isGateway {
		log.Println("Configuring GSM connection as gateway...")
		c.connUtils.ConfigureGateway(connSettings, common.RouteMetricHighestPriority)
	}

	log.Println("Configuring DNS for GSM connection...")
	c.connUtils.ConfigureDNS(connSettings, c.dnsConfig)

	log.Println("Updating GSM connection settings...")
	if err := c.connManager.UpdateConnection(conn, connSettings); err != nil {
		return err
	}

	// Activate the connection to apply the updated settings to the device.
	// This step also updates the IP routing table when the GSM interface is set as the default gateway.
	// Retrying activation here is a convenience, since returning an error when the connection is down
	// (e.g., in 'disconnected' or 'failed' states) would require the caller to handle activation separately,
	// which is not feasible externally.
	log.Println("Activating GSM connection with new settings")
	if err := c.connManager.ActivateConnectionWithRetry(conn, device); err != nil {
		return err
	}

	log.Println("GSM connection configured successfully")
	return nil
}

func (c *GSMConfigurator) retrievingConnectionDetails(apn, username string) (nm.Connection, nm.ConnectionSettings, error) {
	log.Println("Retrieving connection for APN: ", apn)

	nmSettings, err := c.newNetworkManagerSettings()
	if err != nil {
		log.Println("Failed to create a NetworkManager settings instance: ", err)
		return nil, nil, err
	}

	connections, err := nmSettings.ListConnections()
	if err != nil {
		log.Println("Failed to list connections: ", err)
		return nil, nil, err
	}

	for _, connection := range connections {
		connSettings, err := connection.GetSettings()
		if err != nil {
			log.Println("Failed to get connection settings: ", err)
			continue
		}

		if c.isGSMConnection(connSettings) {
			if c.isConnectionMatch(connSettings, apn, username) {
				log.Println("Connection retrieved for APN: ", apn)
				return connection, connSettings, nil
			}
		}
	}

	log.Println("No matching connection found for APN: ", apn)
	return nil, nil, nil
}

func (c *GSMConfigurator) newNetworkManagerSettings() (nm.Settings, error) {
	return nm.NewSettings()
}

func (c *GSMConfigurator) isGSMConnection(settings nm.ConnectionSettings) bool {
	connMap, ok := settings[common.ConnectionKey]
	if !ok || connMap == nil {
		log.Println(common.LogNoConnectionKey)
		return false
	}

	if connType, ok := connMap[common.TypeKey].(string); ok {
		if connType == common.GSMSetting {
			log.Println("Connection is of GSM type")
			return true
		}

		log.Println("Connection is not of GSM type")
		return false
	}
	return false
}

func (c *GSMConfigurator) isConnectionMatch(settings nm.ConnectionSettings, apn, username string) bool {
	existingApn := c.retrieveApn(settings)
	existingUsername := c.retrieveUsername(settings)

	if existingApn == apn && existingUsername == username {
		return true
	}
	log.Println("Connection not matched")
	return false
}

func (c *GSMConfigurator) retrieveApn(settings nm.ConnectionSettings) string {
	gsmMap, ok := settings[common.GSMSetting]
	if !ok || gsmMap == nil {
		log.Println("Connection settings do not have GSM key")
		return ""
	}

	if apn, ok := gsmMap[common.APNKey].(string); ok {
		log.Println("Retrieved APN for existing connection")
		return apn
	}
	return ""
}

func (c *GSMConfigurator) retrieveUsername(settings nm.ConnectionSettings) string {
	gsmMap, ok := settings[common.GSMSetting]
	if !ok || gsmMap == nil {
		log.Println("Connection settings do not have GSM key")
		return ""
	}

	if username, ok := gsmMap[common.UsernameKey].(string); ok {
		log.Println("Retrieved username for existing connection")
		return username
	}
	return ""
}
