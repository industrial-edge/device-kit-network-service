/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

// Package gsm provides GSM connection management functionality
// for handling cellular/mobile network connections via NetworkManager.
package gsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"
	"networkservice/internal/networking/connectionmanager"
	"networkservice/internal/networking/interfaces"
	"networkservice/internal/networking/utils"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/google/uuid"
)

// gsmSnapshot stores the state of the active GSM connection for rollback purposes.
// It captures both the NetworkManager connection settings and the persisted file configuration,
// allowing a complete restoration if the new connection creation fails.
// Only the active, valid GSM connection is backed up since the device supports only one
// GSM connection at a time and there is only one shared GSM config file.
type gsmSnapshot struct {
	connectionSettings nm.ConnectionSettings
	hasConnection      bool
	// fileConfig holds the parsed GSM config file content as a map (e.g. {"Apn":"...", "Pin":"..."}).
	fileConfig    map[string]string
	hasFileConfig bool
	// wasActive indicates whether the backed-up connection was active at backup time.
	// During rollback, if wasActive is false, we only re-create the connection (AddConnection)
	// without activating it.
	wasActive bool
}

// GSMConnection manages GSM/cellular network connections through NetworkManager.
// It handles the complete lifecycle of GSM connections including creation, activation,
// configuration persistence, and cleanup.
type GSMConnection struct {
	// networkManager provides access to NetworkManager D-Bus interface
	networkManager nm.NetworkManager
	// connectionSettingFromNM holds the NetworkManager connection configuration
	connectionSettingFromNM nm.ConnectionSettings
	connManager             interfaces.ConnectionHelper
	fileUtils               utils.FileSystemOperations
	nmSettings              nm.Settings
	// gsmDevice caches the GSM device found during CreateConnection for reuse in rollback
	gsmDevice nm.Device
	// gsmBackup stores the backup of existing GSM connections for rollback
	gsmBackup *gsmSnapshot
}

// NewGSMHandler creates and initializes a new GSM connection handler.

func NewGSMHandler(networkManager nm.NetworkManager) (*GSMConnection, error) {
	log.Println("Initializing new GSM connection handler")

	connection := nm.ConnectionSettings{
		common.ConnectionKey: make(map[string]interface{}), // Connection metadata (ID, type, etc.)
		common.IPV4Key:       make(map[string]interface{}), // IPv4 configuration settings
		common.GSMSetting:    make(map[string]interface{}), // GSM-specific settings (APN, PIN, etc.)
	}
	settingsM, err := newSettings()
	if err != nil {
		log.Printf("ERROR: Failed to create NetworkManager settings interface: %v", err)
		return nil, err
	}

	return &GSMConnection{
		networkManager:          networkManager,
		connectionSettingFromNM: connection,
		connManager:             connectionmanager.NewConnectionManager(networkManager),
		nmSettings:              settingsM,
		fileUtils:               &utils.OsFileSystemOperations{},
	}, nil
}

func newSettings() (nm.Settings, error) {
	return nm.NewSettings()
}

// CreateConnection establishes a new GSM connection using the provided settings.
// This method orchestrates the complete GSM connection setup process including:
// 1. Configuration validation
// 2. GSM device detection
// 3. Connection preparation
// 4. Backup of existing connections for rollback
// 5. Cleanup of existing connections
// 6. Connection activation
// 7. Configuration persistence
// If any step after backup fails, a deferred rollback restores the previous state.

