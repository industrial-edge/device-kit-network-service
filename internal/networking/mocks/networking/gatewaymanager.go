/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"github.com/stretchr/testify/mock"
)

type MockGatewayManager struct {
	mock.Mock
}

func (m *MockGatewayManager) Reset(interfaceName string) error {
	args := m.Called(interfaceName)
	return args.Error(0)
}
