/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	"errors"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	mockgmm "networkservice/internal/networking/mocks/gomodemmanager"
	mockgnm "networkservice/internal/networking/mocks/gonetworkmanager"
	"testing"
	"time"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/godbus/dbus/v5"
	mm "github.com/maltegrosse/go-modemmanager"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_ReadStatus_Success(t *testing.T) {

	mockModemManager := new(mockgmm.MockModemManager)

	gsmStatusReader := NewGSMStatusReader(mockModemManager)

	mockDevice := &mockgnm.MockDeviceWired{}
	mockDevice.On("GetPropertyUdi").Return("/org/modemmanager/Modem/0", nil)
	mockModem := mockgmm.MockModem{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	mockModemManager.On("GetModems").Return([]mm.Modem{&mockModem}, nil)

	mockModem.On("GetObjectPath").Return(dbus.ObjectPath("/org/modemmanager/Modem/0"))
	mockModem.On("GetState").Return(mm.MmModemStateRegistered, nil)
	mockModem.On("GetAccessTechnologies").Return([]mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyLte}, nil)
	mockModem.On("GetSignalQuality").Return(uint32(75), true, nil)
	mockModem.On("GetStateFailedReason").Return(mm.MmModemStateFailedReasonNone, nil)

	mockBearer := &mockgmm.MockBearer{}
	mockBearer.On("GetObjectPath").Return(dbus.ObjectPath("/org/modemmanager/Bearer/0"))
	mockBearer.On("GetConnected").Return(true, nil)
	mockBearer.On("GetIp4Config").Return(mm.BearerIpConfig{Address: "10.10.10.2", Dns1: "1.1.1.1"}, nil)

	mockModem.On("GetBearers").Return([]mm.Bearer{mockBearer}, nil)

	patches.ApplyFunc(getCarrierInfo, func(_ mm.Modem) (*v1.ConnectionStatus_CellularStatus, error) {
		return &v1.ConnectionStatus_CellularStatus{
			OperatorCode:      "12345",
			OperatorName:      "Test Operator",
			RegistrationState: common.RegStateRegistered,
			RoamingState:      common.RoamingStateHome,
		}, nil
	})

	// Call the method under test
	status, err := gsmStatusReader.ReadStatus(mockDevice)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, common.ConnStateDisconnected, status.ConnectionState)
	assert.NotEmpty(t, status.GetUpdatedAt())
	assert.Equal(t, "12345", status.GetGsm().OperatorCode)
	assert.Equal(t, "Test Operator", status.GetGsm().OperatorName)
	assert.Equal(t, common.RegStateRegistered, status.GetGsm().RegistrationState)
	assert.Equal(t, common.AccessTech4G, status.GetGsm().AccessTechnology)
	assert.Equal(t, uint32(75), status.GetGsm().SignalQualityPercent)
	assert.True(t, status.GetGsm().BearerConnected)
	_, parseErr := time.Parse(time.RFC3339, status.GetUpdatedAt())
	assert.NoError(t, parseErr)
}


func Test_ReadStatus_NoModemFound(t *testing.T) {

	mockModemManager := new(mockgmm.MockModemManager)
	gsmStatusReader := NewGSMStatusReader(mockModemManager)
	mockDevice := &mockgnm.MockDeviceWired{}
	mockDevice.On("GetPropertyUdi").Return("/org/modemmanager/Modem/99", nil)
	mockModemManager.On("GetModems").Return([]mm.Modem{}, nil)

	status, err := gsmStatusReader.ReadStatus(mockDevice)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, common.ConnStateNoModem, status.ConnectionState)
	assert.Equal(t, "no modem found", status.ConnectionFailedReason)
}


func Test_getCarrierInfo_Success(t *testing.T) {
	mockModem := &mockgmm.MockModem{}
	mockModem.On("GetObjectPath").Return(dbus.ObjectPath("/org/modemmanager/Modem/0"))

	mockModem3gpp := &mockgmm.MockModem3gpp{}
	mockModem3gpp.On("GetOperatorName").Return("Test Operator", nil)
	mockModem3gpp.On("GetOperatorCode").Return("12345", nil)
	mockModem3gpp.On("GetRegistrationState").Return(mm.MmModem3gppRegistrationStateHome, nil)

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	originalNewModem3gpp := newModem3gpp
	defer func() { newModem3gpp = originalNewModem3gpp }()
	newModem3gpp = func(_ dbus.ObjectPath) (mm.Modem3gpp, error) {
		return mockModem3gpp, nil
	}

	result, err := getCarrierInfo(mockModem)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Operator", result.OperatorName)
	assert.Equal(t, "12345", result.OperatorCode)
	assert.Equal(t, common.RegStateRegistered, result.RegistrationState)
	assert.Equal(t, common.RoamingStateHome, result.RoamingState)
}

func Test_getCarrierInfo_NilModem(t *testing.T) {
	result, err := getCarrierInfo(nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Equal(t, "modem cannot be nil", status.Convert(err).Message())
}

func Test_getCarrierInfo_Nil3GPP(t *testing.T) {
	mockModem := &mockgmm.MockModem{}
	mockModem.On("GetObjectPath").Return(dbus.ObjectPath("/org/modemmanager/Modem/0"))

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	originalNewModem3gpp := newModem3gpp
	defer func() { newModem3gpp = originalNewModem3gpp }()
	newModem3gpp = func(_ dbus.ObjectPath) (mm.Modem3gpp, error) {
		return nil, nil
	}

	result, err := getCarrierInfo(mockModem)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.OperatorName)
	assert.Equal(t, "", result.OperatorCode)
	assert.Equal(t, "", result.RegistrationState)
	assert.Equal(t, "", result.RoamingState)
}