func (g *GSMConnection) CreateConnection(settings *v1.ConnectionSettings) error {
	log.Printf("Starting GSM connection creation process for connection: %s", settings.Name)
	// Step 1: Extract GSM configuration from connection settings
	gsmConfig, err := g.getGSMInterface(settings)
	if err != nil {
		return fmt.Errorf("failed to get GSM interface: %w", err)
	}
	log.Printf("GSM configuration extracted successfully for APN: %s", gsmConfig.Gsm.Apn)

	// Step 2: Validate GSM configuration parameters

	err = g.validateGSMConfig(gsmConfig)
	if err != nil {
		return err
	}
	log.Println("GSM configuration validation passed")

	// Step 3: Locate available GSM device

	gsmDevice, err := g.findGSMDevice()
	if err != nil {
		return err
	}
	g.gsmDevice = gsmDevice
	log.Printf("GSM device found at path: %s", gsmDevice.GetPath())

	// Step 4: Prepare connection settings with device-specific information

	err = g.prepareGSMConnection(settings.Name, gsmConfig, gsmDevice)
	if err != nil {
		log.Printf("ERROR: Failed to prepare GSM connection: %v", err)
		return fmt.Errorf("failed to prepare GSM connection: %w", err)
	}

	// Step 5: Backup existing GSM connections before deletion for rollback support
	if err := g.doBackup(); err != nil {
		return fmt.Errorf("failed to backup existing GSM state: %w", err)
	}

	// Defer rollback: if rollbackNeeded is set to true by a subsequent step,
	// the deferred function will attempt to restore the previous state.
	var rollbackOnErr bool
	defer func() {
		if rollbackOnErr {
			log.Println("Rollback triggered due to error during GSM connection creation")
			if rbErr := g.doRollback(); rbErr != nil {
				log.Printf("ERROR: Rollback failed: %v", rbErr)
			}
		}
	}()
	// Step 6: Clean up any existing GSM connections to avoid conflicts
	err = g.deactivateAndDeleteExistingGSMConnections()
	if err != nil {
		log.Printf("ERROR: Failed to delete existing GSM connections: %v", err)
		rollbackOnErr = true
		return fmt.Errorf("failed to delete existing GSM connections: %w", err)
	}
	log.Println("Existing GSM connections cleanup completed")

	// Step 7: Add the new GSM connection to NetworkManager

	conn, err := g.addConnectionWithGSM()
	if err != nil {
		log.Printf("ERROR: Failed to add GSM connection: %v", err)
		rollbackOnErr = true
		return err
	}
	log.Println("GSM connection added to NetworkManager successfully")

	// Step 8: Activate the connection and verify activation status

	// Activate the connection to apply the updated connection settings to the device.
	// This also ensures the IP route table is updated when setting GSM interface as default gateway.
	err = g.connManager.ActivateConnectionWithRetry(conn, gsmDevice)
	if err != nil {
		rollbackOnErr = true
		return g.handleActivationFailure(err)
	}

	log.Println("GSM connection activated successfully")
	// Step 9: Persist GSM configuration to file for future reference

	err = g.writeGSMConfigToFile(gsmConfig)
	if err != nil {
		log.Printf("ERROR: Failed to write GSM config to file: %v", err)
		rollbackOnErr = true
		return fmt.Errorf("failed to persist GSM configuration: %w", err)
	}
	log.Printf("GSM connection creation completed successfully for: %s", settings.Name)
	return nil
}

// handleActivationFailure performs cleanup after activation failure and returns
// the original activation error unless cleanup itself fails.
func (g *GSMConnection) handleActivationFailure(activationErr error) error {
	log.Println("GSM connection not activated successfully", activationErr)
	if err := g.deactivateAndDeleteExistingGSMConnections(); err != nil {
		return err
	}
	return activationErr
}

// RemoveConnection removes an existing GSM connection and cleans up associated resources.
// This method performs the following operations:
// 1. Validates the connection settings contain GSM configuration
// 2. Removes the connection from NetworkManager
// 3. Cleans up persisted configuration files

func (g *GSMConnection) RemoveConnection(settings *v1.ConnectionSettings) error {
	log.Printf("Starting GSM connection removal process for connection: %s", settings.Name)

	// Step 1: Validate that the settings contain GSM configuration

	if settings.ConnectionType != v1.ConnectionSettings_GSM || settings.Name == "" {
		log.Println("ERROR: Minimal parameters required for GSM connection removal are missing")
		return fmt.Errorf("minimal parameters(connection type gsm and name) required for GSM connection removal are missing")
	}

	log.Println("GSM interface validation passed for removal")
	// Step 2: Remove the GSM connection from NetworkManager
	targetName := settings.Name

	err := g.deleteGSMConnection(targetName)
	if err != nil {
		log.Printf("ERROR: Failed to delete GSM connection '%s': %v", targetName, err)
		return fmt.Errorf("gsm connection deletion failed: %w", err)
	}
	log.Printf("GSM connection '%s' removed from NetworkManager successfully", targetName)

	// Step 3: Remove persisted GSM configuration files

	err = os.RemoveAll(common.GSMConfigParentPath)
	if err != nil {
		log.Printf("WARNING: Failed to remove GSM config file: %v", err)
		// Note: We don't return error here as the main connection removal was successful
	} else {
		log.Println("GSM configuration files cleaned up successfully ")
	}

	log.Printf("GSM connection removal completed successfully for: %s", settings.Name)
	return nil
}

