/*
 * Copyright © Siemens 2020 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"os"
	"strings"
	"time"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

// dict Dictionary type
type dict map[string]interface{}

// DBusDict Dictionary type
type DBusDict map[string]dbus.Variant

func parseStaticIPConfig(connection nm.ConnectionSettings) *v1.Interface_StaticConf {
	dict := connection[common.IPV4Key][common.AddressDataKey].([]map[string]interface{})

	config := &v1.Interface_StaticConf{}
	if len(dict) > 0 {
		if dict[0][common.AddressKey] != nil {
			config.IPv4 = dict[0][common.AddressKey].(string)
		}

		if dict[0][common.PrefixKey] != nil {
			config.NetMask = common.ParseNetMask(dict[0][common.PrefixKey].(uint32))
		}
		if connection[common.IPV4Key][common.GatewayKey] != nil {
			config.Gateway = connection[common.IPV4Key][common.GatewayKey].(string)
		}

	}
	return config
}

// Parse IPv4 Config
func parseDHCPIPv4Config(ipv4conf nm.IP4Config) *v1.Interface_StaticConf {

	config := &v1.Interface_StaticConf{}

	if ipv4conf != nil {
		ipv4Address, _ := ipv4conf.GetPropertyAddressData()
		if len(ipv4Address) > 0 {
			config.IPv4 = ipv4Address[0].Address
			config.NetMask = common.ParseNetMask(uint32(ipv4Address[0].Prefix))
		}
		config.Gateway, _ = ipv4conf.GetPropertyGateway()
	}
	return config
}

func parseDns(dnsArray []nm.IP4NameserverData) *v1.Interface_Dns {
	dns := &v1.Interface_Dns{}
	tmpDnsList := list.New()

	// copy nonempty members of dnsArray to tmpDnsArray list
	for _, dnsEntry := range dnsArray {
		if len(dnsEntry.Address) != 0 {
			tmpDnsList.PushBack(dnsEntry.Address)
		}
	}
	// Set PrimaryDns and SecondaryDns values, if this fields exists in dns entry list
	listElement := tmpDnsList.Front()
	if tmpDnsList.Len() > 0 {
		dns.PrimaryDNS = listElement.Value.(string)
		if tmpDnsList.Len() > 1 {
			listElement = listElement.Next()
			dns.SecondaryDNS = listElement.Value.(string)
		}
	}

	return dns
}

// parseRoutes parses the route data from the connection settings into a slice
// of Interface_Route.
func parseRoutes(routeArray []map[string]any) []*v1.Interface_Route {
	routes := []*v1.Interface_Route{}

	for _, routeEntry := range routeArray {
		route := parseRoute(routeEntry)
		routes = append(routes, route)
	}

	return routes
}

// parseRoute parses a single route entry from the connection settings.
// Like the rest of the connection settings parsing functions, it expects the
// route data to be in a specific format, which is a slice of maps with
// specific keys for destination, prefix, next hop, and metric and valid
// values: mandatory destination and prefix, and optional next hop and metric.
func parseRoute(routeEntry map[string]any) *v1.Interface_Route {
	getString := func(val any, key string) string {
		if val == nil {
			return ""
		}
		s, ok := val.(string)
		if !ok {
			log.Printf("Failed to cast %s to string: %v", key, val)
			return ""
		}
		return s
	}

	getUint32 := func(val any, key string) uint32 {
		if val == nil {
			return 0
		}
		u, ok := val.(uint32)
		if !ok {
			log.Printf("Failed to cast %s to uint32: %v", key, val)
			return 0
		}
		return u
	}

	dst := getString(routeEntry[common.DestinationKey], common.DestinationKey)
	prefix := getUint32(routeEntry[common.PrefixKey], common.PrefixKey)
	nextHop := getString(routeEntry[common.NextHopKey], common.NextHopKey)
	metric := getUint32(routeEntry[common.MetricKey], common.MetricKey)
	dstLen := net.IPv4len * 8

	return &v1.Interface_Route{
		Destination: dst,
		Netmask:     net.IP(net.CIDRMask(int(prefix), dstLen)).String(),
		NextHop:     nextHop,
		Metric:      metric,
	}
}

func listConnections(device nm.Device) []nm.Connection {
	var connections []nm.Connection

	interfaceName, err := device.GetPropertyInterface()
	if err != nil {
		log.Printf("Error getting interface interfaceName: %v", err)
	}

	availableConnectionsForDevice, err := device.GetPropertyAvailableConnections()
	if err != nil {
		log.Printf("Error getting available connections: %v", err)
	}

	activeConnection, err := device.GetPropertyActiveConnection()
	if err != nil {
		log.Printf("Error getting active connection: %v", err)
	}

	var activeConnectionUUID, activeConnectionId string
	if activeConnection != nil {
		activeConnectionUUID, activeConnectionId = getActiveConnectionDetails(activeConnection)
	}

	for _, connection := range availableConnectionsForDevice {
		if isValidConnection(connection, interfaceName, activeConnectionUUID, activeConnectionId) {
			connections = append(connections, connection)
		}
	}
	return connections
}

func getActiveConnectionDetails(activeConnection nm.ActiveConnection) (string, string) {
	activeConnectionUUID, err := activeConnection.GetPropertyUUID()
	if err != nil {
		log.Printf("Error getting UUID: %v", err)
	}

	activeConnectionId, err := activeConnection.GetPropertyID()
	if err != nil {
		log.Printf("Error getting ID: %v", err)
	}

	return activeConnectionUUID, activeConnectionId
}

func isValidConnection(connection nm.Connection, name, activeConnectionUUID, activeConnectionID string) bool {
	settings, _ := connection.GetSettings()
	interfaceName := settings[common.ConnectionKey][common.InterfaceNameKey]
	connectionType := settings[common.ConnectionKey][common.TypeKey]
	connectionUUID := settings[common.ConnectionKey][common.UUIDKey]
	connectionID := settings[common.ConnectionKey][common.IDKey]

	// This is an important check to ensure that this method and its callers
	// only process Ethernet connections.
	if connectionType != common.EthernetType {
		return false
	}

	if interfaceName != nil && interfaceName != "" {
		return interfaceName == name
	}

	return connectionUUID == activeConnectionUUID && connectionID == activeConnectionID
}

func DBusToProto(device nm.Device) (retVal *v1.Interface) {
	if device == nil {
		return nil
	}

	var values nm.ConnectionSettings
	var allConnections []nm.Connection
	var mac string

	deviceName, _ := device.GetPropertyInterface()
	deviceType, _ := device.GetPropertyDeviceType()
	var connectionType v1.Interface_InterfaceTypeEnum
	// Only Ethernet devices expose a reliable MAC via NetworkManager's wired device API.
	if deviceType == nm.NmDeviceTypeEthernet {
		var err error
		wired, errWired := toDeviceWired(device.GetPath())
		if errWired != nil {
			log.Printf("Error creating wired: %v", errWired)
		}
		mac, err = wired.GetPropertyHwAddress()
		if err != nil {
			log.Printf("Error getting hardware address: %v", err)
		}
		connectionType = v1.Interface_ETHERNET
	}

	if deviceType == nm.NmDeviceTypeModem {
		connectionType = v1.Interface_GSM
	}
	conn, err := device.GetPropertyActiveConnection()
	allConnections = listConnections(device)
	if err == nil && conn != nil {
		log.Printf("Found active connection for device: %s", deviceName)
		ipv4, _ := conn.GetPropertyIP4Config()
		props, _ := conn.GetPropertyConnection()
		values, _ = props.GetSettings()
		retVal = convertToProto(values, ipv4, mac)

	} else if len(allConnections) > 0 {
		log.Printf("Found %d active connections for device: %s", len(allConnections), deviceName)
		values, _ = allConnections[0].GetSettings()
		retVal = convertToProto(values, nil, mac)

	} else {
		log.Printf("No active connections for device: %s", deviceName)
		retVal = &v1.Interface{
			MacAddress:    mac,
			InterfaceName: deviceName,
		}
	}

	// Common fields
	retVal.InterfaceName = deviceName
	retVal.Label, _ = getLabelForInterface(deviceName)
	retVal.InterfaceType = &connectionType // e.g., "ethernet" or "gsm"

	// L2Conf is derived from Docker macvlan and is only applicable for Ethernet interfaces here.
	if deviceType == nm.NmDeviceTypeEthernet {
		retVal.L2Conf = dockerNetworkGetMacvlanConnection(deviceName)
	}

	// GSM-specific config
	if deviceType == nm.NmDeviceTypeModem {
		retVal.GsmConfiguration = extractGsmConf(values)
	}

	return retVal
}

func toDeviceWired(objectPath dbus.ObjectPath) (nm.DeviceWired, error) {
	return nm.NewDeviceWired(objectPath)
}

func extractGsmConf(values nm.ConnectionSettings) *v1.Interface_GsmConf {
	gsm, ok := values[common.GSMSetting]
	if !ok {
		return nil
	}

	conf := &v1.Interface_GsmConf{}

	// APN is NOT masked
	if v, ok := gsm[common.APNKey].(string); ok {
		conf.Apn = v
	}

	// Sensitive fields → masked
	//the gonetworkmanager library (v2.2.0) does not deliver back any key for PIN and PASSWORD ,
	// so this maskValue is just a safety mechanism in case this changes
	if v, ok := gsm[common.PINKey].(string); ok {
		conf.Pin = maskValue(v)
	}

	if v, ok := gsm[common.UsernameKey].(string); ok {
		conf.Username = v
	}

	if v, ok := gsm[common.PasswordKey].(string); ok {
		conf.Password = maskValue(v)
	}

	return conf
}

//

func maskValue(v string) string {
	if v == "" {
		return ""
	}
	return "****"
}

// Converts DBus data (nm.ConnectionSettings) to Device Model Proto
func convertToProto(connection nm.ConnectionSettings, ipv4Config nm.IP4Config, mac string) *v1.Interface {

	retVal := &v1.Interface{}
	retVal.MacAddress = strings.ToUpper(mac)

	if connection[common.IPV4Key][common.MethodKey] == common.Auto {
		retVal.DHCP = common.Enabled
		retVal.Static = parseDHCPIPv4Config(ipv4Config)
	} else {
		retVal.Static = parseStaticIPConfig(connection)
		retVal.DHCP = common.Disabled
	}

	if ipv4Config != nil {
		dnsArray, _ := ipv4Config.GetPropertyNameserverData()
		retVal.DNSConfig = parseDns(dnsArray)
	}

	if ipv4, have := connection[common.IPV4Key]; have {
		if rawRouteData, have := ipv4[common.RouteDataKey]; have {
			if routeData, ok := rawRouteData.([]map[string]any); ok {
				retVal.Routes = parseRoutes(routeData)
			} else {
				log.Println("failed to cast routes:", mac)
			}
		}
	}

	return retVal
}

// newSettingsFromProto creates new NetworkManager->ConnectionSettings from given device model proto data.
// It takes a v1.Interface and a deviceName as parameters and returns a nm.ConnectionSettings.
func newSettingsFromProto(protoData *v1.Interface, deviceName string) nm.ConnectionSettings {
	connection := initializeConnectionSettings()
	ipAssignmentMethod := determineIpAssignmentMethod(protoData)
	applyConnectionSetting(ipAssignmentMethod, protoData, connection)
	identifier := determineIdentifier(protoData)
	setConnectionDetails(connection, protoData, identifier, ipAssignmentMethod, deviceName)

	return connection
}

// determineIpAssignmentMethod determines the connection suffix based on the protoData.
func determineIpAssignmentMethod(protoData *v1.Interface) string {
	if protoData.DHCP == common.Enabled {
		return common.DHCP
	}
	return common.Static
}

// applyConnectionSetting applies the connection settings.
func applyConnectionSetting(connectionSuffix string, protoData *v1.Interface, connection nm.ConnectionSettings) {
	if protoData.GatewayInterface {
		connection[common.IPV4Key][common.RouteMetricKey] = 1
	}
	if connectionSuffix == common.DHCP {
		putDHCP(connection)
	} else {
		putStaticIP(protoData, connection)
	}

	putDNSConfig(protoData, connection)

	putExtraRoutes(protoData, connection)
}

// ConfigureExistingGatewayInterfacesExceptProtoData sets the route metric for all Ethernet device connections
// if the GatewayInterface flag is enabled in the provided protoData.
func ConfigureExistingGatewayInterfacesExceptProtoData(protoData *v1.Interface, networkConfigurator NetworkConfigurator) error {
	if !protoData.GatewayInterface {
		return nil
	}

	allEthernetDevices := networkConfigurator.getAllEthernetDevices()

	for _, ethernetDevice := range allEthernetDevices {
		if err := setGatewayInterfaceForDeviceConnections(ethernetDevice, protoData, networkConfigurator); err != nil {
			return err
		}
	}

	return nil
}

// setGatewayInterfaceForDeviceConnections iterates over all connections of the given Ethernet device
// and sets the route metric for each connection based on the provided protoData.
func setGatewayInterfaceForDeviceConnections(ethernetDevice nm.DeviceWired, protoData *v1.Interface, networkConfigurator NetworkConfigurator) error {
	for _, connection := range listConnections(ethernetDevice) {
		if err := checkAndUpdateGatewayInterfaceForConnection(connection, protoData, ethernetDevice, networkConfigurator); err != nil {
			return err
		}
	}
	return nil
}

// deactivateAndActivateConnection deactivates the current active connection of the Ethernet device
// and activates the provided connection. Logs the process and handles errors appropriately.
func deactivateAndActivateConnection(ethernetDevice nm.DeviceWired, connection nm.Connection, networkConfigurator NetworkConfigurator) error {
	activeConnection, err := ethernetDevice.GetPropertyActiveConnection()
	if err != nil {
		return fmt.Errorf("failed to get active connection: %w", err)
	}

	if activeConnection == nil {
		ethernetDeviceMacAddress, _ := ethernetDevice.GetPropertyHwAddress()
		log.Printf("No active connection found for device: %v", ethernetDeviceMacAddress)
		return nil
	}

	if err = networkConfigurator.gnm.DeactivateConnection(activeConnection); err != nil {
		return fmt.Errorf("failed to deactivate connection: %w", err)
	}

	if _, err = networkConfigurator.gnm.ActivateConnection(connection, ethernetDevice, nil); err != nil {
		return fmt.Errorf("failed to activate connection: %w", err)
	}

	log.Println("Connection successfully reactivated")
	return nil
}

// checkAndUpdateGatewayInterfaceForConnection updates the route metric for the given connection
// if the MAC address or label in protoData does not match the current settings.
func checkAndUpdateGatewayInterfaceForConnection(connection nm.Connection, protoData *v1.Interface, ethernetDevice nm.DeviceWired,
	networkConfigurator NetworkConfigurator) error {
	settings, err := connection.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings for connection: %w", err)
	}

	if settings[common.EthernetType][common.MACAddressKey] == nil {
		if err := setMacAddressInSettings(settings, ethernetDevice); err != nil {
			return err
		}
	}

	macStr := getMacAddressFromSettings(settings)
	if willGatewayInterfaceBeUpdated(protoData, macStr, settings) {
		if err := changePriorityOfGatewayInterface(settings, connection); err != nil {
			return err
		}
		err := deactivateAndActivateConnection(ethernetDevice, connection, networkConfigurator)
		if err != nil {
			return err
		}
	}

	return nil
}

// setMacAddressInSettings sets the MAC address in the connection settings
// by retrieving the permanent hardware address from the Ethernet device.
func setMacAddressInSettings(settings nm.ConnectionSettings, ethernetDevice nm.DeviceWired) error {
	retValue, _ := ethernetDevice.GetPropertyPermHwAddress()
	if retValue == "" {
		retValue, _ = ethernetDevice.GetPropertyHwAddress()
	}
	macAddr, err := net.ParseMAC(retValue)
	if err != nil {
		return err
	}
	settings[common.EthernetType][common.MACAddressKey] = []uint8(macAddr)
	return nil
}

// willGatewayInterfaceBeUpdated checks if the route metric needs to be updated
// based on the MAC address or label in the provided protoData.
func willGatewayInterfaceBeUpdated(protoData *v1.Interface, macStr string, settings nm.ConnectionSettings) bool {
	if protoData.MacAddress != "" {
		return strings.ToUpper(protoData.MacAddress) != macStr
	} else if protoData.Label != "" {
		return strings.ToLower(getInterfaceForLabel(protoData.Label)) != settings[common.ConnectionKey][common.InterfaceNameKey]
	}
	return false
}

// getMacAddressFromSettings retrieves the MAC address from the connection settings
// and returns it as a formatted string.
func getMacAddressFromSettings(settings nm.ConnectionSettings) string {
	mac := settings[common.EthernetType][common.MACAddressKey].([]byte)
	macStr := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])

	return strings.ToUpper(macStr)
}

// changePriorityOfGatewayInterface updates the route metric in the connection settings
// and removes IPv6 addresses and routes. Logs the update process.
func changePriorityOfGatewayInterface(settings nm.ConnectionSettings, connection nm.Connection) error {
	settings[common.IPV4Key][common.RouteMetricKey] = int32(-1)

	delete(settings["ipv6"], "addresses")
	delete(settings["ipv6"], "routes")

	err := connection.Update(settings)
	if err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	log.Printf("Connection ID: %v has been updated with Route Metric value: %v", settings[common.ConnectionKey][common.IDKey],
		settings[common.IPV4Key][common.RouteMetricKey])

	return nil
}

// initializeConnectionSettings initializes the connection settings.
func initializeConnectionSettings() nm.ConnectionSettings {
	return nm.ConnectionSettings{
		common.ConnectionKey: make(dict),
		common.IPV4Key:       make(dict),
		common.EthernetType:  make(dict),
	}
}

// putDHCP puts the DHCP configuration.
func putDHCP(connection nm.ConnectionSettings) {
	connection[common.IPV4Key][common.MethodKey] = common.Auto
}

// putStaticIP puts the static IP configuration.
func putStaticIP(protoData *v1.Interface, connection nm.ConnectionSettings) {
	connection[common.IPV4Key][common.MethodKey] = common.Manual
	if protoData.Static != nil {
		if protoData.Static.Gateway != "" {
			connection[common.IPV4Key][common.GatewayKey] = protoData.Static.Gateway
		}
		address := dbus.MakeVariantWithSignature(protoData.Static.IPv4, dbus.ParseSignatureMust("s"))
		prefix := dbus.MakeVariantWithSignature(common.ParseNetMaskSize(protoData.Static.NetMask), dbus.ParseSignatureMust("u"))

		ipDict := make(DBusDict)
		ipDict[common.AddressKey] = address // IP address, e.g: "192.168.0.1"
		ipDict[common.PrefixKey] = prefix   // Subnet, e.g: 24

		connection[common.IPV4Key][common.AddressDataKey] = []DBusDict{ipDict}
	}
}

// putDNSConfig puts the DNS configuration.
func putDNSConfig(protoData *v1.Interface, connection nm.ConnectionSettings) {
	if protoData.DNSConfig != nil {
		var dns1, dns2 uint32
		if len(protoData.DNSConfig.PrimaryDNS) > 0 {
			dns1 = common.IPToUInt32LI(protoData.DNSConfig.PrimaryDNS)
			connection[common.IPV4Key][common.DNSKey] = []uint32{dns1}

			if len(protoData.DNSConfig.SecondaryDNS) > 0 {
				dns2 = common.IPToUInt32LI(protoData.DNSConfig.SecondaryDNS)
				connection[common.IPV4Key][common.DNSKey] = []uint32{dns1, dns2}
			}
		}

		if protoData.DHCP == common.Enabled {
			connection[common.IPV4Key][common.DNSIgnoreAutoKey] = common.Yes
		}
	}
}

// putExtraRoutes puts additional IPv4 routes into the connection settings.
// It expects the route data in protoData has already been validated and in
// the correct format.
func putExtraRoutes(protoData *v1.Interface, connection nm.ConnectionSettings) {
	var routes []DBusDict

	for _, protoRoute := range protoData.GetRoutes() {
		route := buildRouteDict(protoRoute)
		routes = append(routes, route)
	}

	if len(routes) > 0 {
		ipv4Map := connection[common.IPV4Key]
		if ipv4Map == nil {
			ipv4Map = make(map[string]any)
			connection[common.IPV4Key] = ipv4Map
		}
		ipv4Map[common.RouteDataKey] = routes
	}
}

// buildRouteDict creates a DBusDict for a single route. It expects protoRoute
// to have been validated beforehand and contain the necessary fields.
func buildRouteDict(protoRoute *v1.Interface_Route) DBusDict {
	route := make(DBusDict)

	setRouteDestination(route, protoRoute.GetDestination())
	setRoutePrefix(route, protoRoute.GetNetmask())
	setRouteNextHop(route, protoRoute.GetNextHop())
	setRouteMetric(route, protoRoute.GetMetric())

	return route
}

// setRouteDestination sets the destination IP address for the route. It expects
// the destination to be a valid IP address string.
func setRouteDestination(route DBusDict, dst string) {
	route["dest"] = dbus.MakeVariant(net.ParseIP(dst).String())
}

// setRoutePrefix sets the prefix size for the route. It expects the netmask to be
// a valid CIDR notation string (e.g., "255.255.255.0")
func setRoutePrefix(route DBusDict, netmask string) {
	route["prefix"] = dbus.MakeVariant(common.ParseNetMaskSize(netmask))
}

// setRouteNextHop sets the next hop IP address for the route. It expects the next hop
// to be empty or a valid IP address string. If the next hop is empty, it will not set this field.
func setRouteNextHop(route DBusDict, nh string) {
	if len(nh) > 0 {
		if net.ParseIP(nh) != nil {
			route["next-hop"] = dbus.MakeVariant(nh)
		}
	}
}

// setRouteMetric sets the metric for the route.
func setRouteMetric(route DBusDict, metric uint32) {
	route["metric"] = dbus.MakeVariant(metric)
}

// determineIdentifier determines the identifier for the connection ID.
func determineIdentifier(protoData *v1.Interface) string {
	identifier := ""
	if protoData.Label != "" {
		identifier = protoData.Label
	} else if protoData.MacAddress != "" {
		identifier = protoData.MacAddress
	}
	return identifier
}

// setConnectionDetails sets the connection ID, UUID, and timestamp.
func setConnectionDetails(connection nm.ConnectionSettings, protoData *v1.Interface, identifier string, connectionSuffix string, deviceName string) {
	connection[common.ConnectionKey][common.IDKey] = fmt.Sprintf("%s_%s", identifier, connectionSuffix)
	connection[common.ConnectionKey][common.UUIDKey] = uuid.New().String()
	connection[common.ConnectionKey][common.TimeStampKey] = time.Now().Unix()
	connection[common.ConnectionKey][common.TypeKey] = common.EthernetType
	connection[common.ConnectionKey][common.InterfaceNameKey] = deviceName

	putMACAddress(protoData, connection)
}

// putMACAddress puts the MAC address and sets the constants.MACAddressKey.
func putMACAddress(protoData *v1.Interface, connection nm.ConnectionSettings) {
	uintMac, err := net.ParseMAC(protoData.MacAddress)
	if err != nil {
		log.Printf("Error parsing MAC address: %v", err)
	}
	connection[common.EthernetType][common.MACAddressKey] = uintMac
}

func GetMapWithUppercase(inputMap map[string]string) map[string]string {
	outputMap := make(map[string]string)
	for key, value := range inputMap {
		outputMap[strings.ToUpper(key)] = strings.ToUpper(value)
	}

	return outputMap
}

func WriteMapToFile(mapToBeWritten map[string]string, fileName string) error {
	buffer, err := json.Marshal(GetMapWithUppercase(mapToBeWritten))

	if err == nil {
		err = os.WriteFile(fileName, buffer, 0666) //ioutil.WriteFile(fileName, buffer, 0666)
	}

	return err
}

func readMapFromFile(fileName string) (map[string]string, error) {

	var parsedMap map[string]string
	buffer, err := os.ReadFile(fileName) //ioutil.ReadFile(fileName)

	if err == nil {
		err = json.Unmarshal(buffer, &parsedMap)
	}

	return parsedMap, err
}

func getInterfaceForLabel(label string) string {
	var interfaceName string
	labelMap, err := readMapFromFile(common.LabelMapFileName)

	if err == nil {
		interfaceName = labelMap[strings.ToUpper(label)]
	} else {
		log.Println(err)
	}

	return interfaceName
}

func getLabelForInterface(interfaceName string) (string, error) {
	labelMap, err := readMapFromFile(common.LabelMapFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read label map from file: %w", err)
	}

	upperInterfaceName := strings.ToUpper(interfaceName)
	for label, value := range labelMap {
		if value == upperInterfaceName {
			return label, nil
		}
	}

	return "", fmt.Errorf("interface not found: %s", interfaceName)
}

func isGSMInterface(iface *v1.Interface) bool {
	return iface.GetGsmConfiguration() != nil || iface.GetInterfaceType() == v1.Interface_GSM
}

func gsmInterfaceMatches(i1 *v1.Interface, i2 *v1.Interface) bool {
	apn1 := i1.GetGsmConfiguration().GetApn()
	apn2 := i2.GetGsmConfiguration().GetApn()

	username1 := i1.GetGsmConfiguration().GetUsername()
	username2 := i2.GetGsmConfiguration().GetUsername()

	return apn1 == apn2 && username1 == username2
}

func ethernetInterfaceMatches(i1 *v1.Interface, i2 *v1.Interface) bool {
	mac1 := strings.ToUpper(i1.GetMacAddress())
	mac2 := strings.ToUpper(i2.GetMacAddress())

	label1 := strings.ToLower(i1.GetLabel())
	label2 := strings.ToLower(i2.GetLabel())

	return (mac2 != "" && mac1 == mac2) || (label2 != "" && label1 == label2)
}

// isInterfaceConfigured checks whether an interface has a valid network configuration.
// Returns true if:
//   - it is a GSM interface (with GSM config or GSM type), OR
//   - it has DHCP enabled, OR
//   - it has a valid static IP address
//
// Returns false for unconfigured interfaces (e.g., disconnected Ethernet with no
// DHCP and no static IP). Processing such interfaces causes NetworkManager errors:
// "ipv4.addresses: this property cannot be empty for 'method=manual'"
func isInterfaceConfigured(iface *v1.Interface) bool {
	if iface == nil {
		return false
	}

	if isGSMInterface(iface) {
		return true
	}

	if iface.DHCP == common.Enabled {
		return true
	}

	if iface.Static != nil && iface.Static.IPv4 != "" {
		return true
	}

	return false
}
