/*
 * Copyright (c) 2021 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"container/list"
	"encoding/json"
	"fmt"
	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"strings"
	"time"
)

// dict Dictionary type
type dict map[string]interface{}

// DBusDict Dictionary type
type DBusDict map[string]dbus.Variant

// Backup configuration can not applied to dbus! some fields needs to be removed Create a new connection
// INSTANCE based on ipv4 and name from backup
func retrieveSettingsFromBackup(backup nm.ConnectionSettings) nm.ConnectionSettings {
	connection := make(nm.ConnectionSettings)
	connection[ConnectionKey] = make(dict)
	connection[IPV4Key] = make(dict)
	connection[EthernetType] = make(dict)
	connection[ConnectionKey][IDKey] = backup[ConnectionKey][IDKey]
	connection[ConnectionKey][TypeKey] = backup[ConnectionKey][TypeKey]
	connection[ConnectionKey][InterfaceNameKey] = backup[ConnectionKey][InterfaceNameKey]
	connection[ConnectionKey][UUIDKey] = uuid.New().String()
	connection[ConnectionKey][TimeStampKey] = time.Now().UnixNano()
	connection[EthernetType] = backup[EthernetType]
	connection[EthernetType][MACAddressKey] = backup[EthernetType][MACAddressKey]
	connection[IPV4Key] = backup[IPV4Key]
	return connection
}

func parseStaticIPConfig(connection nm.ConnectionSettings) *v1.Interface_StaticConf {
	dict := connection[IPV4Key][AddressDataKey].([]map[string]interface{})

	config := &v1.Interface_StaticConf{}
	if len(dict) > 0 {
		if dict[0][AddressKey] != nil {
			config.IPv4 = dict[0][AddressKey].(string)
		}

		if dict[0][PrefixKey] != nil {
			config.NetMask = ParseNetMask(dict[0][PrefixKey].(uint32))
		}
		if connection[IPV4Key][GatewayKey] != nil {
			config.Gateway = connection[IPV4Key][GatewayKey].(string)
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
			config.NetMask = ParseNetMask(uint32(ipv4Address[0].Prefix))
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

func listConnections(device nm.Device) []nm.Connection {

	var connections []nm.Connection
	val, _ := nm.NewSettings()
	//do not change
	//active and available connections functions does not include all connections under a device.
	allConnections, _ := val.ListConnections()
	name, _ := device.GetPropertyInterface()

	if allConnections != nil {
		for _, element := range allConnections {
			settings, _ := element.GetSettings()
			if settings[ConnectionKey][InterfaceNameKey] == name &&
				settings[ConnectionKey][TypeKey] == EthernetType {
				connections = append(connections, element)
			}
		}
	}
	return connections
}

func DBusToProto(device nm.DeviceWired) *v1.Interface {
	if device == nil {
		return nil
	}

	//gnm,_:=nm.NewNetworkManager()
	var values nm.ConnectionSettings
	var retVal *v1.Interface
	var allConnections []nm.Connection

	conn, err := device.GetPropertyActiveConnection()
	allConnections = listConnections(device)

	mac, _ := device.GetPropertyHwAddress()
	deviceName, _ := device.GetPropertyInterface()

	if err == nil && conn != nil {
		IPv4Wrapper, _ := conn.GetPropertyIP4Config()
		props, _ := conn.GetPropertyConnection()
		values, _ = props.GetSettings()
		retVal = convertToProto(values, IPv4Wrapper, mac)
	} else if allConnections != nil && len(allConnections) > 0 {
		values, _ = allConnections[0].GetSettings()
		retVal = convertToProto(values, nil, mac)
	} else {
		retVal = &v1.Interface{MacAddress: mac}
		retVal = &v1.Interface{Label: deviceName}
	}

	// get device interface name
	interfaceName, _ := device.GetPropertyInterface()
	log.Println("interfacename :", interfaceName)

	// get layer2 config from device
	l2device := dockerNetworkGetMacvlanConnection(interfaceName)
	retVal.L2Conf = l2device

	retVal.InterfaceName = interfaceName
	retVal.Label = getLabelForInterface(interfaceName)

	return retVal
}

// Converts DBus data (nm.ConnectionSettings) to Device Model Proto
func convertToProto(connection nm.ConnectionSettings, ipv4Config nm.IP4Config, mac string) *v1.Interface {

	retVal := &v1.Interface{}
	retVal.MacAddress = strings.ToUpper(mac)

	if connection[IPV4Key][MethodKey] == Auto {
		retVal.DHCP = Enabled
		retVal.Static = parseDHCPIPv4Config(ipv4Config)
	} else {
		retVal.Static = parseStaticIPConfig(connection)
		retVal.DHCP = Disabled
	}

	if val, ok := connection[IPV4Key]; ok {
		if v, o := val[RouteMetricKey]; o {
			if v != nil && v.(int64) == 1 {
				retVal.GatewayInterface = true
			}
		}
	}

	//if connection[IPV4Key][RouteMetricKey] != nil && connection[IPV4Key][RouteMetricKey].(int64) == 1 {
	//	retVal.GatewayInterface = true
	//}

	if ipv4Config != nil {
		dnsArray, _ := ipv4Config.GetPropertyNameserverData()
		retVal.DNSConfig = parseDns(dnsArray)
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
	if protoData.DHCP == Enabled {
		return DHCP
	}
	return Static
}

// applyConnectionSetting applies the connection settings.
func applyConnectionSetting(connectionSuffix string, protoData *v1.Interface, connection nm.ConnectionSettings) {
	if protoData.GatewayInterface {
		connection[IPV4Key][RouteMetricKey] = 1
	}
	if connectionSuffix == DHCP {
		putDHCP(connection)
	} else {
		putStaticIP(protoData, connection)
	}

	putDNSConfig(protoData, connection)
}

// initializeConnectionSettings initializes the connection settings.
func initializeConnectionSettings() nm.ConnectionSettings {
	return nm.ConnectionSettings{
		ConnectionKey: make(dict),
		IPV4Key:       make(dict),
		EthernetType:  make(dict),
	}
}

// putDHCP puts the DHCP configuration.
func putDHCP(connection nm.ConnectionSettings) {
	connection[IPV4Key][MethodKey] = Auto
}

// putStaticIP puts the static IP configuration.
func putStaticIP(protoData *v1.Interface, connection nm.ConnectionSettings) {
	connection[IPV4Key][MethodKey] = Manual
	if protoData.Static != nil {
		if protoData.Static.Gateway != "" {
			connection[IPV4Key][GatewayKey] = protoData.Static.Gateway
		}
		address := dbus.MakeVariantWithSignature(protoData.Static.IPv4, dbus.ParseSignatureMust("s"))
		prefix := dbus.MakeVariantWithSignature(ParseNetMaskSize(protoData.Static.NetMask), dbus.ParseSignatureMust("u"))

		ipDict := make(DBusDict)
		ipDict[AddressKey] = address // IP address, e.g: "192.168.0.1"
		ipDict[PrefixKey] = prefix   // Subnet, e.g: 24

		connection[IPV4Key][AddressDataKey] = []DBusDict{ipDict}
	}
}

// putDNSConfig puts the DNS configuration.
func putDNSConfig(protoData *v1.Interface, connection nm.ConnectionSettings) {
	if protoData.DNSConfig != nil {
		var dns1, dns2 uint32
		if len(protoData.DNSConfig.PrimaryDNS) > 0 {
			dns1 = IPToUInt32LI(protoData.DNSConfig.PrimaryDNS)
			connection[IPV4Key][DNSKey] = []uint32{dns1}

			if len(protoData.DNSConfig.SecondaryDNS) > 0 {
				dns2 = IPToUInt32LI(protoData.DNSConfig.SecondaryDNS)
				connection[IPV4Key][DNSKey] = []uint32{dns1, dns2}
			}
		}

		if protoData.DHCP == Enabled {
			connection[IPV4Key][DNSIgnoreAutoKey] = Yes
		}
	}
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
	connection[ConnectionKey][IDKey] = fmt.Sprintf("%s_%s", identifier, connectionSuffix)
	connection[ConnectionKey][UUIDKey] = uuid.New().String()
	connection[ConnectionKey][TimeStampKey] = time.Now().Unix()
	connection[ConnectionKey][TypeKey] = EthernetType
	connection[ConnectionKey][InterfaceNameKey] = deviceName

	putMACAddress(protoData, connection)
}

// putMACAddress puts the MAC address and sets the MACAddressKey.
func putMACAddress(protoData *v1.Interface, connection nm.ConnectionSettings) {
	uintMac, err := net.ParseMAC(protoData.MacAddress)
	if err != nil {
		log.Printf("Error parsing MAC address: %v", err)
	}
	connection[EthernetType][MACAddressKey] = uintMac
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
		err = ioutil.WriteFile(fileName, buffer, 0666)
	}

	return err
}

func readMapFromFile(fileName string) (map[string]string, error) {

	var parsedMap map[string]string
	buffer, err := ioutil.ReadFile(fileName)

	if err == nil {
		err = json.Unmarshal(buffer, &parsedMap)
	}

	return parsedMap, err
}

func getInterfaceForLabel(label string) string {
	var interfaceName string
	labelMap, err := readMapFromFile(LabelMapFileName)

	if err == nil {
		interfaceName = labelMap[strings.ToUpper(label)]
	} else {
		log.Println(err)
	}

	return interfaceName
}

func getLabelForInterface(interfaceName string) string {
	labelMap, err := readMapFromFile(LabelMapFileName)

	if err == nil {
		for label, value := range labelMap {
			if value == strings.ToUpper(interfaceName) {
				return label
			}
		}

		log.Printf("There is no label for interface %s.", interfaceName)
	} else {
		log.Println(err)
	}

	return ""
}