// validateGSMConfig validates the GSM configuration parameters to ensure they meet
// the requirements for establishing a GSM connection. This includes checking for
// required fields like APN and validating their format.
func (g *GSMConnection) validateGSMConfig(gsmConfig *v1.ConnectionSettings_Gsm) error {

	err := HasValidGSMConfig(gsmConfig)
	if err != nil {
		log.Printf("ERROR: GSM configuration validation failed: %v", err)
		return fmt.Errorf("gsm configuration validation error: %w", err)
	}

	return nil
}

// findGSMDevice locates and returns the first available GSM/modem device from
// the system's network devices. This method queries NetworkManager for all
// available devices and filters for GSM-capable modems.

func (g *GSMConnection) findGSMDevice() (nm.Device, error) {
	devices, err := g.networkManager.GetDevices()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve network devices from NetworkManager: %v", err)
		return nil, err
	}

	gsmDevice, err := g.findGSMDeviceFromList(devices)
	if err != nil {
		log.Printf("ERROR: No GSM device found in device list: %v", err)
		return nil, fmt.Errorf("gsm device lookup failed: %w", err)
	}
	return gsmDevice, nil
}

// prepareGSMConnection configures the NetworkManager connection settings with
// GSM-specific parameters including device interface name, connection name,
// and GSM configuration details (APN, authentication, etc.).
//

func (g *GSMConnection) prepareGSMConnection(connectionName string, gsmConf *v1.ConnectionSettings_Gsm, gsmDevice nm.Device) error {
	log.Printf("Preparing NetworkManager connection settings for GSM connection: %s", connectionName)

	// Extract device interface name (e.g., wwan0, ppp0) from the GSM device
	// For GSM modules, the device name (e.g., /dev/ttyUSBx) is abstracted into a logical interface name (e.g., wwan0),
	// so we retrieve the deviceName from the interfaceName property in nmcli.
	// This behavior is different from Ethernet devices, where the device name directly matches the interface name (e.g., eth0).
	deviceName, err := gsmDevice.GetPropertyInterface()
	if err != nil {
		log.Printf("ERROR: Failed to get GSM device interface name: %v", err)
		return fmt.Errorf("could not get GSM device interface name: %w", err)
	}
	log.Printf("GSM device interface name retrieved: %s", deviceName)

	// Apply GSM configuration to NetworkManager connection settings
	SetConnectionDetailsWithGSM(g.connectionSettingFromNM, gsmConf, deviceName, connectionName)
	log.Printf("GSM connection preparation completed for device %s with connection name: %s", deviceName, connectionName)
	return nil
}

// SetConnectionDetailsWithGSM sets the connection ID, UUID, and timestamp.
// It also sets GSM-specific configuration details.
func SetConnectionDetailsWithGSM(connection nm.ConnectionSettings, gsmConfig *v1.ConnectionSettings_Gsm,
	deviceName string, connectionName string) {
	connection[common.ConnectionKey][common.IDKey] = connectionName // Connection Name
	connection[common.ConnectionKey][common.UUIDKey] = uuid.New().String()
	connection[common.ConnectionKey][common.TimeStampKey] = time.Now().Unix()
	connection[common.ConnectionKey][common.TypeKey] = common.GSMSetting
	connection[common.ConnectionKey][common.AutoConnectKey] = true
	connection[common.ConnectionKey][common.InterfaceNameKey] = deviceName // interface name
	connection[common.GSMSetting][common.APNKey] = gsmConfig.Gsm.Apn
	connection[common.IPV4Key][common.MethodKey] = common.Auto

	// Only set username and password if they are not empty
	if len(gsmConfig.Gsm.Username) > 0 && len(gsmConfig.Gsm.Password) > 0 {
		connection[common.GSMSetting][common.UsernameKey] = gsmConfig.Gsm.Username
		connection[common.GSMSetting][common.PasswordKey] = gsmConfig.Gsm.Password
	}
	if len(gsmConfig.Gsm.Pin) > 0 {
		connection[common.GSMSetting][common.PINKey] = gsmConfig.Gsm.Pin
	}
}

// getGSMInterface extracts and validates GSM configuration from the connection settings.
// This method ensures the connection type is GSM and that the GSM configuration
// wrapper contains valid GSM-specific settings.

func (g *GSMConnection) getGSMInterface(newSettings *v1.ConnectionSettings) (*v1.ConnectionSettings_Gsm, error) {

	// Validate connection type is GSM
	if newSettings.ConnectionType != v1.ConnectionSettings_GSM {
		log.Printf("ERROR: Invalid connection type. Expected GSM, got: %v", newSettings.ConnectionType)
		return nil, fmt.Errorf("connection type is not GSM, got: %v", newSettings.ConnectionType)
	}
	// Extract GSM configuration wrapper from oneof field
	gsmWrapper, ok := newSettings.ConnectionConfig.(*v1.ConnectionSettings_Gsm)
	if !ok || gsmWrapper.Gsm == nil {
		log.Println("ERROR: GSM connection type specified but GSM configuration is missing or invalid")
		return nil, errors.New("GSM connection selected but GSM config is missing")
	}

	return gsmWrapper, nil
}

