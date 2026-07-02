/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package connectionmanager

import (
	"errors"
	"log"
	"time"

	"networkservice/internal/networking/common"
	"networkservice/internal/networking/interfaces"

	nm "github.com/Wifx/gonetworkmanager/v2"
)

type ConnectionManager struct {
	nm           nm.NetworkManager
	maxRetries   int
	initialDelay time.Duration
}

// NewConnectionManager creates a new instance of ConnectionManager without any retry configuration.
// Use WithRetries to set retry options.
func NewConnectionManager(nm nm.NetworkManager) interfaces.ConnectionHelper {
	return &ConnectionManager{
		nm:           nm,
		maxRetries:   common.DefaultMaxRetries,
		initialDelay: common.DefaultInitialDelay,
	}
}

func (m *ConnectionManager) FindConnectionWithStatus(device nm.Device) (nm.Connection, bool, error) {
	log.Println("Looking for device active connections..")
	activeConn, err := device.GetPropertyActiveConnection()
	if err != nil {
		log.Println("Failed to get active connection: ", err)
		return nil, false, err
	}
	if activeConn == nil {
		log.Println("No active connection found for device")
		conn, err := m.getConnection(device)
		if err != nil {
			return nil, false, err
		}
		return conn, false, nil
	}
	conn, err := activeConn.GetPropertyConnection()
	if err != nil {
		log.Println("Failed to get connection from active connection: ", err)
		return nil, false, err
	}
	state, err := activeConn.GetPropertyState()
	if err != nil {
		log.Println("Failed to get active connection state: ", err)
		return conn, false, err
	}
	if state != nm.NmActiveConnectionStateActivated {
		log.Println("Active connection found but not in activated state, current state: ", state.String())
		return conn, false, nil
	}
	log.Println("Connection with activated state found for device")

	return conn, true, nil
}

func (m *ConnectionManager) getConnection(device nm.Device) (nm.Connection, error) {

	connections, err := device.GetPropertyAvailableConnections()
	if err != nil {
		log.Println("Failed to get available connections: ", err)
		return nil, err
	}

	if len(connections) == 0 {
		log.Println("No available connections found for device")
		return nil, nil
	}

	// Prior to update of ethernet connection, old connection profile is deleted,
	// there should not be more than 1 available connection profile for ethernet devices.
	// For GSM devices, only 1 connection profile should exist as well
	// since old profiles are deleted by the RemoveConnection RPC before creating a new one.
	log.Println("Connection for device retrieved successfully")
	return connections[0], nil
}

func (m *ConnectionManager) GetConnectionSettings(conn nm.Connection) (nm.ConnectionSettings, error) {
	log.Println("Getting connection settings...")

	connSettings, err := conn.GetSettings()
	if err != nil {
		log.Println("Failed to get connection settings: ", err)
		return nil, err
	}

	log.Println("Connection settings retrieved successfully")
	return connSettings, nil
}

func (m *ConnectionManager) UpdateConnection(conn nm.Connection, connSettings nm.ConnectionSettings) error {
	log.Println("Updating connection settings...")

	if err := conn.Update(connSettings); err != nil {
		log.Println("Failed to update connection settings: ", err)
		return err
	}

	log.Println("Connection settings updated successfully")
	return nil
}

func (m *ConnectionManager) AddConnection(connSettings nm.ConnectionSettings) (nm.Connection, error) {
	log.Println("Adding new connection...")

	nmSettings, err := m.newNetworkManagerSettings()
	if err != nil {
		log.Println("Failed to create a NetworkManager settings instance: ", err)
		return nil, err
	}

	conn, err := nmSettings.AddConnection(connSettings)
	if err != nil {
		log.Println("Failed to add connection settings: ", err)
		return nil, err
	}

	log.Println("Connection added successfully")
	return conn, nil
}

func (m *ConnectionManager) newNetworkManagerSettings() (nm.Settings, error) {
	return nm.NewSettings()
}

func (m *ConnectionManager) ActivateConnectionWithRetry(conn nm.Connection, device nm.Device) error {
	return m.retryWithFibonacciBackoff(m.maxRetries, func() error {
		return m.ActivateConnection(conn, device)
	})
}

// with maxRetries attempts and Fibonacci backoff strategy, we actively wait for the connection to be activated
//  instead of relying on the ActivateConnection call to return an error if activation fails.
// with 7 attempts and an initial delay of 1 second, 33 seconds of cumulative waiting time, 
// which is less than the gRPC timeout.
func (m *ConnectionManager) retryWithFibonacciBackoff(attempts int, fn func() error) error {
	prevDelay := 0 * time.Second
	currDelay := m.initialDelay
	var err error

	for i := 1; i <= attempts; i++ {
		if err = fn(); err != nil {
			log.Printf("Retry attempt %d failed: %v. Retrying after %s...", i, err, currDelay)
			time.Sleep(currDelay)
			prevDelay, currDelay = currDelay, currDelay+prevDelay
		} else {
			return nil
		}
	}
	return err
}

func (m *ConnectionManager) ActivateConnection(conn nm.Connection, device nm.Device) error {
	log.Println("Activating connection...")

	// Activating an active connection via gonetworkmanager does not throw an error,
	// so we do not need to deactivate it first.
	activeConn, err := m.nm.ActivateConnection(conn, device, nil)
	if err != nil {
		log.Println("Failed to activate connection: ", err)
		return err
	}

	// Activation takes some seconds, so we actively wait until
	// the connection is activated instead of relying on the ActivateConnection call to return an error if activation fails.
	return m.retryWithFibonacciBackoff(m.maxRetries, func() error {
		if !m.isActivated(activeConn) {
			return errors.New("connection not activated")
		}
		return nil
	})

}

func (m *ConnectionManager) isActivated(activeConn nm.ActiveConnection) bool {
	log.Println("Checking if connection is activated...")

	state, err := activeConn.GetPropertyState()
	if err != nil {
		log.Println("Failed to get active connection state: ", err)
		return false
	}

	// Considering Activated state as successful activation
	if state == nm.NmActiveConnectionStateActivated {
		log.Println("Active connection is in Activated state")
		return true
	}

	log.Println("Active connection is not in Activated state, current state: ", state.String())
	return false
}

func (m *ConnectionManager) DeleteConnection(conn nm.Connection) error {
	log.Println("Deleting connection...")

	if err := conn.Delete(); err != nil {
		log.Println("Failed to delete connection: ", err)
		return err
	}

	log.Println("Connection deleted successfully")
	return nil
}

func (m *ConnectionManager) DeactivateConnection(device nm.Device) error {
	log.Println("Deactivating connection...")

	activeConn, err := device.GetPropertyActiveConnection()
	if err != nil {
		log.Println("Failed to get active connection: ", err)
		return err
	}

	if activeConn == nil {
		log.Println("No active connection found for device, skipping deactivation")
		return nil
	}

	if err := m.nm.DeactivateConnection(activeConn); err != nil {
		log.Println("Failed to deactivate connection: ", err)
		return err
	}

	log.Println("Connection deactivated successfully")
	return nil
}
