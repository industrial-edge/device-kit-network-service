/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package mocks

import (
	"os"

	"github.com/stretchr/testify/mock"
)

// MockFileSystem is a mock implementation of the FileSystem interface.
type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	args := m.Called(path, data, perm)
	return args.Error(0)
}

func (m *MockFileSystem) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystem) Chmod(name string, mode os.FileMode) error {
	args := m.Called(name, mode)
	return args.Error(0)
}

func (m *MockFileSystem) Chown(name string, uid, gid int) error {
	args := m.Called(name, uid, gid)
	return args.Error(0)
}

func (m *MockFileSystem) Stat(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}