// writeGSMConfigToFile persists GSM configuration to a JSON file for future reference
// and recovery purposes. The configuration includes APN, PIN, and authentication
// credentials (if provided). Sensitive data like PINs and passwords are included
// for functional requirements but should be handled securely.

func (g *GSMConnection) writeGSMConfigToFile(protoData *v1.ConnectionSettings_Gsm) error {
	log.Println("Preparing to persist GSM configuration to file system")

	gsmConfig := protoData.Gsm
	log.Printf("Creating configuration data structure for APN: %s", gsmConfig.Apn)

	// Build configuration map with required and optional parameters
	configContent := map[string]string{
		common.GSMConfigKeyAPN: gsmConfig.Apn, // Always required
		common.GSMConfigKeyPIN: gsmConfig.Pin,
	}

	// Add optional authentication credentials if provided
	if len(gsmConfig.Username) > 0 {
		log.Println("Adding username to GSM configuration")
		configContent[common.GSMConfigKeyUsername] = gsmConfig.Username
	}
	if len(gsmConfig.Password) > 0 {
		log.Println("Adding password to GSM configuration")
		configContent[common.GSMConfigKeyPassword] = gsmConfig.Password
	}

	// Serialize configuration to JSON format
	log.Println("Serializing GSM configuration to JSON format")
	buffer, err := json.Marshal(configContent)
	if err != nil {
		log.Println("ERROR: Failed to marshal GSM config to JSON:")
		return fmt.Errorf("failed to marshal GSM config to JSON: %v", err)
	}

	// Ensure clean file system state and create directory structure
	log.Println("Persisting GSM config at file system")

	// Remove existing config file if it exists, but preserve parent directory structure
	if err := g.fileUtils.Stat(common.GSMConfigPath); err == nil {
		log.Println("Removing existing GSM config file")
		err = g.fileUtils.RemoveAll(common.GSMConfigPath)
		if err != nil {
			log.Println("ERROR: Failed to remove existing GSM config file:")
			return fmt.Errorf("failed to remove existing GSM config file %v", err)
		}
	}

	// Create directory structure using the explicit parent path
	err = g.fileUtils.MkdirAll(common.GSMConfigParentPath, os.FileMode(0755)) // Readable and executable by all users
	if err != nil {
		log.Println("ERROR: Failed to create GSM config directory:")
		return fmt.Errorf("failed to create GSM config directory: %v", err)
	}

	// Write configuration to file with secure permissions
	log.Println("Writing GSM configuration to file")
	err = g.fileUtils.WriteFile(common.GSMConfigPath, buffer, os.FileMode(0600)) // Readable by owner only
	if err != nil {
		log.Println("ERROR: Failed to write GSM config to file:")
		return fmt.Errorf("failed to write GSM config to file: %v", err)
	}

	log.Println("GSM configuration successfully persisted to file system")
	return nil
}

// findGSMDeviceFromList searches through a list of network devices to locate
// the first available GSM modem device. This method filters devices by type
// to find NetworkManager modem devices suitable for GSM connections.

func (g *GSMConnection) findGSMDeviceFromList(devices []nm.Device) (nm.Device, error) {
	log.Printf("Searching for GSM modem in %d network devices", len(devices))

	var gsmDevice nm.Device

	for _, device := range devices {

		deviceType, err := device.GetPropertyDeviceType()
		if err != nil {
			log.Printf("WARNING: Failed to get device type for %s: %v", device.GetPath(), err)
			continue
		}

		// Check if this device is a GSM modem
		// Assumption that there is only one GSM module connected at a time,
		// hence considering the first found GSM device as the intended one
		if deviceType == nm.NmDeviceTypeModem {
			log.Printf("GSM modem device found: %s", device.GetPath())
			gsmDevice = device
			break
		}
	}

	// Validate that we found a GSM device
	if gsmDevice == nil {
		errMsg := "no GSM modem device found among network devices"
		log.Printf("ERROR: %s", errMsg)
		return nil, errors.New(errMsg)
	}

	log.Printf("GSM device selection completed successfully: %s", gsmDevice.GetPath())
	return gsmDevice, nil
}

