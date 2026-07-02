/*
 * Copyright © Siemens 2020 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"errors"
	"fmt"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"strings"
)

type verifyResult struct {
	builder strings.Builder
	retVal  bool
}

// Checks all preconditions before applying any settings.
func verify(newSettings *v1.NetworkSettings, configurator *NetworkConfigurator) (bool, error) {
	resultOut := &verifyResult{
		builder: strings.Builder{},
		retVal:  true,
	}

	for _, element := range newSettings.Interfaces {

		// GSM interface should be verified first, 
		// Check the APN and if specified, the DNS server addresses, then it checks that the nameserver IP(s) are valid.
		// if the GSM interface we are validating then we should not continue with other validations like MAC address, 
		// static IP configuration and routes because they are not relevant for GSM interface 
		if element.GetGsmConfiguration() != nil {
			verifyGSMConfigNotEmpty(element, resultOut)
			if element.DNSConfig != nil {
				verifyDNS(element, resultOut)
			}
			continue
		}

		if element.Label == "" {
			verifyMAC(element, resultOut, configurator)
		}

		if element.Static != nil {
			verifyStaticConf(element, resultOut)
		}
		if element.DNSConfig != nil {
			verifyDNS(element, resultOut)
		}

		verifyRoutes(element, resultOut)
	}

	verifySingleDefaultGateway(newSettings, resultOut)

	errorMessages := resultOut.builder.String()
	var err error
	if len(errorMessages) > 0 {
		err = errors.New(errorMessages)
		log.Println("Verification result:")
		log.Println(errorMessages)
	}
	return resultOut.retVal, err
}

func verifyGSMConfigNotEmpty(element *v1.Interface, result *verifyResult) {
	if len(element.GetGsmConfiguration().GetApn()) == 0 {
		result.retVal = false
		result.builder.WriteString("Minimum GSM configuration is empty - APN is missing\n")
		return
	}
}

func verifyMAC(element *v1.Interface, result *verifyResult, configurator *NetworkConfigurator) {
	_, err := net.ParseMAC(element.MacAddress)
	if err != nil {
		result.retVal = false
		result.builder.WriteString(fmt.Sprintf("wrong mac address %s \n", element.MacAddress))
	} else {
		val := configurator.getDeviceWithMac(element.MacAddress)
		if val == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("no network interface found for the specified mac address: mac address %s \n", element.MacAddress))
		}
	}
}
func verifyStaticConf(element *v1.Interface, result *verifyResult) {
	if len(element.Static.IPv4) > 0 {
		val := net.ParseIP(element.Static.IPv4)
		if val == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong ip address %s \n", element.Static.IPv4))
		}
	}
	if len(element.Static.Gateway) > 0 {
		val := net.ParseIP(element.Static.Gateway)
		if val == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong gateway address %s \n", element.Static.Gateway))
		}
	}
	if len(element.Static.NetMask) > 0 {
		val := net.ParseIP(element.Static.NetMask)
		if val == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong netmask address %s \n", element.Static.NetMask))
		}
	}
}

func verifyDNS(element *v1.Interface, result *verifyResult) {
	if len(element.DNSConfig.PrimaryDNS) > 0 {
		val := net.ParseIP(element.DNSConfig.PrimaryDNS)
		if val == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong dns address %s \n", element.DNSConfig.PrimaryDNS))
		}
	}
	if len(element.DNSConfig.SecondaryDNS) > 0 {
		val := net.ParseIP(element.DNSConfig.SecondaryDNS)
		if val == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong dns address %s \n", element.DNSConfig.SecondaryDNS))
		}
	}
}

func verifyRoutes(element *v1.Interface, result *verifyResult) {
	for _, route := range element.Routes {
		if val := route.GetDestination(); net.ParseIP(val) == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong route destination address %s \n", val))
		}

		if val := route.GetNetmask(); net.ParseIP(val) == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong route netmask address %s \n", val))
		}

		if val := route.GetNextHop(); val != "" && net.ParseIP(val) == nil {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong route next-hop address %s \n", val))
		}

		if val := route.GetMetric(); val < common.MinRouteMetric {
			result.retVal = false
			result.builder.WriteString(fmt.Sprintf("wrong route metric %d <= 1 \n", val))
		}
	}
}

func verifySingleDefaultGateway(newSettings *v1.NetworkSettings, result *verifyResult) {
	count := 0
	for _, element := range newSettings.Interfaces {
		if element.GatewayInterface {
			count++
		}
	}

	// Only one default gateway interface is allowed
	// especially important when GSM is used as default gateway
	// the other ethernet interfaces must not be set as default gateway
	if count > 1 {
		result.retVal = false
		result.builder.WriteString("more than one default gateway interface is not allowed \n")
	}
}
