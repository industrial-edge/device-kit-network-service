/*
 * Copyright (c) 2020 Siemens AG
 * mailto: dsprojindustrialedgeteamiceedge.tr@internal.siemens.com
 */

package networking

import (
	"encoding/json"
	"log"
	"math"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
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
	gateway string
	auxiliaryAddresses map[string]string
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
	AuxiliaryAddresses map[string]string `json:"AuxiliaryAddresses"`
}

type Container struct {
	Name string `json:"Name"`
	EndPointID string `json:"EndpointID"`
	MacAddress string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
}

var execCommand = exec.Command

func dockerNetworkGetMacvlanConnection(interfaceName string) *v1.Interface_L2{
	retVal := &v1.Interface_L2{}
	// Run docker network ls
	// docker network ls --format "{{.Name}}" --filter driver=macvlan
	cmdLs := "docker network ls --format \"{{.Name}}\" --filter driver=macvlan"
	macVlanName, err := execCommand("/bin/bash", "-c", cmdLs).Output()
	if err != nil || len(macVlanName) <= 0 {
		log.Println("docker network ls : ", err)
		return retVal
	}

	cmdInspect := "docker network inspect " + string(macVlanName)
	inspectOutJSON, err2 := execCommand("/bin/bash", "-c", cmdInspect).Output()
	if err2 != nil || len(inspectOutJSON) <= 0 {
		log.Println("docker network inspect  : ", err2)
		return retVal
	}

	var structuredData []DockerNetworkLS
	err = json.Unmarshal(inspectOutJSON, &structuredData)
	if err != nil {
		log.Println(err)
		return retVal
	}
	subnetPrefixString :=  strings.Split(structuredData[0].IPAM.Config[0].Subnet, "/")
	subnetPrefixUint, _ := strconv.ParseUint(subnetPrefixString[1], 10, 32)

	startIP := strings.Split(structuredData[0].IPAM.Config[0].IPRange, "/")
	startIPPrefix, _ := strconv.ParseUint(startIP[1], 10, 32)
    // prefix should be between 0-32
	if startIPPrefix > 32  || startIPPrefix < 0{
		return retVal
	}
	exponent := 32 - startIPPrefix
	ipRange := int(math.Pow(2, float64(exponent)))

	// check interface has layer2 config
    if structuredData[0].Options["parent"] == interfaceName{
		retVal.NetMask = ParseNetMask(uint32(subnetPrefixUint))
		retVal.StartingAddressIPv4 =  startIP[0]
		retVal.Range = strconv.Itoa(ipRange)
		retVal.Gateway = structuredData[0].IPAM.Config[0].Gateway
		retVal.AuxiliaryAddresses = make(map[string]string)
		// Copy from the  l2device.auxiliaryAddresses map to the l2proto.AuxiliaryAddresses map
		for key, value := range structuredData[0].IPAM.Config[0].AuxiliaryAddresses {
			retVal.AuxiliaryAddresses[key] = value
		}
    }

	return retVal
}