// activeConnection activates a GSM connection using the specified connection profile
// and GSM device. This method triggers NetworkManager to establish the actual
// cellular connection using the prepared configuration.
//
// Parameters:
//   - conn: NetworkManager connection profile to activate
//   - gsmDevice: GSM device to use for the connection
//
// Returns:
//   - nm.ActiveConnection: Handle to the activated connection for status monitoring
//   - error: Error if connection activation fails
func (g *GSMConnection) activateConnection(conn nm.Connection, gsmDevice nm.Device) (nm.ActiveConnection, error) {

	// Request NetworkManager to activate the connection
	activeConn, err := g.networkManager.ActivateConnection(conn, gsmDevice, nil)
	if err != nil {
		log.Printf("ERROR: Failed to activate GSM connection: %v", err)
		return nil, fmt.Errorf("failed to activate connection: %w", err)
	}

	log.Println("GSM connection activation initiated successfully")
	return activeConn, nil
}

// addConnectionWithGSM adds the prepared GSM connection configuration to NetworkManager.
// This method creates a persistent connection profile that can be activated and
// managed by NetworkManager. The connection settings must be prepared before calling.
//
// Returns:
//   - nm.Connection: Handle to the created NetworkManager connection
//   - error: Error if connection creation fails
func (g *GSMConnection) addConnectionWithGSM() (nm.Connection, error) {
	return g.addNmConnection(g.connectionSettingFromNM)
}

// addNmConnection adds a connection profile to NetworkManager with the given settings.
func (g *GSMConnection) addNmConnection(settings nm.ConnectionSettings) (nm.Connection, error) {
	conn, err := g.nmSettings.AddConnection(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to add GSM connection: %w", err)
	}
	log.Printf("GSM connection added to NetworkManager: %s", conn.GetPath())
	return conn, nil
}

// getAllConnections retrieves all saved NetworkManager connection profiles from
// the system. This method provides access to the complete connection database
// for operations like searching, cleanup, and validation.
//
// Returns:
//   - []nm.Connection: List of all NetworkManager connections
//   - error: Error if connection retrieval fails
func (g *GSMConnection) getAllConnections() ([]nm.Connection, error) {
	log.Println("Retrieving all NetworkManager connection profiles")

	// Query NetworkManager for all stored connections
	connections, err := g.nmSettings.ListConnections()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve connection list from NetworkManager: %v", err)
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}

	log.Printf("Retrieved %d connection profiles from NetworkManager", len(connections))
	return connections, nil
}

// deleteGSMConnection removes a specific GSM connection from NetworkManager by name.
// This method searches for connections with the specified name and removes them
// from the NetworkManager connection database.
//
// Parameters:
//   - targetName: Name of the GSM connection to delete
//
// Returns:
//   - error: nil if deletion succeeds or connection doesn't exist, error if deletion fails
func (g *GSMConnection) deleteGSMConnection(targetName string) error {
	log.Printf("Searching for GSM connection to delete: %s", targetName)

	// Search for existing GSM connection with the specified name
	// Using constants.IDKey to match against connection ID field in NetworkManager
	connList := g.findConnectionBySettingValue(common.IDKey, targetName)
	if len(connList) > 0 {
		for _, conn := range connList {
			log.Printf("Found existing GSM connection to delete: %s", targetName)

			// Attempt to delete the connection from NetworkManager
			if err := conn.Delete(); err != nil {
				log.Printf("ERROR: Failed to delete GSM connection '%s': %v", targetName, err)
				return err
			}

			log.Printf("GSM connection '%s' deleted successfully", targetName)
		}
	} else {
		log.Printf("No GSM connection found with %s=%s", common.IDKey, targetName)

	}

	return nil
}

// findConnectionBySettingValue searches for a GSM connection with the specified setting key-value pair.
// This method examines all NetworkManager connections to find ones that match the
// given criteria (typically connection ID or type).
//
// Parameters:
//   - connectionSettingKey: Setting key to search for (e.g., "id", "type")
//   - connectionSettingValue: Expected value for the setting key
//
// Returns:
//   - nm.Connection: Found connection matching the criteria, or nil if not found
func (g *GSMConnection) findConnectionBySettingValue(connectionSettingKey, connectionSettingValue string) []nm.Connection {
	log.Printf("Searching for GSM connection with %s=%s", connectionSettingKey, connectionSettingValue)

	connections, err := g.getAllConnections()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve connections for search: %v", err)
		return nil
	}
	matchingConnections := make([]nm.Connection, 0)
	for _, conn := range connections {

		// Get connection properties for examination
		props, err := conn.GetSettings()
		if err != nil {
			log.Printf("WARNING: Failed to get settings for connection %v", err)
			continue
		}

		// Check if connection has the required setting structure
		connSettings, exists := props[common.ConnectionKey]
		if !exists {
			log.Println("WARNING: Connection missing in connection settings")
			continue
		}

		// Extract the setting value and compare
		if actualValue, exists := connSettings[connectionSettingKey]; exists {
			if actualValueStr, ok := actualValue.(string); ok {

				if actualValueStr == connectionSettingValue {
					log.Println("GSM connection match found:")
					matchingConnections = append(matchingConnections, conn)

				}
			}
		}
	}
	return matchingConnections
}

