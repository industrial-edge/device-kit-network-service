/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package connectionmanager

import (
	"networkservice/internal/networking/common"
	"testing"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionUtils(t *testing.T) {
	// Arrange & Act
	got := NewConnectionUtils()

	// Assert
	assert.NotNil(t, got, "expected NewConnectionUtils to return a non-nil value")
}

func TestConnUtils_RetrieveDeviceName_RainyDayScenarios(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		settings nm.ConnectionSettings
		want     string
	}{
		{
			name:     "NilConnectionSettings",
			settings: nm.ConnectionSettings{},
			want:     "",
		},
		{
			name: "EmptyConnectionMap",
			settings: nm.ConnectionSettings{
				"connection": map[string]any{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConnUtils{}
			got := c.RetrieveDeviceName(tt.settings)

			assert.Equal(t, tt.want, got, "unexpected device name retrieved")
		})
	}
}

func TestConnUtils_RetrieveDeviceName(t *testing.T) {
	c := ConnUtils{}
	settings := nm.ConnectionSettings{
		"connection": map[string]any{
			"interface-name": "eth0",
		},
	}
	want := "eth0"
	got := c.RetrieveDeviceName(settings)

	assert.Equal(t, want, got, "unexpected device name retrieved")
}

func Test_ConnUtils_ConfigureGateway(t *testing.T) {
	c := ConnUtils{}
	connSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{},
	}
	c.ConfigureGateway(connSettings, common.RouteMetricLowestPriority)

	ipv4Settings := connSettings["ipv4"]
	assert.NotNil(t, ipv4Settings, "Expected ipv4 settings to be present")
	assert.Equal(t, common.RouteMetricLowestPriority, ipv4Settings[common.RouteMetricKey], "Expected route metric to be set to lowest priority")
}

func Test_ConnUtils_CleanIPv6Settings(t *testing.T) {
	c := ConnUtils{}
	connSettings := nm.ConnectionSettings{
		"ipv6": map[string]any{
			"addresses": []string{"2001:db8::1"},
			"routes":    []string{"2001:db8::/64"},
		},
	}
	c.cleanIPv6Settings(connSettings)

	ipv6Settings := connSettings["ipv6"]
	assert.NotNil(t, ipv6Settings, "Expected ipv6 settings to be present")
	assert.NotContains(t, ipv6Settings, "addresses", "Expected ipv6 addresses to be removed")
	assert.NotContains(t, ipv6Settings, "routes", "Expected ipv6 routes to be removed")
}

func Test_ConnUtils_ConfigureDNS(t *testing.T) {
	c := ConnUtils{}
	connSettings := nm.ConnectionSettings{
		"ipv4": map[string]any{},
	}
	dnsConfig := &v1.Interface_Dns{
		PrimaryDNS:   "2.2.2.2",
		SecondaryDNS: "1.1.1.1",
	}
	c.ConfigureDNS(connSettings, dnsConfig)

	ipv4Settings := connSettings["ipv4"]
	assert.NotNil(t, ipv4Settings, "Expected ipv4 settings to be present")
	dnsSetting := ipv4Settings[common.DNSKey]
	assert.NotNil(t, dnsSetting, "Expected DNS setting to be present in ipv4 settings")
	assert.IsType(t, []uint32{}, dnsSetting, "Expected DNS setting to be of type []uint32")

}
