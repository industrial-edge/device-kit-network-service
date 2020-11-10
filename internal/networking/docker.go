/*
 * Copyright (c) 2020 Siemens AG
 * mailto: dsprojindustrialedgeteamiceedge.tr@internal.siemens.com
 */

package networking

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// Docker macvlan Layer 2 config parameters
type L2Config struct {
	interfaceName string
	netmask  string
	startIp string
	ipRange int
}


type DockerNetworkLS struct {
	Name string `json:"Name"`
	Id string `json:"Id"`
	Scope string `json:"Scope"`
	Driver string `json:"Driver"`
	EnableIPv6 bool `json:"EnableIPv6"`
	IPAM IPAM `json:"IPAM"`
	Internal bool `json:"Internal"`
	Containers map[string]Container `json:"Containers"`
	Options map[string]string `json:"Options"`
	Labels interface{} `json:"Labels"`
}

type IPAM struct {
	Driver string `json:"Driver"`
	Options interface{} `json:"Options"`
	Config []Conf `json:"Config"`
}

type Conf struct {
	Subnet string `json:"Subnet"`
	IPRange string `json:"IPRange"`
	Gateway string `json:"Gateway"`
}

type Container struct {
	Name string `json:"Name"`
	EndPointID string `json:"EndpointID"`
	MacAddress string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
}

func dockerNetworkGetMacvlanConnection() L2Config{
	var l2Conf L2Config
	// Run docker network ls
	// docker network ls --format "{{.Name}}" --filter driver=macvlan
	cmdLs := "docker network ls --format \"{{.Name}}\" --filter driver=macvlan"
	macVlanName, err := exec.Command("/bin/bash", "-c", cmdLs).Output()

	if err != nil || len(macVlanName) <= 0 {
		log.Println("docker network ls : ", err)
		return l2Conf
	}

	cmdInspect := "docker network inspect " + string(macVlanName)
	inspectOutJson, err2 := exec.Command("/bin/bash", "-c", cmdInspect).Output()

	if err2 != nil || len(inspectOutJson) <= 0 {
		log.Println("docker network inspect  : ", err2)
		return l2Conf
	}

	var structuredData []DockerNetworkLS
	err = json.Unmarshal(inspectOutJson, &structuredData)
	if err != nil {
		fmt.Println(err)
		return l2Conf
	}
	subnetPrefixString :=  strings.Split(structuredData[0].IPAM.Config[0].Subnet, "/")
	subnetPrefixUint, _ := strconv.ParseUint(subnetPrefixString[1], 10, 32)

	startIP := strings.Split(structuredData[0].IPAM.Config[0].IPRange, "/")
	startIPPrefix, _ := strconv.ParseUint(startIP[1], 10, 32)
    // prefix should be between 0-32
	if startIPPrefix > 32  || startIPPrefix < 0{
		return l2Conf
	}
	exponent := 32 - startIPPrefix
	ipRange := int(math.Pow(2, float64(exponent)))

	l2Conf.interfaceName = structuredData[0].Options["parent"]
	l2Conf.netmask       = ParseNetMask(uint32(subnetPrefixUint))
	l2Conf.startIp       = startIP[0]
	l2Conf.ipRange       = ipRange

	return l2Conf
}