// deactivateAndDeleteExistingGSMConnections deactivates the current GSM device
// connection and removes all existing GSM connections from NetworkManager.
// This method performs cleanup by searching for all connections with GSM type
// and removing them to avoid conflicts with new GSM connections.
// We assume (and accept the risk) that the device can only support one active GSM connection at a time.
// Thus, we clean up any existing GSM connection before creating a new one to prevent unmanaged connections.
func (g *GSMConnection) deactivateAndDeleteExistingGSMConnections() error {
	log.Println("Starting cleanup of existing GSM connections")
	if err := g.connManager.DeactivateConnection(g.gsmDevice); err != nil {
		return err
	}
	// Search for existing GSM connections by type
	// Using constants.TypeKey and constants.GSMSetting to identify GSM-type connections
	connList := g.findConnectionBySettingValue(common.TypeKey, common.GSMSetting)
	if len(connList) > 0 {
		for _, conn := range connList {
			log.Printf("Found existing GSM connection to clean up: %s", conn.GetPath())

			// Attempt to delete the existing GSM connection
			if err := conn.Delete(); err != nil {
				log.Printf("ERROR: Failed to delete existing GSM connection %s: %v", conn.GetPath(), err)
				return err
			} else {
				log.Printf("Successfully deleted existing GSM connection: %s", conn.GetPath())
			}
		}
	} else {
		log.Println("INFO: No existing GSM connections found - cleanup not needed")
	}

	log.Println("GSM connection cleanup completed")
	return nil
}

// doBackup stores file + connection state before CreateConnection mutates NM.
// File is read first because it is the service-owned source of truth.
func (g *GSMConnection) doBackup() error {
	log.Println("Checking for existing GSM state to back up (config file + NM connection)")

	backup := &gsmSnapshot{}

	// File first: connection matching depends on file values.
	if err := g.backupConfigFile(backup); err != nil {
		return fmt.Errorf("gsm Config file backup failed: %w", err)
	}

	if err := g.backupConnection(backup); err != nil {
		return fmt.Errorf("gsm connection backup failed: %w", err)
	}

	g.gsmBackup = backup

	log.Printf("Backup result: fileConfig=%t, connection=%t, wasActive=%t",
		backup.hasFileConfig, backup.hasConnection, backup.wasActive)
	return nil
}

// backupConfigFile loads the GSM file snapshot used as rollback source of truth.
// Only file read I/O is fatal; missing/invalid content is treated as skip.

func (g *GSMConnection) backupConfigFile(backup *gsmSnapshot) error {
	if err := g.fileUtils.Stat(common.GSMConfigPath); err != nil {
		// If the config file does not exist, this is either a fresh device with no prior GSM
		// configuration or the file creation failed. In both cases there is no source of
		// truth to back up, so we return nil (no error) and skip the backup entirely.
		log.Println("WARNING: No existing GSM config file found — nothing to backup")
		return nil
	}
	data, err := g.fileUtils.ReadFile(common.GSMConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read GSM config file for backup: %w", err)
	}
	if len(data) == 0 {
		log.Println("WARNING: GSM config file is empty — nothing to backup")
		return nil
	}
	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		log.Println("WARNING: GSM config file contains invalid JSON ERROR:-", err)
		return nil
	}
	if err := ValidateGSMConfigFileContent(parsed); err != nil {
		log.Println("WARNING: GSM config file validation failed ERROR:-", err)
		return nil
	}
	backup.fileConfig = parsed
	backup.hasFileConfig = true
	log.Println("Backed up GSM config file successfully")
	return nil
}

