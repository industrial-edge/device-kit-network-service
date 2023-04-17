/*
 * Copyright (c) 2021 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package system

import (
	"log"
	"os/exec"
)



// Configurable Inteface handles configurations on the host services for onboarding.
type Configurable interface {
	DaemonReload(query string) bool
}

//Helper struct for host configurations
type Helper struct {
}


func (f *Helper) DaemonReload(query string) bool {
	errDaemonReload := exec.Command("bash", "-c", query).Run()
	if errDaemonReload == nil {
		log.Println("daemon-reload successfully")
		return true
	}
	log.Println("Failed to daemon-reload")

	return false
}
