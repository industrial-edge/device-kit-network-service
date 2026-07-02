/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"log"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/devicemanager"
	"networkservice/internal/networking/interfaces"
	"time"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/google/uuid"
)

type InterfaceStateManager interface {
	Backup() error
	Restore() error
}

type backup struct {
	ifaceName    string
	ifaceType    common.InterfaceType
	wasGateway   bool
	wasActive    bool
	device       nm.Device
	conn         nm.Connection
	connSettings nm.ConnectionSettings
}

type ifaceInfo struct {
	name      string
	ifaceType common.InterfaceType
	isGateway bool
}

type InterfaceStateHandler struct {
	networkManager     nm.NetworkManager
	existingIfaceInfos []*ifaceInfo
	backups            []*backup
	connManager        interfaces.ConnectionHelper
	deviceManager      interfaces.DeviceHelper
}

func NewInterfaceStateHandler(nm nm.NetworkManager, ifaces []*v1.Interface) InterfaceStateManager {
	return &InterfaceStateHandler{
		networkManager:     nm,
		existingIfaceInfos: buildIfaceInfoList(ifaces),
		backups:            make([]*backup, 0),
		connManager:        connectionmanager.NewConnectionManager(nm),
		deviceManager:      devicemanager.NewDeviceManager(nm),
	}
}

func buildIfaceInfoList(ifaces []*v1.Interface) []*ifaceInfo {
	ifaceInfos := make([]*ifaceInfo, 0)

	for _, iface := range ifaces {
		name := iface.GetInterfaceName()
		isGateway := iface.GetGatewayInterface()
		ifaceType := iface.GetInterfaceType()

		ifaceInfos = append(ifaceInfos, &ifaceInfo{
			name:      name,
			isGateway: isGateway,
			ifaceType: common.InterfaceTypeFromProto(ifaceType),
		})
	}
	return ifaceInfos
}

func (h *InterfaceStateHandler) Backup() error {
	log.Println("Backing up connection details for all existing interfaces...")

	devicesMap, err := h.deviceManager.GetAvailableDevicesMap()
	if err != nil {
		return err
	}

	for _, existingIfaceInfo := range h.existingIfaceInfos {
		deviceName := existingIfaceInfo.name

		device, exists := devicesMap[deviceName]
		if !exists {
			log.Printf("No existing device found for name: %s skipping backup...\n", deviceName)
			continue
		}

		if err := h.backupConnection(existingIfaceInfo, device); err != nil {
			h.backups = make([]*backup, 0) // Clear any partial backups on error
			return err
		}
	}

	log.Println("Successfully backed up connection details for all existing interfaces")
	return nil
}

func (h *InterfaceStateHandler) backupConnection(info *ifaceInfo, device nm.Device) error {
	var deviceName string = info.name

	var conn nm.Connection
	var connSettings nm.ConnectionSettings

	log.Println("Backing up connection details for device: ", deviceName, " ...")

	conn, isConnActive, err := h.connManager.FindConnectionWithStatus(device)
	if err != nil {
		return err
	}

	if conn == nil {
		log.Printf("No connection found for device: %s, skipping backup...\n", deviceName)
		return nil
	}

	connSettings, err = h.connManager.GetConnectionSettings(conn)
	if err != nil {
		return err
	}

	if connSettings == nil {
		log.Printf("No connection settings found for device: %s, skipping backup...\n", deviceName)
		return nil
	}

	// We are not re-generating new UUIDs when backing up connection settings for ethernet interfaces only
	// We do not do this for GSM since we want to ensure that the same connection profile gets restored during rollback
	// Furthermore the GSM connection has properties that are not reflected in the connection settings map (like PIN, Username & Password).
	if info.ifaceType == common.InterfaceTypeEthernet {
		log.Println("Building connection settings for backup of ethernet device: ", deviceName, " ...")
		connSettings = h.buildEthernetSettingsForBackup(connSettings)
	}

	h.backups = append(h.backups, &backup{
		ifaceName:    deviceName,
		wasGateway:   info.isGateway,
		ifaceType:    info.ifaceType,
		wasActive:    isConnActive,
		device:       device,
		conn:         conn,
		connSettings: connSettings,
	})

	log.Println("Successfully backed up connection details for device: ", deviceName)
	return nil
}