// backupConnection captures the best NM candidate that matches the file snapshot.
// Active is preferred; otherwise first matching inactive is used.
func (g *GSMConnection) backupConnection(backup *gsmSnapshot) error {
	// Without a config file we have no source of truth to validate or restore from,
	// so backing up a connection whose correctness we cannot verify is not useful.
	if !backup.hasFileConfig {
		log.Println("WARNING: No config file available — skipping connection backup (no source of truth for recovery)")
		return nil
	}

	gsmConn, isActive, err := g.connManager.FindConnectionWithStatus(g.gsmDevice)
	if err != nil {
		return fmt.Errorf("failed to find connection with status during backup: %w", err)
	}

	// Case 1: Active connection found — back it up and override with file values if mismatch
	if gsmConn != nil && isActive {
		log.Printf("Found active GSM connection for backup: %s", gsmConn.GetPath())
		if err := g.storeConnectionBackup(backup, gsmConn, true); err != nil {
			return fmt.Errorf("failed to store active connection backup: %w", err)
		}
		if !g.connectionMatchesConfigFile(gsmConn, backup.fileConfig) {
			log.Println("WARNING: Active GSM connection does not match config file — overriding backup with file values")
			g.overrideBackupWithFileConfig(backup)
		}
		return nil
	}

	// Case 2: No active connection — search inactive GSM connections matching config file

	log.Println("No active GSM connection found — searching inactive GSM connections matching config file")
	matchedConn := g.findInactiveConnectionMatchingFile(backup.fileConfig)
	if matchedConn == nil {
		log.Println("No inactive GSM connection matching config file found — skipping connection backup")
		return nil
	}

	log.Printf("Found inactive GSM connection matching config file: %s", matchedConn.GetPath())
	if err := g.storeConnectionBackup(backup, matchedConn, false); err != nil {
		return fmt.Errorf("failed to store inactive connection backup: %w", err)
	}
	return nil
}

// storeConnectionBackup snapshots NM settings for rollback recreation.
func (g *GSMConnection) storeConnectionBackup(backup *gsmSnapshot, conn nm.Connection, wasActive bool) error {
	settings, err := conn.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings for GSM connection %s: %w", conn.GetPath(), err)
	}
	backup.connectionSettings = settings
	backup.hasConnection = true
	backup.wasActive = wasActive
	log.Printf("Backed up GSM connection: %s (wasActive=%t)", conn.GetPath(), wasActive)
	return nil
}

// findInactiveConnectionMatchingFile selects the first inactive GSM profile matching file data.
func (g *GSMConnection) findInactiveConnectionMatchingFile(fileConfig map[string]string) nm.Connection {
	connList := g.findConnectionBySettingValue(common.TypeKey, common.GSMSetting)
	for _, conn := range connList {
		if g.connectionMatchesConfigFile(conn, fileConfig) {
			return conn
		}
	}
	return nil
}

// overrideBackupWithFileConfig enforces file-wins behavior when NM diverges.
func (g *GSMConnection) overrideBackupWithFileConfig(backup *gsmSnapshot) {
	fileConfig := backup.fileConfig
	gsmSection, ok := backup.connectionSettings[common.GSMSetting]
	if !ok {
		gsmSection = make(map[string]interface{})
		backup.connectionSettings[common.GSMSetting] = gsmSection
	}

	// APN is required; optional fields follow set-or-delete behavior below.
	if apn, exists := fileConfig[common.GSMConfigKeyAPN]; exists && apn != "" {
		gsmSection[common.APNKey] = apn
	}

	// Override optional fields — set if present in file, remove if absent
	fileKeyToNMKey := map[string]string{
		common.GSMConfigKeyUsername: common.UsernameKey,
		common.GSMConfigKeyPassword: common.PasswordKey,
		common.GSMConfigKeyPIN:      common.PINKey,
	}
	for fileKey, nmKey := range fileKeyToNMKey {
		overrideOptionalField(gsmSection, fileConfig, fileKey, nmKey)
	}
}

// overrideOptionalField deletes missing optional values to prevent stale credential restore.
func overrideOptionalField(gsmSection map[string]interface{}, fileConfig map[string]string, fileKey, nmKey string) {
	if val, exists := fileConfig[fileKey]; exists && val != "" {
		gsmSection[nmKey] = val
	} else {
		delete(gsmSection, nmKey)
	}
}

