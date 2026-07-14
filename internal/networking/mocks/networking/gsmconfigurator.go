/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"github.com/stretchr/testify/mock"
)

type MockGSMConfigurator struct {
	mock.Mock
}

func (m *MockGSMConfigurator) Configure() error {
	args := m.Called()
	return args.Error(0)
}
