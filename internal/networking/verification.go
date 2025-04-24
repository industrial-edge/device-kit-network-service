/*
 * Copyright © Siemens 2020 - 2025. ALL RIGHTS RESERVED.
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

		if element.Label == "" {
			verifyMAC(element, resultOut, configurator)
		}


		if element.Static != nil {
			verifyStaticConf(element, resultOut)
		}
		if element.DNSConfig != nil {
			verifyDNS(element, resultOut)
		}
	}
	errorMessages := resultOut.builder.String()
	var err error
	if len(errorMessages) > 0 {
		err = errors.New(errorMessages)
		log.Println("Verification result:")
		log.Println(errorMessages)
	}
	return resultOut.retVal, err
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
			result.builder.WriteString(fmt.Sprintf("device does not exist: mac address %s \n", element.MacAddress))
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
