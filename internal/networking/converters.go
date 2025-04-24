/*
 * Copyright (c) 2021 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

//UTILS for NETWORKING
//low level address parsing and conversion utilities
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// IPFromUInt32BI converts ip in -uint32- BigIndian representation to -string- representation.
//(e.g 12312312312 to "8.8.8.8")
func IPFromUInt32BI(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

// IPFromUInt32LI converts ip in -uint32- LittleIndian representation to -string- representation.
//(e.g 12312312312 to "8.8.8.8")
func IPFromUInt32LI(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

// IPToUInt32BI converts string to Big Indian uint32
func IPToUInt32BI(ip string) uint32 {
	var value uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &value)
	return value
}

// IPToUInt32LI converts string to Little Indian uint32
func IPToUInt32LI(ip string) uint32 {
	var value uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.LittleEndian, &value)
	return value
}

func ParseNetMaskSize(value string) uint32 {
	val := net.ParseIP(value)
	addr := val.To4()

	sz, _ := net.IPv4Mask(addr[0], addr[1], addr[2], addr[3]).Size()
	return uint32(sz)

}
func ParseNetMask(simpleNotation uint32) string {
	var simple string
	_, ipnet, err := net.ParseCIDR("0.0.0.0/" + fmt.Sprint(simpleNotation))
	if err == nil && ipnet != nil {
		simple = fmt.Sprintf("%v.%v.%v.%v", ipnet.Mask[0], ipnet.Mask[1], ipnet.Mask[2], ipnet.Mask[3])
	} else {
		simple = ""
	}

	return simple
}
