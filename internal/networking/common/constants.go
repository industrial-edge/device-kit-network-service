/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package common

import (
	"os"
	"time"
)

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
	// RouteDataKey
	RouteDataKey = "route-data"
	// EthernetType
	EthernetType = "802-3-ethernet"
	// ConnectionKey
	ConnectionKey = "connection"
	// IPV4Key
	IPV4Key = "ipv4"
	// IPV6Key
	IPV6Key = "ipv6"
	// addresses
	AddressesKey = "addresses"
	//routes
	RoutesKey = "routes"
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
	// Destination key
	DestinationKey = "dest"
	// Next-hop key
	NextHopKey = "next-hop"
	// Metric key
	MetricKey = "metric"
	// GSMSetting key
	GSMSetting = "gsm"
	// AutoConnectKey key
	AutoConnectKey = "autoconnect"
	// APNKey key
	APNKey = "apn"
	// UserName key
	UsernameKey = "username"
	// Password key
	PasswordKey = "password"
	// PINKey key
	PINKey = "pin"
	// GSMConfigPath is the path where GSM config info is stored
	GSMConfigPath = "/var/lib/gsm/config/info"

	// GSMConfigParentPath is the parent directory where GSM config info is stored
	GSMConfigParentPath = "/var/lib/gsm/config"

	// GSM config file JSON keys (capitalized, distinct from NM D-Bus keys)
	GSMConfigKeyAPN      = "Apn"
	GSMConfigKeyPIN      = "Pin"
	GSMConfigKeyUsername = "Username"
	GSMConfigKeyPassword = "Password"
	// MinRouteMetric is the minimum value for a route metric for extra routes
	MinRouteMetric        = 2
	ValidRegexForAPN      = "^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*(?:\\.[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*)*$"
	ValidRegexForPIN      = "^\\d{4,8}$"
	ValidRegexForUserName = "^(?:[A-Za-z0-9._-]{1,64})?$"
	// It enforces: Length: 1–64 characters Allowed characters: Uppercase letters A–Z,
	//Lowercase letters a–z, Digits 0–9 nand Many common special characters
	ValidRegexForPassword = "^[A-Za-z0-9!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>\\/?]{1,64}$"

	// Using a Fibonacci backoff strategy with min 1 second and max 7 attempts,
	// since the cumulative delay is ~30 seconds
	// which is reasonable for GSM connection activations.
	DefaultMaxRetries   = 7
	DefaultInitialDelay = 1 * time.Second

	LogNoConnectionKey = "Connection settings do not have connection key"

	ErrLogSuffix = " error: "

	GSMConfigDirPermission  = os.FileMode(0755)
	GSMConfigFilePermission = os.FileMode(0600)
)

// Connection state constants
const (
	ConnStateSIMLock      = "SIM_LOCKED"
	ConnStateNoModem      = "NO_MODEM"
	ConnStateConnected    = "CONNECTED"
	ConnStateConnecting   = "CONNECTING"
	ConnStateDisconnected = "DISCONNECTED"
	ConnStateUnspecified  = "UNSPECIFIED"
)

// Access technology constants
const (
	AccessTech4G          = "4G"
	AccessTech3G          = "3G"
	AccessTech2G          = "2G"
	AccessTechUnspecified = "UNSPECIFIED"
)

// Registration state constants
const (
	RegStateRegistered  = "REGISTERED"
	RegStateSearching   = "SEARCHING"
	RegStateDenied      = "DENIED"
	RegStateIdle        = "IDLE"
	RegStateUnspecified = "UNSPECIFIED"
)

// Roaming state constants
const (
	RoamingStateHome        = "HOME"
	RoamingStateRoaming     = "ROAMING"
	RoamingStateUnspecified = "UNSPECIFIED"
)

// failed reason constants
const (
	FailedReasonNone                = "NO_ERROR"
	FailedReasonUnknown             = "UNKNOWN_ERROR"
	FailedReasonSimMissing          = "SIM_MISSING"
	FailedReasonSimError            = "SIM_ERROR"
	FailedReasonUnknownCapabilities = "UNKNOWN_CAPABILITIES"
	FailedReasonESimNoProfiles      = "ESIM_NO_PROFILES"
)
