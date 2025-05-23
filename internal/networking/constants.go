/*
 * Copyright © Siemens 2020 - 2025. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

// DBUS NetworkManager Key Names
// Please be considerate of the DBUS key values when you are changing the fields

const (
	// MACAddressKey
	MACAddressKey = "mac-address"
	// DNSKey
	DNSKey = "dns"
	// DNSSearchKey
	DNSSearchKey = "dns-search"
	// DNSIgnoreAuto
	DNSIgnoreAutoKey = "ignore-auto-dns"
	// Yes
	Yes = 1
	// Enabled
	Enabled = "enabled"
	// Disabled
	Disabled = "disabled"
	// Auto
	Auto = "auto"
	// Manual
	Manual = "manual"
	// DHCP
	DHCP = "dhcp"
	// Static
	Static = "static"
	// MethodKey
	MethodKey = "method"
	// RouteMetricKey
	RouteMetricKey = "route-metric"
	// EthernetType
	EthernetType = "802-3-ethernet"
	// ConnectionKey
	ConnectionKey = "connection"
	// IPV4Key
	IPV4Key = "ipv4"
	// IDKey
	IDKey = "id"
	// TypeKey
	TypeKey = "type"
	// InterfaceNameKey
	InterfaceNameKey = "interface-name"
	// UUIDKey
	UUIDKey = "uuid"
	// TimeStampKey
	TimeStampKey = "timestamp"
	// AddressKey
	AddressKey = "address"
	// PrefixKey
	PrefixKey = "prefix"
	// GatewayKey
	GatewayKey = "gateway"
	// IPAddressKey
	IPAddressKey = "ip_address"
	// SubnetMaskKey
	SubnetMaskKey = "subnet_mask"
	// DHCPServerIdentifierKey
	DHCPServerIdentifierKey = "dhcp_server_identifier"
	// AddressDataKey
	AddressDataKey = "address-data"
	// LabelMapFileName
	LabelMapFileName = "/var/network.label"
	// Highest Possible Metric Value
	MaxMetricValue = 255
	// Route Destination Value For Outgoing Traffic
	OutgoingRouteDestination = "0.0.0.0"
	// Prefix For Outgoing Traffic
	OutgoingRoutePrefix = 0
)
