/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package factory

import (
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/gsm"
	"networkservice/internal/networking/interfaces"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

func ConnectionFactory(connectionType v1.ConnectionSettings_ConnectionType, nm nm.NetworkManager) interfaces.ConnectionHandler {

	switch connectionType {
	case v1.ConnectionSettings_GSM:
		connHandler, err := gsm.NewGSMHandler(nm)
		if err != nil {
			return nil
		}
		return connHandler
	default:
		return nil
	}
}
