package app

import (
	"context"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	v1 "onboardservice/api/siemens_iedge_dmapi_v1"
	"onboardservice/internal/system"
	"time"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockClientFactory struct {
}

func (m *MockClientFactory) CreateHelper() system.Configurable {
	return &system.MockHelper{}
}

type MockListener struct {
	receivedStatus *v1.OnboardingStatus
	context        context.Context
	doneChannel    chan struct{}
}

func (m *MockListener) Send(onboardingStatus *v1.OnboardingStatus) error {
	log.Println("MOCK RECEVIED STATUS :", onboardingStatus.IsOnboarded)
	m.receivedStatus = onboardingStatus
	return nil
}

func (m *MockListener) SetHeader(md metadata.MD) error {
	return nil
}

func (m *MockListener) SendHeader(md metadata.MD) error {
	return nil
}

func (m *MockListener) SetTrailer(md metadata.MD) {

}

type mockContext struct {
	doneChannel chan struct{}
}

func (m *mockContext) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (m *mockContext) Done() <-chan struct{} {
	return m.doneChannel
}

func (m *mockContext) Err() error {
	panic("implement me")
}

func (m *mockContext) Value(key interface{}) interface{} {
	panic("implement me")
}

func (m *MockListener) Context() context.Context {
	return m.context
}

func (m *MockListener) SendMsg(a interface{}) error {
	return nil
}

func (m *MockListener) RecvMsg(a interface{}) error {
	return nil
}

type mockNetwork struct {
	mock.Mock
}

type mockNtp struct {
	mock.Mock
}

func (m *mockNtp) SetNtpServer(ctx context.Context, in *v1.Ntp, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(1)

}

func (m *mockNtp) GetNtpServer(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.Ntp, error) {
	args := m.Called(in)
	return &v1.Ntp{}, args.Error(1)

}

func (m *mockNtp) GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.Status, error) {
	args := m.Called(in)
	return &v1.Status{}, args.Error(1)

}

type mockSystem struct {
	mock.Mock
}

func (m mockSystem) HardReset(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("implement me")
}

func (m mockSystem) GetCustomSettings(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*anypb.Any, error) {
	panic("implement me")
}

func (m mockSystem) ApplyCustomSettings(ctx context.Context, in *anypb.Any, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("implement me")
}

func (m mockSystem) GetLogFile(ctx context.Context, in *v1.LogRequest, opts ...grpc.CallOption) (*v1.LogResponse, error) {
	panic("implement me")
}

func (m mockSystem) RestartDevice(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(1)
}

func (m mockSystem) ShutdownDevice(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(1)
}

func (m mockSystem) FactoryReset(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(in)
	return &emptypb.Empty{}, args.Error(1)
}

func (m mockSystem) GetModelNumber(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.ModelNumber, error) {
	args := m.Called(in)
	return &v1.ModelNumber{
		ModelNumber: "ipc227e",
	}, args.Error(1)
}

func (m mockSystem) GetFirmwareInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.FirmwareInfo, error) {
	args := m.Called(in)
	return &v1.FirmwareInfo{Version: "v1.0.0"}, args.Error(1)
}

func (m mockSystem) GetResourceStats(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.Stats, error) {
	args := m.Called(in)
	return &v1.Stats{}, args.Error(1)
}

func (m mockSystem) GetLimits(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.Limits, error) {
	args := m.Called(in)
	return &v1.Limits{}, args.Error(1)
}

type mockLed struct {
	mock.Mock
}

func (m *mockLed) ResetAll(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return nil, args.Error(1)
}

type mockedgeCoreRuntime struct {
	mock.Mock
	isGood bool
}

func (m *mockedgeCoreRuntime) Activate(configuration *v1.DeviceConfiguration) (bool, error) {
	args := m.Called(configuration)
	return args.Bool(0), args.Error(1)
}

func (m *mockedgeCoreRuntime) Onboarded() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *mockNetwork) GetAllInterfaces(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*v1.NetworkSettings, error) {
	panic("implement me")
	args := m.Called(ctx, in)
	return &v1.NetworkSettings{}, args.Error(1)
}

func (m *mockNetwork) GetInterfaceWithMac(ctx context.Context, in *v1.NetworkInterfaceRequest, opts ...grpc.CallOption) (*v1.Interface, error) {
	panic("implement me")
}

func (m *mockNetwork) GetInterfaceWithLabel(ctx context.Context, in *v1.NetworkInterfaceRequestWithLabel, opts ...grpc.CallOption) (*v1.Interface, error) {
	panic("implement me")
}

func (m *mockNetwork) ApplySettings(ctx context.Context, in *v1.NetworkSettings, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return &emptypb.Empty{}, args.Error(1)
}