// connectionMatchesConfigFile checks whether NM data matches file snapshot.
// Optional fields are compared only when present in file data.
func (g *GSMConnection) connectionMatchesConfigFile(conn nm.Connection, fileConfig map[string]string) bool {
	connFields, err := extractGSMConnFields(conn)
	if err != nil {
		log.Printf("WARNING: Failed to collect GSM connection fields for validation: %v", err)
		return false
	}

	// APN is required — always compare
	if fileConfig[common.GSMConfigKeyAPN] != connFields[common.GSMConfigKeyAPN] {
		log.Printf("APN mismatch: file=%q, connection=%q", fileConfig[common.GSMConfigKeyAPN], connFields[common.GSMConfigKeyAPN])
		return false
	}

	// Optional fields — only compare if present and non-empty in file
	optionalKeys := []string{common.GSMConfigKeyUsername, common.GSMConfigKeyPIN, common.GSMConfigKeyPassword}
	for _, key := range optionalKeys {
		fileVal, exists := fileConfig[key]
		if !exists || fileVal == "" {
			continue
		}
		if fileVal != connFields[key] {
			log.Printf("%s mismatch between config file and NM connection", key)
			return false
		}
	}

	return true
}

// extractGSMConnFields normalizes settings and secrets into file-key format.
func extractGSMConnFields(conn nm.Connection) (map[string]string, error) {
	settings, err := conn.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	secrets, err := conn.GetSecrets(common.GSMSetting)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}
	return map[string]string{
		common.GSMConfigKeyAPN:      extractStringFromSettings(settings, common.GSMSetting, common.APNKey),
		common.GSMConfigKeyUsername: extractStringFromSettings(settings, common.GSMSetting, common.UsernameKey),
		common.GSMConfigKeyPIN:      extractStringFromSettings(secrets, common.GSMSetting, common.PINKey),
		common.GSMConfigKeyPassword: extractStringFromSettings(secrets, common.GSMSetting, common.PasswordKey),
	}, nil
}

// doRollback restores what was actually captured during backup.
// File and connection are restored independently because either may be absent.
func (g *GSMConnection) doRollback() error {
	if g.gsmBackup == nil {
		log.Println("No backup available, skipping rollback")
		return nil
	}
	log.Println("Starting GSM rollback")

	var rollbackErrs []error
	if g.gsmBackup.hasConnection {
		if err := g.restoreConnection(); err != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("restore connection: %w", err))
		}
	}

	if g.gsmBackup.hasFileConfig {
		if err := g.restoreConfigFile(); err != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("restore config file: %w", err))
		}
	}

	if len(rollbackErrs) > 0 {
		return fmt.Errorf("rollback failed: %w", errors.Join(rollbackErrs...))
	}

	log.Println("GSM rollback completed successfully")
	return nil
}

// restoreConnection recreates the backed-up profile and re-activates only if previously active.
func (g *GSMConnection) restoreConnection() error {
	settings := g.gsmBackup.connectionSettings
	log.Println("Before reset connection settings", settings)
	// Regenerate identity fields to avoid collisions with removed profiles.
	if connMap, ok := settings[common.ConnectionKey]; ok {
		connMap[common.UUIDKey] = uuid.New().String()
		connMap[common.TimeStampKey] = time.Now().Unix()
	}
	log.Println("Before Ipv6", settings)
	if settings[common.IPV6Key] != nil {
		delete(settings[common.IPV6Key], common.AddressesKey)
		delete(settings[common.IPV6Key], common.RoutesKey)
	}
	log.Println("After Ipv6", settings)

	conn, err := g.addNmConnection(settings)
	if err != nil {
		return fmt.Errorf("could not restore GSM connection: %w", err)
	}

	// Only activate if the original connection was active before backup.
	// Activation failure is non-fatal because the connection profile already
	// exists in NM and can be activated manually or on next reboot.
	if g.gsmBackup.wasActive {
		if err := g.activateRestoredConnection(conn); err != nil {
			return fmt.Errorf("could not activate GSM connection: %w", err)
		}
	}
	return nil
}

// activateRestoredConnection attempts to activate a restored GSM connection.
func (g *GSMConnection) activateRestoredConnection(conn nm.Connection) error {
	if err := g.connManager.ActivateConnectionWithRetry(conn, g.gsmDevice); err != nil {
		return fmt.Errorf("failed to activate restored GSM connection: %w", err)
	}
	log.Println("Restored GSM connection activated successfully during rollback")
	return nil
}

// restoreConfigFile restores persisted file state needed for future recovery.
func (g *GSMConnection) restoreConfigFile() error {
	data, err := json.Marshal(g.gsmBackup.fileConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal GSM config for restore: %w", err)
	}
	if err := g.fileUtils.MkdirAll(common.GSMConfigParentPath, common.GSMConfigDirPermission); err != nil {
		return fmt.Errorf("failed to create GSM config directory: %w", err)
	}
	if err := g.fileUtils.WriteFile(common.GSMConfigPath, data, common.GSMConfigFilePermission); err != nil {
		return fmt.Errorf("failed to write GSM config file: %w", err)
	}
	return nil
}
