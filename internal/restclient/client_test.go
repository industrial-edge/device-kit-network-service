package restclient

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	v1 "onboardservice/api/siemens_iedge_dmapi_v1"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const PORT = "3131"

type TestSuite struct {
	suite.Suite
	server *mockServer
}

func (suite *TestSuite) SetupSuite() {
	log.Println("suite starting")
	suite.server = &mockServer{}
	go func() { suite.server.Start() }()
	time.Sleep(1 * time.Second)
	log.Println("suite started")
}

func TestSuiteAll(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TearDownSuite() {
	log.Println("Closing Web Server")

	err := suite.server.Stop()
	if err != nil {
		log.Println("Closed web server")
	}

	log.Println("Test end")
}

func (suite *TestSuite) Test_TryActivate() {

	int1 := v1.Interface{
		GatewayInterface: false,
		MacAddress:       "12:13:14:15:15",
		DHCP:             "disabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.41.41",
			NetMask: "255.255.255.0",
			Gateway: "",
		},
		DNSConfig: nil,
	}

	deviceConfig := &v1.DeviceConfiguration{
		Device: &v1.Device{
			Network:   &v1.NetworkSettings{Interfaces: []*v1.Interface{&int1}},
			NtpServer: nil,
		},
		Onboarding: nil,
		Agents: []*v1.Agent{{
			Name:    "Smith",
			Proxy:   []string{"some proxy id"},
			AgentId: "47",
			Security: &v1.Security{
				BaseUrl:                 "secure by default",
				ClientCredentialProfile: []string{"#StayAtHome"},
				CaChain:                 "This product has been created during covid19 pandemic in April2020",
			}},
		},
		Proxies: nil,
	}

	client := NewClient("http://localhost:" + PORT)
	val, err := client.Activate(deviceConfig)
	suite.True(val)
	suite.Nil(err)

}

func (suite *TestSuite) Test_CheckOnboardedWithMock() {
	client := NewClient("http://localhost:" + PORT)
	val, err := client.Onboarded()
	suite.True(val)
	suite.Nil(err)
}
func (suite *TestSuite) Test_OnboardedWihtFailingRoute() {
	client := NewClient("http://localhost:" + PORT + "/123123")
	val, err := client.Onboarded()
	suite.False(val)
	suite.NotNil(err)
}
func (suite *TestSuite) Test_OnboardedWithInvalidServer() {
	client := NewClient("http://localhost:31311/INVALIDADRESS")
	val, err := client.Onboarded()
	suite.False(val)
	suite.NotNil(err)
}

type mockServer struct {
	server *http.Server
}

func (m *mockServer) Start() {
	s := &http.Server{
		Addr:         "127.0.0.1:" + PORT,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	m.server = s
	http.HandleFunc("/v1/activate", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			fmt.Fprintf(w, "Hello Request was, %q", string(body))
		}

	})
	http.HandleFunc("/v1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			fmt.Fprintf(w, "Hello Request was, %q", string(body))
		}

	})
	fmt.Printf("Starting server for testing HTTP POST...\n")
	s.ListenAndServe()
}
func (m *mockServer) Stop() error {
	return m.server.Close()
}
