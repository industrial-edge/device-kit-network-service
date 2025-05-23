/*
 * Copyright © Siemens 2020 - 2025. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package main

import (
	networkservice "networkservice/app"
	"os"
)

func main() {

	networkserviceapp := networkservice.CreateServiceApp()
	networkserviceapp.StartApp()
	networkserviceapp.StartGRPC(os.Args)
}
