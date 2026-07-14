/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package connectionmanager

import (
	"log"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/interfaces"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

type ConnUtils struct{}

func NewConnectionUtils() interfaces.ConnectionUtils {
	return &ConnUtils{}
}

func (c *ConnUtils) ConfigureGateway(connSettings nm.ConnectionSettings, priority common.RouteMetricPriority) {
	log.Println("Configuring gateway route metric...")

	// Only if ipv4 settings does not exist, create the map
	if _, ok := connSettings[common.IPV4Key]; !ok || connSettings[common.IPV4Key] == nil {
		connSettings[common.IPV4Key] = make(map[string]any)
	}
	connSettings[common.IPV4Key][common.RouteMetricKey] = priority

	// Remove IPv6 addresses and routes to avoid possible type mismatch issues
	// when NetworkManager tries to apply the settings.
	// Since GSM & Ethernet connections do not use IPv6 configurations, this is acceptable.
	c.cleanIPv6Settings(connSettings)

	log.Println("Gateway route metric configured successfully")
}

func (c *ConnUtils) cleanIPv6Settings(connSettings nm.ConnectionSettings) {
	if connSettings[common.IPV6Key] != nil {
		delete(connSettings[common.IPV6Key], common.AddressesKey)
		delete(connSettings[common.IPV6Key], common.RoutesKey)
	}
}

func (c *ConnUtils) RetrieveDeviceName(settings nm.ConnectionSettings) string {
	log.Println("Retrieving device name for connection...")
	connMap, ok := settings[common.ConnectionKey]
	if !ok || connMap == nil {
		log.Println(common.LogNoConnectionKey)
		return ""
	}

	if interfaceName, ok := connMap[common.InterfaceNameKey].(string); ok {
		log.Println("Retrieved device name for connection")
		return interfaceName
	}
	return ""
}

func (c *ConnUtils) ConfigureDNS(connSettings nm.ConnectionSettings, dnsConfig *v1.Interface_Dns) {
	if dnsConfig != nil {
		log.Println("Configuring DNS for connection...")

		// Only if ipv4 settings does not exist, create the map
		if ipv4Map, ok := connSettings[common.IPV4Key]; !ok || ipv4Map == nil {
			connSettings[common.IPV4Key] = make(map[string]any)
		}

		if len(dnsConfig.PrimaryDNS) > 0 {
			log.Println("Setting Primary DNS for connection")
			primaryDNS := common.IPToUInt32LI(dnsConfig.PrimaryDNS)
			connSettings[common.IPV4Key][common.DNSKey] = []uint32{primaryDNS}

			if len(dnsConfig.SecondaryDNS) > 0 {
				log.Println("Setting Secondary DNS for connection")
				secondaryDNS := common.IPToUInt32LI(dnsConfig.SecondaryDNS)
				connSettings[common.IPV4Key][common.DNSKey] = []uint32{primaryDNS, secondaryDNS}
			}
		}

		// Ignore automatic DNS settings from DHCP
		log.Println("Setting to ignore automatic DNS from DHCP for connection")
		connSettings[common.IPV4Key][common.DNSIgnoreAutoKey] = 1

		log.Println("DNS configuration for connection completed")
		return
	}

	log.Println("No DNS configuration provided, skipping DNS setup for connection")
}
