/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package interfaces

import (
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

type ConnectionHelper interface {
	FindConnectionWithStatus(nm.Device) (nm.Connection, bool, error)
	GetConnectionSettings(nm.Connection) (nm.ConnectionSettings, error)
	UpdateConnection(nm.Connection, nm.ConnectionSettings) error
	AddConnection(nm.ConnectionSettings) (nm.Connection, error)
	DeleteConnection(nm.Connection) error
	ActivateConnection(nm.Connection, nm.Device) error
	ActivateConnectionWithRetry(nm.Connection, nm.Device) error
	DeactivateConnection(nm.Device) error
}

type ConnectionUtils interface {
	ConfigureGateway(nm.ConnectionSettings, common.RouteMetricPriority)
	ConfigureDNS(nm.ConnectionSettings, *v1.Interface_Dns)
	RetrieveDeviceName(nm.ConnectionSettings) string
}

type DeviceHelper interface {
	GetAvailableDevicesMap() (map[string]nm.Device, error)
	GetGSMDevice(name string) (nm.Device, error)
}

type ConnectionHandler interface {
	CreateConnection(settings *v1.ConnectionSettings) error
	RemoveConnection(settings *v1.ConnectionSettings) error
}
