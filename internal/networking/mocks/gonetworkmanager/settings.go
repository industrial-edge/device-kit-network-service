/*
 * Copyright © Siemens 2024 - 2025. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gonetworkmanager

import (
	. "github.com/Wifx/gonetworkmanager/v2"
	"github.com/stretchr/testify/mock"
)

type MockSettings struct {
	mock.Mock
}

func (m *MockSettings) ListConnections() ([]Connection, error) {
	args := m.Called()
	return args.Get(0).([]Connection), args.Error(1)
}

func (m *MockSettings) ReloadConnections() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSettings) GetConnectionByUUID(uuid string) (Connection, error) {
	args := m.Called(uuid)
	return args.Get(0).(Connection), args.Error(1)
}

func (m *MockSettings) AddConnection(settings ConnectionSettings) (Connection, error) {
	args := m.Called(settings)
	return args.Get(0).(Connection), args.Error(1)
}

func (m *MockSettings) AddConnectionUnsaved(settings ConnectionSettings) (Connection, error) {
	args := m.Called(settings)
	return args.Get(0).(Connection), args.Error(1)
}

func (m *MockSettings) SaveHostname(hostname string) error {
	args := m.Called(hostname)
	return args.Error(0)
}

func (m *MockSettings) GetPropertyCanModify() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockSettings) GetPropertyHostname() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
