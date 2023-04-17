package app

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	v1 "onboardservice/api/siemens_iedge_dmapi_v1"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Testify suite runner detected as race
var testMutex sync.Mutex = sync.Mutex{}

type TestSuite struct {
	suite.Suite
	mockClients *ClientPack
	ml          mockLed
	mn          mockNtp
	mpi         mockedgeCoreRuntime
	mNet        mockNetwork
	mSys        mockSystem
	app         *MainApp
}

func (suite *TestSuite) SetupSuite() {

	os.Remove("/tmp/temp.sock")
	suite.app = CreateServiceApp(&MockClientFactory{})
	suite.app.Clients = suite.mockClients
	suite.app.serverInstance.isContainerUp = true
	initMocks(suite)

	suite.ml.On("ApplyLedAction", mock.Anything,
		mock.Anything).Return(&emptypb.Empty{}, nil)

	suite.ml.On("ResetAll", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil)
}
func (suite *TestSuite) SetupTest() {
}

func initMocks(suite *TestSuite) {
	suite.ml = mockLed{Mock: mock.Mock{}}
	suite.mn = mockNtp{Mock: mock.Mock{}}
	suite.mpi = mockedgeCoreRuntime{Mock: mock.Mock{}}
	suite.mNet = mockNetwork{Mock: mock.Mock{}}
	suite.mSys = mockSystem{Mock: mock.Mock{}}

	suite.ml.On("ApplyLedAction", mock.Anything,
		mock.Anything).Return(&emptypb.Empty{}, nil)
	suite.mockClients = &ClientPack{}
	suite.mockClients.RestClient = &suite.mpi
	suite.mockClients.NtpClient = &suite.mn
	suite.mockClients.SystemClient = &suite.mSys
	suite.mockClients.NetworkClient = &suite.mNet

	suite.app.Clients = suite.mockClients

}

func TestSuiteAll(t *testing.T) {
	suite.Run(t, new(TestSuite))

}

