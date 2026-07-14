/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"github.com/stretchr/testify/mock"
)

type MockInterfaceStateManager struct {
	mock.Mock
}

func (m *MockInterfaceStateManager) Backup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockInterfaceStateManager) Restore() error {
	args := m.Called()
	return args.Error(0)
}
