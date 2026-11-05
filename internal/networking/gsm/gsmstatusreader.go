/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"log"
	"time"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/godbus/dbus/v5"
	modemManager "github.com/maltegrosse/go-modemmanager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errNoModemFound = "no modem found"

var newModem3gpp = modemManager.NewModem3gpp

const (
	// Time format for the UpdatedAt field in ConnectionStatus
	timeFormat = time.RFC3339
)

// GSMStatusReader performs live D-Bus reads via go-modemmanager and
// gonetworkmanager to return the runtime status of a GSM/LTE/4G interface.
type GSMStatusReader struct {
	mm modemManager.ModemManager
	// nm nm.NetworkManager
}

// NewGSMModemManager creates a new ModemManager instance for use with GSMStatusReader.
func NewGSMModemManager() (modemManager.ModemManager, error) {
	return modemManager.NewModemManager()
}

// NewGSMStatusReader creates a GSMStatusReader.
func NewGSMStatusReader(modemManager modemManager.ModemManager) *GSMStatusReader {
	return &GSMStatusReader{mm: modemManager}
}

// IsSupport returns true only for NM modem-type devices.
func IsSupport(deviceType nm.NmDeviceType) bool {
	return deviceType == nm.NmDeviceTypeModem
}

// ReadStatus reads the current status of the given NM device and returns a ConnectionStatus proto message.
func (r *GSMStatusReader) ReadStatus(device nm.Device) (*v1.ConnectionStatus, error) {
	// Validate device parameter
	if device == nil {
		return nil, status.Error(codes.InvalidArgument, "device cannot be nil")
	}

	modem, err := r.getModemFromDevice(device)
	if err != nil {
		log.Println("Failed to get modem from device:", err)
		if status.Code(err) == codes.NotFound {
			st := status.Convert(err)
			return &v1.ConnectionStatus{
				ConnectionState:        common.ConnStateNoModem,
				ConnectionFailedReason: st.Message(),
			}, nil
		}
		return nil, err
	}

	cs, err := r.buildInterfaceStatus(modem)
	if err != nil {
		log.Println("Failed to build interface status:", err)
		return nil, status.Errorf(codes.Unavailable, "Interface Status Unavailable: %v", err)
	}
	return cs, nil
}

// getModemFromDevice retrieves the ModemManager modem corresponding to the given NM device.
// It returns an error if the modem is not found or if there is an issue retrieving the modem.
func (r *GSMStatusReader) getModemFromDevice(device nm.Device) (modemManager.Modem, error) {

	udi, err := device.GetPropertyUdi()
	if err != nil {
		log.Printf("GSMStatusReader: GetPropertyUdi failed for device: %v", err)
		return nil, status.Errorf(codes.Unavailable, "Failed to get UDI for device: %v", err)
	}

	modems, err := r.mm.GetModems()
	if err != nil {
		log.Println("Failed to get modems from ModemManager:", err)
		return nil, status.Errorf(codes.Unavailable, "Failed to get modems: %v", err)
	}

	for _, modem := range modems {
		if modem.GetObjectPath() == dbus.ObjectPath(udi) {
			return modem, nil
		}
	}
	return nil, status.Error(codes.NotFound, errNoModemFound)
}

// buildInterfaceStatus constructs a ConnectionStatus proto message from the given ModemManager modem.
// It retrieves the modem's state, access technologies, signal quality, and carrier information.
func (r *GSMStatusReader) buildInterfaceStatus(modem modemManager.Modem) (*v1.ConnectionStatus, error) {
	// Validate modem parameter
	if modem == nil {
		return nil, status.Error(codes.Internal, "modem cannot be nil")
	}

	modemState, err := modem.GetState()
	if err != nil {
		modemState = modemManager.MmModemStateUnknown
		log.Printf("Failed to get modem state for modem %v: %v", modem, err)
	}

	accessTechs, err := modem.GetAccessTechnologies()
	if err != nil {
		accessTechs = []modemManager.MMModemAccessTechnology{
			modemManager.MmModemAccessTechnologyUnknown,
		}
		log.Printf("Failed to get access technologies for modem %v: %v", modem, err)
	}

	signalQuality, _, err := modem.GetSignalQuality()
	if err != nil {
		log.Printf("Failed to get signal quality for modem %v: %v", modem, err)
	}

	failedReason, err := modem.GetStateFailedReason()
	if err != nil {
		log.Printf("Failed to get state failed reason for modem %v: %v", modem, err)
	}

	carrierInfo, err := getCarrierInfo(modem)
	if err != nil {
		log.Printf("Failed to get carrier info for modem %v: %v", modem, err)
		return nil, err
	}

	var bearerConn bool
	if bearers, err := modem.GetBearers(); err == nil {
		for _, b := range bearers {
			connected, err := b.GetConnected()
			if err != nil || !connected {
				log.Printf("Bearer %v is not connected or failed to get connection status: %v", b, err)
				continue
			}
			bearerConn = true
			break
		}
	} else {
		log.Printf("Failed to get bearers for modem %v: %v", modem, err)
	}
	now := time.Now().UTC()
	timeString := now.Format(timeFormat)
	conStatus := &v1.ConnectionStatus{
		ConnectionState:        mapProtoConnectionState(modemState, bearerConn),
		ConnectionFailedReason: mapStateFailedReasonToString(failedReason),
		UpdatedAt:              timeString,
		Details: &v1.ConnectionStatus_Gsm{
			Gsm: &v1.ConnectionStatus_CellularStatus{
				AccessTechnology:     mapProtoAccessTechnology(accessTechs),
				SignalQualityPercent: signalQuality,
				BearerConnected:      bearerConn,
				OperatorName:         carrierInfo.OperatorName,
				OperatorCode:         carrierInfo.OperatorCode,
				RegistrationState:    carrierInfo.RegistrationState,
				RoamingState:         carrierInfo.RoamingState,
			},
		},
	}

	return conStatus, nil
}

/*
	----- Private functions -----
*/

// getCarrierInfo retrieves carrier information from the modem's 3GPP interface.
// Returns partial data if individual fields fail to retrieve;
//
//	returns an empty (non-nil) CellularStatus if the modem doesn't support 3GPP.
func getCarrierInfo(modem modemManager.Modem) (*v1.ConnectionStatus_CellularStatus, error) {
	// Validate modem parameter
	if modem == nil {
		return nil, status.Error(codes.Internal, "modem cannot be nil")
	}

	carrierInfo := &v1.ConnectionStatus_CellularStatus{}

	m3gpp, err := newModem3gpp(modem.GetObjectPath())
	if err != nil {
		log.Printf("Failed to create 3GPP modem interface: %v", err)
		// Not an error if modem doesn't support 3GPP; return empty carrier info
		return carrierInfo, nil
	}
	if m3gpp == nil {
		log.Printf("Modem 3GPP interface is nil for modem: %v", modem)
		// Not an error; return empty carrier info
		return carrierInfo, nil
	}

	if v, err := m3gpp.GetOperatorName(); err == nil {
		carrierInfo.OperatorName = v
	}
	if v, err := m3gpp.GetOperatorCode(); err == nil {
		carrierInfo.OperatorCode = v
	}
	if reg, err := m3gpp.GetRegistrationState(); err == nil {
		carrierInfo.RegistrationState, carrierInfo.RoamingState = mapProtoRegistration(reg)
	}
	return carrierInfo, nil
}