func (suite *TestSuite) TearDownSuite() {
	suite.app.StopGRPC()
	log.Println("Test suite end")
}
func (suite *TestSuite) Test_ApplyAllConfigurations() {
	suite.mn.On("SetNtpServer", mock.Anything, mock.Anything).Return(&emptypb.Empty{},
		nil).Once()
	suite.mNet.On("ApplySettings", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()
	suite.ml.On("ApplyLedAction", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()
	_, err := suite.app.serverInstance.ApplyConfiguration(context.Background(), deviceConfig)
	suite.Nil(err, "apply test failed", err)
}
func (suite *TestSuite) Test_ApplyAllConfigurationsWithNetworkFail() {
	suite.mn.On("SetNtpServer", mock.Anything, mock.Anything).Return(&emptypb.Empty{},
		nil).Once()
	suite.mNet.On("ApplySettings", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, errors.New("virtual error")).Once()
	suite.ml.On("ApplyLedAction", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()
	_, err := suite.app.serverInstance.ApplyConfiguration(context.Background(), deviceConfig)
	suite.NotNil(err, "apply test failed", err)
}
func (suite *TestSuite) Test_ApplyAllConfigurationsWithNtpFail() {

	suite.mn.On("SetNtpServer", mock.Anything, mock.Anything).Return(&emptypb.Empty{},
		errors.New("virtual error")).Once()
	suite.ml.On("ApplyLedAction", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()
	_, err := suite.app.serverInstance.ApplyConfiguration(context.Background(), deviceConfig)
	suite.NotNil(err, "apply test failed", err)
}
func (suite *TestSuite) TestOnboardServer_OnboardWithUSB() {

	suite.mn.On("SetNtpServer", mock.Anything, mock.Anything).Return(&emptypb.Empty{},
		nil).Once()
	suite.mNet.On("ApplySettings", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()
	suite.ml.On("ApplyLedAction", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()

	suite.mpi.On("Activate", mock.Anything).
		Return(true, nil).Once()
	suite.mpi.On("Onboarded").Return(false, nil).Once()
	_, err := suite.app.serverInstance.OnboardWithUSB(context.Background(), deviceConfig)
	suite.Nil(err, "apply test failed", err)

}

func (suite *TestSuite) TestOnboardServer_OnboardWithUSB_FailActivate() {
	suite.mpi.On("Onboarded").Return(false, nil).Once()
	suite.mn.On("SetNtpServer", mock.Anything, mock.Anything).Return(&emptypb.Empty{},
		nil).Once()
	suite.mpi.On("Activate", mock.Anything).
		Return(false, errors.New("virtual error")).Once()
	suite.mNet.On("ApplySettings", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()
	suite.ml.On("ApplyLedAction", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil).Once()

	_, err2 := suite.app.serverInstance.OnboardWithUSB(context.Background(), deviceConfig)
	suite.NotNil(err2, "apply test failed", err2)

}
func (suite *TestSuite) TestOnboardServer_OnboardWithUSB_AlreadyOnboardedCase() {
	suite.mpi.On("Onboarded").Return(true, nil).Once()

	_, err2 := suite.app.serverInstance.OnboardWithUSB(context.Background(), deviceConfig)
	suite.NotNil(err2, "apply test failed", err2)

}

func (suite *TestSuite) TestOnboardServer_SetOnboardingStatus() {

	_, err := suite.app.serverInstance.SetOnboardingStatus(context.Background(), &v1.OnboardingStatus{IsOnboarded: true})
	suite.Nil(err, "apply test failed", err)
}

func (suite *TestSuite) TestOnboardServer_SetOnboardingStatusWithFalse() {
	_, err := suite.app.serverInstance.SetOnboardingStatus(context.Background(), &v1.OnboardingStatus{IsOnboarded: false})

	suite.Nil(err, "apply test failed", err)
}
func (suite *TestSuite) TestOnboardServer_SetOnboardingStatusWithNil() {
	_, err := suite.app.serverInstance.SetOnboardingStatus(context.Background(), nil)

	suite.NotNil(err, "apply test failed", err)
}

func (suite *TestSuite) TestOnboardServer_CreateHelper() {
	factory := ClientFactoryImpl{}
	suite.NotNil(factory.CreateHelper())
}

func (suite *TestSuite) TestOnboardServer_StartApp() {
	//first call should return error
	suite.mpi.On("Onboarded").Return(false, errors.New("virtual error")).Once()
	//second call no error
	suite.mpi.On("Onboarded").Return(false, nil).Once()
	suite.app.StartApp()
}

func (suite *TestSuite) TestOnboardServer_Subscribe() {
	channel := make(chan struct{})
	listener := &MockListener{
		receivedStatus: &v1.OnboardingStatus{
			IsOnboarded: false,
			Message:     "",
		},

		context:     &mockContext{doneChannel: channel},
		doneChannel: channel,
	}

	go func(testSuite *TestSuite) {
		testSuite.app.serverInstance.ListenOnboardState(&emptypb.Empty{}, listener)
	}(suite)

	time.Sleep(200 * time.Millisecond)

	suite.app.serverInstance.SetOnboardingStatus(context.TODO(), &v1.OnboardingStatus{IsOnboarded: true})
	time.Sleep(200 * time.Millisecond)

	channel <- struct{}{}
	time.Sleep(200 * time.Millisecond)

}

func (suite *TestSuite) Test_VerifyArgsForStartGRPC_WithDummySocketForUnix() {
	//Create App to use
	tArgs := []string{"onboardservice", "unix", "/tmpa/temp.sock"}

	tErr := suite.app.StartGRPC(tArgs)

	//Connection failure is expected with dummy sock
	suite.NotNil(tErr, "Did not get expected result. got: %q", tErr)

}
func (suite *TestSuite) Test_VerifyArgsForStartGRPC() {

	//Create App to use
	go func() {
		tArgs := []string{"onboardservice", "unix", "/tmp/temp.sock"}
		tErr := suite.app.StartGRPC(tArgs)

		//Connection failure is expected with dummy sock
		suite.Nil(tErr, "Did not get expected result. got: %q", tErr)
	}()

	time.Sleep(100 * time.Millisecond)

}

func (suite *TestSuite) Test_VerifyArgsForStartGRPC_TCP() {
	//Create App to use
	go func() {
		tArgs := []string{"onboardservice", "tcp", ":4444"}
		tErr := suite.app.StartGRPC(tArgs)

		//Connection failure is expected with dummy sock
		suite.Nil(tErr, "Did not get expected result. got: %q", tErr)
	}()

	time.Sleep(100 * time.Millisecond)

}
func (suite *TestSuite) Test_VerifyArgsForStartGRPCInvalidArgs() {
	//Create App to use

	tArgs := []string{"onboardservice", "unix", "hey", "/tmp/temp.sock"}
	tErr := suite.app.StartGRPC(tArgs)

	//Connection failure is expected with dummy sock
	suite.NotNil(tErr, "Did not get expected result. got: %q", tErr)

}

func (suite *TestSuite) TestOnboardServer_ClientFactory() {
	cf := ClientFactoryImpl{}
	var clients = cf.CreateClients()

	suite.NotNil(clients, "Client factory did not run")
}

var int1 = v1.Interface{
	GatewayInterface: false,
	MacAddress:       "12:13:14:15:15",
	DHCP:             "enabled",
	Static:           nil,
	DNSConfig:        nil,
}

var deviceConfig = &v1.DeviceConfiguration{
	Device: &v1.Device{
		Network:   &v1.NetworkSettings{Interfaces: []*v1.Interface{&int1}},
		NtpServer: []string{"some.server"},
	},
	Onboarding: &v1.Onboarding{PotalUrl: "edgeCoreRuntime.rest.cli"},
	Agents:     nil,
	Proxies: []*v1.ProxySettings{&v1.ProxySettings{
		ProxyId:            "",
		ProxyType:          "",
		Host:               "",
		Protocol:           "",
		AuthenticationType: "",
		User:               "",
		Password:           "",
		NoProxy:            "",
		CustomPorts:        nil,
	}},
}

func (m *MockClientFactory) CreateClients() *ClientPack {
	return &ClientPack{
		NetworkClient: &mockNetwork{},
		NtpClient:     &mockNtp{},
		RestClient:    &mockedgeCoreRuntime{},
	}
}