func Test_getCarrierInfo_Get3GPPError(t *testing.T) {
	mockModem := &mockgmm.MockModem{}
	mockModem.On("GetObjectPath").Return(dbus.ObjectPath("/org/modemmanager/Modem/0"))

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	originalNewModem3gpp := newModem3gpp
	defer func() { newModem3gpp = originalNewModem3gpp }()
	newModem3gpp = func(_ dbus.ObjectPath) (mm.Modem3gpp, error) {
		return nil, errors.New("failed to get 3GPP interface")
	}
	result, err := getCarrierInfo(mockModem)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func Test_getCarrierInfo_FieldErrors(t *testing.T) {
	mockModem := &mockgmm.MockModem{}
	mockModem.On("GetObjectPath").Return(dbus.ObjectPath("/org/modemmanager/Modem/0"))

	mockModem3gpp := &mockgmm.MockModem3gpp{}
	mockModem3gpp.On("GetOperatorName").Return("", errors.New("operator name error"))
	mockModem3gpp.On("GetOperatorCode").Return("12345", nil)
	mockModem3gpp.On("GetRegistrationState").Return(mm.MMModem3gppRegistrationState(0), errors.New("registration state error"))

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	originalNewModem3gpp := newModem3gpp
	defer func() { newModem3gpp = originalNewModem3gpp }()
	newModem3gpp = func(_ dbus.ObjectPath) (mm.Modem3gpp, error) {
		return mockModem3gpp, nil
	}

	result, err := getCarrierInfo(mockModem)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.OperatorName)
	assert.Equal(t, "12345", result.OperatorCode)
	assert.Equal(t, "", result.RegistrationState)
	assert.Equal(t, "", result.RoamingState)
}

func TestGSMStatusReader_buildInterfaceStatus_withGetAccessTechnologiesError(t *testing.T) {

	mockModemManager := new(mockgmm.MockModemManager)

	gsmStatusReader := NewGSMStatusReader(mockModemManager)

	mockModem := mockgmm.MockModem{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(getCarrierInfo, func(_ mm.Modem) (*v1.ConnectionStatus_CellularStatus, error) {
		return &v1.ConnectionStatus_CellularStatus{}, nil
	})

	mockModem.On("GetState").Return(mm.MmModemStateRegistered, nil)
	mockModem.On("GetAccessTechnologies").Return(nil, errors.New("get access technologies error"))
	mockModem.On("GetSignalQuality").Return(uint32(75), true, nil)
	mockModem.On("GetStateFailedReason").Return(mm.MmModemStateFailedReasonNone, nil)
	mockModem.On("GetBearers").Return([]mm.Bearer{}, nil)

	result, err := gsmStatusReader.buildInterfaceStatus(&mockModem)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, common.AccessTechUnspecified, result.GetGsm().AccessTechnology)
}

func TestGSMStatusReader_buildInterfaceStatus_withGetSignalQualityError(t *testing.T) {

	mockModemManager := new(mockgmm.MockModemManager)

	gsmStatusReader := NewGSMStatusReader(mockModemManager)
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(getCarrierInfo, func(_ mm.Modem) (*v1.ConnectionStatus_CellularStatus, error) {
		return &v1.ConnectionStatus_CellularStatus{}, nil
	})
	mockModem := mockgmm.MockModem{}
	mockModem.On("GetState").Return(mm.MmModemStateRegistered, nil)
	mockModem.On("GetAccessTechnologies").Return([]mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyLte}, nil)
	mockModem.On("GetSignalQuality").Return(uint32(0), false, errors.New("get signal quality error"))
	mockModem.On("GetStateFailedReason").Return(mm.MmModemStateFailedReasonNone, nil)
	mockModem.On("GetBearers").Return([]mm.Bearer{}, nil)

	result, err := gsmStatusReader.buildInterfaceStatus(&mockModem)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint32(0), result.GetGsm().SignalQualityPercent)
}

func TestGSMStatusReader_buildInterfaceStatus_withGetBearersError(t *testing.T) {

	mockModemManager := new(mockgmm.MockModemManager)

	gsmStatusReader := NewGSMStatusReader(mockModemManager)

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(getCarrierInfo, func(_ mm.Modem) (*v1.ConnectionStatus_CellularStatus, error) {
		return &v1.ConnectionStatus_CellularStatus{
			OperatorCode:      "12345",
			OperatorName:      "Test Operator",
			RegistrationState: common.RegStateRegistered,
			RoamingState:      common.RoamingStateHome,
		}, nil
	})

	mockModem := mockgmm.MockModem{}
	mockModem.On("GetState").Return(mm.MmModemStateConnected, nil)
	mockModem.On("GetAccessTechnologies").Return([]mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyLte}, nil)
	mockModem.On("GetSignalQuality").Return(uint32(75), true, nil)
	mockModem.On("GetStateFailedReason").Return(mm.MmModemStateFailedReasonNone, nil)
	mockModem.On("GetBearers").Return(nil, errors.New("get bearers error"))

	result, err := gsmStatusReader.buildInterfaceStatus(&mockModem)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.GetGsm().BearerConnected)
	assert.Equal(t, common.ConnStateConnecting, result.ConnectionState)
}

func TestSupports(t *testing.T) {
	assert.True(t, IsSupport(nm.NmDeviceTypeModem))
	assert.False(t, IsSupport(nm.NmDeviceTypeWifi))
}
