/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"fmt"
	"log"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/gsm"
)

type InterfaceConfigurator interface {
	Configure() error
}

// Determines interface type and create appropriate connector, returns error if interface type is unsupported.
// GSM and Ethernet are the supported connectors.
// Additional checks for GSM is added in verification.go class.
func NewInterfaceConfigurator(nc *NetworkConfigurator, iface *v1.Interface) (InterfaceConfigurator, error) {
	log.Println("Initializing new configurator to update interface settings")

	if iface.InterfaceType == nil {
		log.Println("InterfaceType not specified by client, using fallback logic")

		if iface.GetGsmConfiguration() != nil {
			log.Println("GSM configuration found, initializing GSM configurator")
			return gsm.NewGSMConfigurator(nc.gnm, iface), nil
		} else if iface.GetMacAddress() != "" || iface.GetLabel() != "" {
			log.Println("MAC address found, initializing Ethernet configurator")
			return NewEthernetConfigurator(nc, iface)
		} else {
			return nil, fmt.Errorf("interface type is not specified and no GSM configuration, Labels or MAC found")
		}
	}

	switch iface.GetInterfaceType() {
	case v1.Interface_GSM:
		log.Println("Client specified GSM interface type")
		return gsm.NewGSMConfigurator(nc.gnm, iface), nil
	case v1.Interface_ETHERNET:
		log.Println("Client specified ETHERNET interface type")
		return NewEthernetConfigurator(nc, iface)
	default:
		return nil, fmt.Errorf("unsupported interface type: %v", iface.GetInterfaceType())
	}
}
