/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package common

import v1 "networkservice/api/siemens_iedge_dmapi_v1"

type RouteMetricPriority int32

const (
	RouteMetricLowestPriority  RouteMetricPriority = -1
	RouteMetricHighestPriority RouteMetricPriority = 1
)

func (r RouteMetricPriority) ToInt32() int32 {
	return int32(r)
}

type InterfaceType int

const (
	InterfaceTypeUnknown InterfaceType = iota
	InterfaceTypeEthernet
	InterfaceTypeGSM
)

// Translator
func InterfaceTypeFromProto(protoType v1.Interface_InterfaceTypeEnum) InterfaceType {
	switch protoType {
	case v1.Interface_ETHERNET:
		return InterfaceTypeEthernet
	case v1.Interface_GSM:
		return InterfaceTypeGSM
	default:
		return InterfaceTypeUnknown
	}
}
