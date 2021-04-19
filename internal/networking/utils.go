/*
 * Copyright (c) 2020 Siemens AG
 * mailto: dsprojindustrialedgeteamiceedge.tr@internal.siemens.com
 */

package networking

import (
	"container/list"
	"fmt"
	nm "github.com/Wifx/gonetworkmanager"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
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
//INSTANCE based on ipv4 and name from backup
func retrieveSettingsFromBackup(backup nm.ConnectionSettings) nm.ConnectionSettings {
	connection := make(nm.ConnectionSettings)
	connection[ConnectionKey] = make(dict)
	connection[IPV4Key] = make(dict)
	connection[ConnectionKey][IDKey] = backup[ConnectionKey][IDKey]
	connection[ConnectionKey][TypeKey] = backup[ConnectionKey][TypeKey]
	connection[ConnectionKey][InterfaceNameKey] = backup[ConnectionKey][InterfaceNameKey]
	connection[ConnectionKey][UUIDKey] = uuid.New().String()
	connection[ConnectionKey][TimeStampKey] = time.Now().UnixNano()
	connection[EthernetType] = backup[EthernetType]
	connection[IPV4Key] = backup[IPV4Key]
	return connection
}

func parseStaticIPConfig(connection nm.ConnectionSettings) *v1.Interface_StaticConf {
	dict := connection[IPV4Key][AddressDataKey].([]map[string]dbus.Variant)

	config := &v1.Interface_StaticConf{}
	if len(dict) > 0 {
		if dict[0][AddressKey].Value() != nil {
			config.IPv4 = dict[0][AddressKey].Value().(string)
		}

		if dict[0][PrefixKey].Value() != nil {
			config.NetMask = ParseNetMask(dict[0][PrefixKey].Value().(uint32))
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
		ipv4Address , _ := ipv4conf.GetPropertyAddressData()
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
	}

	// get device interface name
	interfaceName , _:= device.GetPropertyInterface()
	log.Println("interfacename :", interfaceName)

	// get layer2 config from device
	l2device := dockerNetworkGetMacvlanConnection(interfaceName)
	retVal.L2Conf = l2device

	retVal.InterfaceName = interfaceName

	return retVal
}

// Converts DBus data (nm.ConnectionSettings) to Device Model Proto
func convertToProto(connection nm.ConnectionSettings, ipv4Config  nm.IP4Config, mac string) *v1.Interface {

	retVal := &v1.Interface{}
	retVal.MacAddress = strings.ToUpper(mac)

	if connection[IPV4Key][MethodKey] == Auto {
		retVal.DHCP = Enabled
		retVal.Static = parseDHCPIPv4Config(ipv4Config)
	} else {
		retVal.Static = parseStaticIPConfig(connection)
		retVal.DHCP = Disabled
	}

	if val, ok :=  connection[IPV4Key]; ok {
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

//Creates new NetworkManager->ConnectionSettings from given device model proto data
func newSettingsFromProto(protoData *v1.Interface, deviceName string) nm.ConnectionSettings {
	connection := make(nm.ConnectionSettings)
	connection[ConnectionKey] = make(dict)
	connection[IPV4Key] = make(dict)
	connection[EthernetType] = make(dict)
	var connectionSuffix string

	if protoData.GatewayInterface {
		connection[IPV4Key][RouteMetricKey] = 1
	}
	if protoData.DHCP == Enabled {
		connection[IPV4Key][MethodKey] = Auto
		connectionSuffix = DHCP
	} else {
		connection[IPV4Key][MethodKey] = Manual
		connection[IPV4Key][GatewayKey] = protoData.Static.Gateway

		//DBus variant type is needed to set manual IP via DBus
		address := dbus.MakeVariantWithSignature(protoData.Static.IPv4, dbus.ParseSignatureMust("s"))
		prefix := dbus.MakeVariantWithSignature(ParseNetMaskSize(protoData.Static.NetMask), dbus.ParseSignatureMust("u"))

		ipDict := make(DBusDict)
		ipDict[AddressKey] = address //ip      e.g: "192.168.0.1"
		ipDict[PrefixKey] = prefix   //subnet  e.g:  24

		connection[IPV4Key][AddressDataKey] = []DBusDict{ipDict}
		connectionSuffix = "static"

	}
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
	}

	connection[ConnectionKey][IDKey] = fmt.Sprintf("%s_%s_%s", deviceName, protoData.MacAddress, connectionSuffix)
	connection[ConnectionKey][UUIDKey] = uuid.New().String()
	connection[ConnectionKey][TimeStampKey] = time.Now().Unix()

	connection[ConnectionKey][TypeKey] = EthernetType
	connection[ConnectionKey][InterfaceNameKey] = deviceName
	uintMac, _ := net.ParseMAC(protoData.MacAddress)
	connection[EthernetType][MACAddressKey] = uintMac

	return connection
}
