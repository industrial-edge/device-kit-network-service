/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"fmt"
	"regexp"
	"strings"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
)

const (
	maxApnLenght     = 100
	maxApnHostLength = 63
)

// An APN is considered valid if it meets the following criteria:
//  1. Its total length does not exceed 100 characters.
//  2. If it contains dots ('.') to separate labels, each individual label
//     (part between dots) must have a length between 1 and 63 characters, inclusive.
func validateAPN(apn string) bool {

	if len(apn) > maxApnLenght {
		return false
	}
	if strings.Contains(apn, ".") {
		labels := strings.Split(apn, ".")
		for _, l := range labels {
			if len(l) < 1 || len(l) > maxApnHostLength {
				return false
			}
		}
	}
	validAPN := regexp.MustCompile(common.ValidRegexForAPN)
	return validAPN.MatchString(apn)
}

func validatePIN(pin string) bool {
	validPIN := regexp.MustCompile(common.ValidRegexForPIN)
	return validPIN.MatchString(pin)
}
func validateUserName(username string) bool {
	validUsername := regexp.MustCompile(common.ValidRegexForUserName)
	return validUsername.MatchString(username)
}

func validatePassword(password string) bool {
	validPassword := regexp.MustCompile(common.ValidRegexForPassword)
	return validPassword.MatchString(password)
}

func HasValidGSMConfig(gsmSettings *v1.ConnectionSettings_Gsm) error {
	if gsmSettings.Gsm == nil {
		return fmt.Errorf("GSM configuration is nil")
	}
	gsmSettings.Gsm.Apn = strings.TrimSpace(gsmSettings.Gsm.Apn)
	gsmSettings.Gsm.Pin = strings.TrimSpace(gsmSettings.Gsm.Pin)
	gsmSettings.Gsm.Username = strings.TrimSpace(gsmSettings.Gsm.Username)
	gsmSettings.Gsm.Password = strings.TrimSpace(gsmSettings.Gsm.Password)
	if len(gsmSettings.Gsm.Apn) == 0 {
		return fmt.Errorf("GSM configuration is empty")
	}

	if !validateAPN(gsmSettings.Gsm.Apn) {
		return fmt.Errorf("invalid APN format")
	}
	if len(gsmSettings.Gsm.Pin) > 0 {
		if !validatePIN(gsmSettings.Gsm.Pin) {
			return fmt.Errorf("invalid PIN format")
		}
	}
	if err := isValidGsmConfigUserNamePassword(gsmSettings); err != nil {
		return err
	}

	return nil
}

func isValidGsmConfigUserNamePassword(gsmSettings *v1.ConnectionSettings_Gsm) error {
	if len(gsmSettings.Gsm.Username) > 0 && len(gsmSettings.Gsm.Password) == 0 {
		return fmt.Errorf("Password is required when Username is provided")
	}
	if len(gsmSettings.Gsm.Password) > 0 && len(gsmSettings.Gsm.Username) == 0 {
		return fmt.Errorf("Username is required when Password is provided")
	}
	if len(gsmSettings.Gsm.Username) > 0 && len(gsmSettings.Gsm.Password) > 0 {
		if !validateUserName(gsmSettings.Gsm.Username) {
			return fmt.Errorf("invalid Username format")
		}
		if !validatePassword(gsmSettings.Gsm.Password) {
			return fmt.Errorf("invalid Password format")
		}
	}
	return nil
}
