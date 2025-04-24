package gonetworkmanager

import (
	. "github.com/Wifx/gonetworkmanager/v2"
	"github.com/stretchr/testify/mock"
)

type MockIP4Config struct {
	mock.Mock
}

func (m *MockIP4Config) GetPropertyAddresses() ([]IP4Address, error) {
	args := m.Called()
	return args.Get(0).([]IP4Address), args.Error(1)
}

func (m *MockIP4Config) GetPropertyAddressData() ([]IP4AddressData, error) {
	args := m.Called()
	return args.Get(0).([]IP4AddressData), args.Error(1)
}

func (m *MockIP4Config) GetPropertyGateway() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockIP4Config) GetPropertyRoutes() ([]IP4Route, error) {
	args := m.Called()
	return args.Get(0).([]IP4Route), args.Error(1)
}

func (m *MockIP4Config) GetPropertyRouteData() ([]IP4RouteData, error) {
	args := m.Called()
	return args.Get(0).([]IP4RouteData), args.Error(1)
}

func (m *MockIP4Config) GetPropertyNameservers() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIP4Config) GetPropertyNameserverData() ([]IP4NameserverData, error) {
	args := m.Called()
	return args.Get(0).([]IP4NameserverData), args.Error(1)
}

func (m *MockIP4Config) GetPropertyDomains() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIP4Config) GetPropertySearches() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIP4Config) GetPropertyDnsOptions() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIP4Config) GetPropertyDnsPriority() (uint32, error) {
	args := m.Called()
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockIP4Config) GetPropertyWinsServerData() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockIP4Config) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
