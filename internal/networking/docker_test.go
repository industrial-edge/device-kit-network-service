package networking

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

var mockedExitStatus = 0
var mockedStdout = []byte(`[{
        "Name": "zzz_layer2_net1",
        "Id": "ea08c1af38e37b81912d8d1b583984cd60e9cbe1f8dd71eb60532e7055b2dfe5",
        "Created": "2021-03-08T06:55:45.721369455Z",
        "Scope": "local",
        "Driver": "macvlan",
        "EnableIPv6": false,
        "IPAM": {
            "Driver": "default",
            "Options": {},
            "Config": [
                {
                    "Subnet": "192.168.0.0/16",
                    "IPRange": "192.168.18.24/29",
                    "Gateway": "192.168.18.24",
                    "AuxiliaryAddresses": {
                        "auxAddress0": "192.168.18.26",
                        "auxAddress1": "192.168.18.27"
                    }
                }
            ]
        },
        "Internal": false,
        "Attachable": false,
        "Ingress": false,
        "ConfigFrom": {
            "Network": ""
        },
        "ConfigOnly": false,
        "Containers": {},
        "Options": {
            "com.docker.network.bridge.enable_icc": "true",
            "com.docker.network.bridge.enable_ip_masquerade": "false",
            "parent": "ens18"
        },
        "Labels": {}
    }]`)
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	fmt.Println("Fake EXEC Called")
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + string(mockedStdout),
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func Test_DockerNetworkGetMacvlanConnection(t *testing.T) {

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	outL2 := dockerNetworkGetMacvlanConnection("ens18")

	if outL2.Gateway != "192.168.18.24" {
		t.Errorf("Expected %s, got %s","192.168.18.24"  , outL2.Gateway)
	}
}


func Test_DockerNetworkGetMacvlanConnectionWithDummyInterfaceName(t *testing.T) {

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	outL2 := dockerNetworkGetMacvlanConnection("DummyInterfaceName")

	if outL2.Gateway != "" {
		t.Errorf("Expected %s, got %s",""  , outL2.Gateway)
	}
}