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

	nm "github.com/Wifx/gonetworkmanager/v2"
	modemManager "github.com/maltegrosse/go-modemmanager"
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

// ValidateGSMConfigFileContent validates a parsed GSM config file map.
// It checks the mandatory APN field and, if optional fields (Pin, Username, Password)
// are present, validates their format against the same rules used for incoming requests.
func ValidateGSMConfigFileContent(parsed map[string]string) error {
	apn := parsed[common.GSMConfigKeyAPN]
	if len(apn) == 0 {
		return fmt.Errorf("config file missing required APN field")
	}
	if !validateAPN(apn) {
		return fmt.Errorf("config file has invalid APN format: %s", apn)
	}

	if pin, ok := parsed[common.GSMConfigKeyPIN]; ok && len(pin) > 0 {
		if !validatePIN(pin) {
			return fmt.Errorf("config file has invalid PIN format")
		}
	}

	return validateFileCredentials(parsed[common.GSMConfigKeyUsername], parsed[common.GSMConfigKeyPassword])
}

// validateFileCredentials validates the username/password pair from a GSM config file.
// Both must be present together (if either is provided), and each must match the expected format.
func validateFileCredentials(username, password string) error {
	if len(username) > 0 && len(password) == 0 {
		return fmt.Errorf("config file has Username but missing Password")
	}
	if len(password) > 0 && len(username) == 0 {
		return fmt.Errorf("config file has Password but missing Username")
	}
	if len(username) > 0 && !validateUserName(username) {
		return fmt.Errorf("config file has invalid Username format")
	}
	if len(password) > 0 && !validatePassword(password) {
		return fmt.Errorf("config file has invalid Password format")
	}

	return nil
}

// extractStringFromSettings safely extracts a string value from nested ConnectionSettings.
func extractStringFromSettings(settings nm.ConnectionSettings, section, key string) string {
	sectionMap, ok := settings[section]
	if !ok || sectionMap == nil {
		return ""
	}
	if val, ok := sectionMap[key].(string); ok {
		return val
	}
	return ""
}

func mapStateFailedReasonToString(reason modemManager.MMModemStateFailedReason) string {
	switch reason {
	case modemManager.MmModemStateFailedReasonNone:
		return common.FailedReasonNone
	case modemManager.MmModemStateFailedReasonUnknown:
		return common.FailedReasonUnknown
	case modemManager.MmModemStateFailedReasonSimMissing:
		return common.FailedReasonSimMissing
	case modemManager.MmModemStateFailedReasonSimError:
		return common.FailedReasonSimError
	case modemManager.MmModemStateFailedReasonUnknownCapabilities:
		return common.FailedReasonUnknownCapabilities
	case modemManager.MmModemStateFailedReasonEsimWithoutProfiles:
		return common.FailedReasonESimNoProfiles
	default:
		return common.FailedReasonUnknown
	}
}

// mapProtoConnectionState maps ModemManager modem state to connection state proto enum.
// Note: MmModemStateRegistered means the modem is registered to a network but has no active data bearer.
// MmModemStateEnabled means the modem is powered on but not yet registered.
// In both cases, this returns ConnStateDisconnected because there is no data connection (bearer).
// Only MmModemStateConnected with an active bearer returns ConnStateConnected.
func mapProtoConnectionState(state modemManager.MMModemState, bearerConn bool) string {

	switch state {
	case modemManager.MmModemStateLocked:
		return common.ConnStateSIMLock
	case modemManager.MmModemStateConnected:
		if bearerConn {
			return common.ConnStateConnected
		}
		// Connected modem state may briefly report before bearer state settles.
		return common.ConnStateConnecting
	case modemManager.MmModemStateConnecting, modemManager.MmModemStateInitializing, modemManager.MmModemStateEnabling:
		return common.ConnStateConnecting
	case modemManager.MmModemStateDisconnecting, modemManager.MmModemStateDisabled, modemManager.MmModemStateDisabling,
		modemManager.MmModemStateEnabled, modemManager.MmModemStateRegistered, modemManager.MmModemStateSearching,
		modemManager.MmModemStateFailed:
		return common.ConnStateDisconnected
	default:
		return common.ConnStateUnspecified
	}
}
func mapProtoAccessTechnology(techs []modemManager.MMModemAccessTechnology) string {
	result := common.AccessTechUnspecified

	for _, t := range techs {
		switch t {
		case modemManager.MmModemAccessTechnologyLte:
			return common.AccessTech4G
		case modemManager.MmModemAccessTechnologyUmts, modemManager.MmModemAccessTechnologyHsdpa,
			modemManager.MmModemAccessTechnologyHsupa, modemManager.MmModemAccessTechnologyHspa,
			modemManager.MmModemAccessTechnologyHspaPlus:
			if result == common.AccessTechUnspecified || result == common.AccessTech2G {
				result = common.AccessTech3G
			}
		case modemManager.MmModemAccessTechnologyGsm, modemManager.MmModemAccessTechnologyGsmCompact,
			modemManager.MmModemAccessTechnologyGprs, modemManager.MmModemAccessTechnologyEdge:
			if result == common.AccessTechUnspecified {
				result = common.AccessTech2G
			}
		default:
			return result
		}
	}

	return result
}

func mapProtoRegistration(reg modemManager.MMModem3gppRegistrationState) (string, string) {
	switch reg {
	case modemManager.MmModem3gppRegistrationStateHome,
		modemManager.MmModem3gppRegistrationStateHomeSmsOnly,
		modemManager.MmModem3gppRegistrationStateHomeCsfbNotPreferred:
		return common.RegStateRegistered, common.RoamingStateHome
	case modemManager.MmModem3gppRegistrationStateRoaming,
		modemManager.MmModem3gppRegistrationStateRoamingSmsOnly,
		modemManager.MmModem3gppRegistrationStateRoamingCsfbNotPreferred:
		return common.RegStateRegistered, common.RoamingStateRoaming
	case modemManager.MmModem3gppRegistrationStateSearching:
		return common.RegStateSearching, common.RoamingStateUnspecified
	case modemManager.MmModem3gppRegistrationStateDenied,
		modemManager.MmModem3gppRegistrationStateEmergencyOnly:
		return common.RegStateDenied, common.RoamingStateUnspecified
	case modemManager.MmModem3gppRegistrationStateIdle:
		return common.RegStateIdle, common.RoamingStateUnspecified
	default:
		return common.RegStateUnspecified, common.RoamingStateUnspecified
	}
}
