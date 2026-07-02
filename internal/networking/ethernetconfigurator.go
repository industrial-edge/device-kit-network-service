/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"log"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

type EthernetConfigurator struct {
	nc     *NetworkConfigurator
	iface  *v1.Interface
	device nm.DeviceWired
}

func NewEthernetConfigurator(nc *NetworkConfigurator, iface *v1.Interface) (*EthernetConfigurator, error) {
	log.Println("Initializing ethernet configurator to update interface settings")

	device, err := nc.getDeviceBy(iface)
	if err != nil {
		log.Println("Error getting device for new ethernet configurator: ", err)
		return nil, err
	}

	log.Println("Successfully instantiated ethernet configurator")
	return &EthernetConfigurator{
		nc:     nc,
		iface:  iface,
		device: device,
	}, nil
}

func (c *EthernetConfigurator) Configure() error {
	log.Println("Configuring ethernet connection...")

	settings, err := c.nc.prepareSettings(c.iface, c.device)
	if err != nil {
		return err
	}

	if err := c.nc.updateConnections(c.device, settings); err != nil {
		return err
	}

	log.Println("Ethernet connection configured successfully")
	return nil
}
