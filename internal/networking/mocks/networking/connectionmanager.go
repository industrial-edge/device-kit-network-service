/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	nm "github.com/Wifx/gonetworkmanager/v2"
	"github.com/stretchr/testify/mock"
)

type MockConnectionManager struct {
	mock.Mock
}

func (m *MockConnectionManager) FindConnectionWithStatus(device nm.Device) (nm.Connection, bool, error) {
	args := m.Called(device)
	mockConn := args.Get(0)
	if mockConn == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return mockConn.(nm.Connection), args.Bool(1), args.Error(2)
}

func (m *MockConnectionManager) GetConnectionSettings(conn nm.Connection) (nm.ConnectionSettings, error) {
	args := m.Called(conn)
	mockConnSettings := args.Get(0)
	if mockConnSettings == nil {
		return nil, args.Error(1)
	}
	return mockConnSettings.(nm.ConnectionSettings), args.Error(1)
}

func (m *MockConnectionManager) UpdateConnection(conn nm.Connection, settings nm.ConnectionSettings) error {
	args := m.Called(conn, settings)
	return args.Error(0)
}

func (m *MockConnectionManager) AddConnection(settings nm.ConnectionSettings) (nm.Connection, error) {
	args := m.Called(settings)
	conn := args.Get(0)
	if conn == nil {
		return nil, args.Error(1)
	}
	return conn.(nm.Connection), args.Error(1)
}

func (m *MockConnectionManager) DeleteConnection(conn nm.Connection) error {
	args := m.Called(conn)
	return args.Error(0)
}

func (m *MockConnectionManager) ActivateConnection(conn nm.Connection, device nm.Device) error {
	args := m.Called(conn, device)
	return args.Error(0)
}

func (m *MockConnectionManager) ActivateConnectionWithRetry(conn nm.Connection, device nm.Device) error {
	args := m.Called(conn, device)
	return args.Error(0)
}

func (m *MockConnectionManager) DeactivateConnection(device nm.Device) error {
	args := m.Called(device)
	return args.Error(0)
}