func (h *InterfaceStateHandler) buildEthernetSettingsForBackup(backup nm.ConnectionSettings) nm.ConnectionSettings {
	connection := make(nm.ConnectionSettings)
	connection[common.ConnectionKey] = make(dict)
	connection[common.IPV4Key] = make(dict)
	connection[common.EthernetType] = make(dict)

	connection[common.ConnectionKey][common.IDKey] = backup[common.ConnectionKey][common.IDKey]
	connection[common.ConnectionKey][common.TypeKey] = backup[common.ConnectionKey][common.TypeKey]
	connection[common.ConnectionKey][common.InterfaceNameKey] = backup[common.ConnectionKey][common.InterfaceNameKey]
	connection[common.ConnectionKey][common.UUIDKey] = uuid.New().String()
	connection[common.ConnectionKey][common.TimeStampKey] = time.Now().UnixNano()

	connection[common.EthernetType] = backup[common.EthernetType]
	connection[common.EthernetType][common.MACAddressKey] = backup[common.EthernetType][common.MACAddressKey]

	connection[common.IPV4Key] = backup[common.IPV4Key]

	return connection
}

func (h *InterfaceStateHandler) Restore() error {
	h.reorderBackups()

	log.Println("Restoring connection settings for all backups...")

	for _, backup := range h.backups {
		if err := h.restoreBackup(backup); err != nil {
			return err
		}
	}

	log.Println("Successfully restored connection settings for all backups")
	return nil
}

func (h *InterfaceStateHandler) restoreBackup(backup *backup) error {
	log.Println("Restoring connection settings for device: ", backup.ifaceName, " ...")

	switch backup.ifaceType {
	case common.InterfaceTypeEthernet:
		if err := h.restoreEthernetConnection(backup); err != nil {
			return err
		}
	case common.InterfaceTypeGSM:
		if err := h.restoreGSMConnection(backup); err != nil {
			return err
		}
	default:
		log.Printf("Unsupported interface type for device: %s, skipping restore...\n", backup.ifaceName)
		return nil
	}

	log.Println("Successfully restored connection settings for device: ", backup.ifaceName)
	return nil
}

func (h *InterfaceStateHandler) restoreEthernetConnection(backup *backup) error {

	device := backup.device

	conn, isActiveConn, err := h.connManager.FindConnectionWithStatus(device)
	if err != nil {
		return err
	}

	if conn != nil {
		// If connection already exists, delete it before restoring the backup connection settings
		// this ensures that only one connection profile exists per ethernet device.
		log.Println("Deleting existing (newly added) ethernet connection as part of rollback")
		if err := h.connManager.DeleteConnection(conn); err != nil {
			return err
		}
	}

	log.Println("Adding backup connection for ethernet device as part of rollback")

	newConn, err := h.connManager.AddConnection(backup.connSettings)
	if err != nil {
		return err
	}

	if err := h.restoreEthernetConnState(backup, isActiveConn, newConn); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceStateHandler) restoreEthernetConnState(backup *backup, isActive bool, newConn nm.Connection) error {
	wasActive := backup.wasActive
	device := backup.device

	if isActive && !wasActive {
		// If the connection is currently active but was not active before, we deactivate it
		// during rollback to restore the original state.
		log.Println("Deactivating ethernet connection as part of rollback")
		if err := h.connManager.DeactivateConnection(device); err != nil {
			return err
		}
		return nil
	}

	if wasActive {
		// Irrespective of whether the connection is active or not, if it was active before, we activate it
		// during rollback to ensure that the original state is restored.
		log.Println("Activating ethernet connection as part of rollback")
		if err := h.connManager.ActivateConnectionWithRetry(newConn, device); err != nil {
			return err
		}
		return nil
	}

	// If the connection was not active before and is not active now, no action is needed
	log.Println("Ethernet connection state does not require changes during rollback")
	return nil
}

func (h *InterfaceStateHandler) restoreGSMConnection(backup *backup) error {
	device := backup.device
	conn := backup.conn
	connSettings := backup.connSettings

	log.Println("Restoring GSM connection settings as part of rollback")
	if err := h.connManager.UpdateConnection(conn, connSettings); err != nil {
		return err
	}

	log.Println("Activating GSM connection with retry as part of rollback")
	if err := h.connManager.ActivateConnectionWithRetry(conn, device); err != nil {
		return err
	}

	return nil
}

func (h *InterfaceStateHandler) reorderBackups() {
	log.Println("Reordering backups to prioritize gateway interfaces...")

	for i, b := range h.backups {
		if b.wasGateway {
			if i != 0 {
				h.backups[0], h.backups[i] = h.backups[i], h.backups[0]
			}
			log.Println("Backups reordered")
			return
		}
	}

	log.Println("No need to reorder backups")
}
