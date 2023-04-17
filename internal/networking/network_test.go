package networking

import (
	"fmt"
	"github.com/Wifx/gonetworkmanager"
	"log"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
)

var mutex sync.Mutex

func Test_Conversions(t *testing.T) {

	localhost := "127.0.0.1"
	litteIP := IPToUInt32LI("127.0.0.1")
	if localhost != IPFromUInt32LI(litteIP) {
		t.Fail()
	}

	fmt.Println(IPToUInt32BI("127.0.0.1"))
	IPToUInt32BI("255.255.255.0")

	sz := ParseNetMaskSize("255.255.255.0")
	fmt.Println(sz)

	if sz != 24 {
		t.Fail()
	}

}

// for testing on real device please enter your ethernet typed interface's MAC address.
const yourMac = "00:0C:29:C2:51:81"

// for testing on real device please enter a label (e.g "X1" "testLabel")
const yourNetworkDeviceLabelName = "TestLabel"

// for testing on real device please enter your ethernet typed interface's Interface name. (e.g "enp0s3" "ens5")
const yourNetworkDeviceInterfaceName = "eth1"

func Test_GetAllInterfaces(t *testing.T) {
	ParseNetMask(24)

	val := "8.8.8.8"
	intForm := IPToUInt32LI(val)
	stringForm := IPFromUInt32BI(intForm)

	if val != stringForm {
		log.Fatal("Test failed")

	}

	sut := NewNetworkConfigurator()
	list := sut.GetEthernetInterfaces()
	log.Println(list)
	if list == nil {
		log.Fatal("Test failed")
	}

}

func Test_UnknownMac(t *testing.T) {
	sut := NewNetworkConfigurator()
	if sut.getDeviceWithMac("sdfasdfasdfasdf") != nil {
		log.Println("Test failed")
	}
}

func Test_ApplyNewNetworkSettingsManualWithWrongIP(t *testing.T) {

	mutex.Lock()

	testData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       yourMac,
		DHCP:             "disabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "392.168.1.270",
			NetMask: "999.255.255.0",
			Gateway: "192.1685.1.1",
		},
		DNSConfig: &v1.Interface_Dns{PrimaryDNS: "8.8.8.8.7", SecondaryDNS: "888.4.4.4"},
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	sut := NewNetworkConfigurator()
	ok, errs := sut.ArePreconditionsOk(newSettings)
	defer mutex.Unlock()

	if ok || errs == nil {
		t.Fail()
	}
	log.Println(errs)

}

func Test_ApplyNewNetworkSettingsManualNilDns(t *testing.T) {

	mutex.Lock()

	testData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       yourMac,
		DHCP:             "disabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.1.27",
			NetMask: "255.255.255.0",
			Gateway: "192.168.1.1",
		},
		L2Conf: &v1.Interface_L2{
			StartingAddressIPv4: "192.168.1.27",
			NetMask:             "255.255.255.0",
			Range:               "8",
			Gateway:             "192.168.1.27",
			AuxiliaryAddresses:  map[string]string{"aux1": "192.168.1.28"},
		},
		InterfaceName: "enp2s0",
		Label:         "enp0s8",
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	sut := NewNetworkConfigurator()
	ok, errs := sut.ArePreconditionsOk(newSettings)

	if ok == false || errs != nil {
		defer mutex.Unlock()
		t.Fail()
	}
	log.Println(errs)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()

	newOne := sut.GetInterfaceWithMac(testData.MacAddress)

	log.Println("Read new configuration from system :")
	log.Println("GOT: ", newOne)
	log.Println("WANT: ", testData)

	log.Println(newOne)

	if !proto.Equal(newOne.Static, testData.Static) {
		t.Fail()
	}
	if newOne == nil {
		t.Fail()
	}

	if err != nil {
		t.Fail()
	}
}

func Test_ApplyNewNetworkSettingsManualEmptyDns(t *testing.T) {

	mutex.Lock()
	testData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       yourMac,
		DHCP:             "disabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.1.13",
			NetMask: "255.255.255.0",
			Gateway: "192.168.1.1",
		},
		DNSConfig: &v1.Interface_Dns{},
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	sut := NewNetworkConfigurator()
	ok, errs := sut.ArePreconditionsOk(newSettings)

	if ok == false || errs != nil {
		defer mutex.Unlock()
		t.Fail()
	}
	log.Println(errs)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()
	newOne := sut.GetInterfaceWithMac(testData.MacAddress)

	log.Println("Read new configuration from system :")
	log.Println("GOT: ", newOne)
	log.Println("WANT: ", testData)

	if !proto.Equal(newOne.Static, testData.Static) {
		t.Fail()
	}

	if newOne == nil {
		t.Fail()
	}

	if err != nil {
		t.Fail()
	}
}

func Test_ApplyNewNetworkSettingsManual(t *testing.T) {

	mutex.Lock()

	testData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       yourMac,
		DHCP:             "disabled",
		Static: &v1.Interface_StaticConf{
			IPv4:    "192.168.1.13",
			NetMask: "255.255.255.0",
			Gateway: "192.168.1.1",
		},
		DNSConfig: &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "4.4.4.4"},
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	sut := NewNetworkConfigurator()
	ok, errs := sut.ArePreconditionsOk(newSettings)

	if ok == false || errs != nil {
		defer mutex.Unlock()
		t.Fail()
	}
	log.Println(errs)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()
	newOne := sut.GetInterfaceWithMac(testData.MacAddress)

	log.Println("Read new configuration from system :")
	log.Println("GOT: ", newOne)
	log.Println("WANT: ", testData)

	log.Println(newOne)
	if !proto.Equal(newOne.Static, testData.Static) {
		t.Fail()
	}

	if !proto.Equal(newOne.GetDNSConfig(), testData.GetDNSConfig()) {
		t.Logf("Expected DNS etry: %s, got: %s", testData.GetDNSConfig(), newOne.GetDNSConfig())
		t.Fail()
	}

	if newOne == nil {
		t.Fail()
	}

	if err != nil {
		t.Fail()
	}
}

func Test_ApplyNewNetworkSettingsAutoOnlyMac(t *testing.T) {
	mutex.Lock()

	testData := &v1.Interface{
		GatewayInterface: true,
		MacAddress:       yourMac,
		DHCP:             "enabled",
		Static:           &v1.Interface_StaticConf{},
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "4.4.4.4"},
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	goNm, _ := gonetworkmanager.NewNetworkManager()
	sut := NewNetworkConfiguratorWithNM(goNm)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()
	if err != nil {
		t.Fail()
	}
	newOne := sut.GetInterfaceWithMac(testData.MacAddress)
	log.Println("Read new configuration from system :")
	log.Println(newOne)

	if !cmp.Equal(newOne.MacAddress, testData.MacAddress) {
		t.Logf("Expected Mac: %s , got: %s", testData.MacAddress, newOne.MacAddress)
		t.Fail()
	}

	if !proto.Equal(newOne.GetDNSConfig(), testData.GetDNSConfig()) {
		t.Logf("Expected DNS etry: %s, got: %s", testData.GetDNSConfig(), newOne.GetDNSConfig())
		t.Fail()
	}

	if !cmp.Equal(newOne.DHCP, testData.DHCP) {
		t.Logf("expected DHCP: %s , got:  %s", testData.DHCP, newOne.DHCP)
		t.Fail()
	}

}

func Test_ApplyNewNetworkSettingsAutoWithLabel(t *testing.T) {
	mutex.Lock()

	testData := &v1.Interface{
		GatewayInterface: true,
		DHCP:             "enabled",
		Static:           &v1.Interface_StaticConf{},
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "4.4.4.4"},
		Label:            yourNetworkDeviceLabelName,
		InterfaceName:    "",
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	// Prepare label to interface map file.
	labelMap := make(map[string]string)
	labelMap[yourNetworkDeviceLabelName] = yourNetworkDeviceInterfaceName
	WriteMapToFile(labelMap, LabelMapFileName)

	gonm, _ := gonetworkmanager.NewNetworkManager()
	sut := NewNetworkConfiguratorWithNM(gonm)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()
	if err != nil {
		t.Fail()
	}
	newOne := sut.GetInterfaceWithLabel(testData.Label)
	log.Println("Read new configuration from system :")
	log.Println(newOne)

	if !cmp.Equal(strings.ToUpper(newOne.InterfaceName), strings.ToUpper(yourNetworkDeviceInterfaceName)) {
		t.Logf("Expected Label: %s , got: %s", yourNetworkDeviceInterfaceName, newOne.InterfaceName)
		t.Fail()
	}

	if !cmp.Equal(newOne.DHCP, testData.DHCP) {
		t.Logf("Expected Label: %s , got: %s", testData.DHCP, newOne.DHCP)
		t.Fail()
	}

}

func Test_ApplyNetworkSettingsWithLabel(t *testing.T) {
	mutex.Lock()

	goNm, _ := gonetworkmanager.NewNetworkManager()
	sut := NewNetworkConfiguratorWithNM(goNm)

	//testData := sut.GetInterfaceWithLabel("enp0s8")
	testData := &v1.Interface{
		GatewayInterface: true,
		DHCP:             "enabled",
		Static:           &v1.Interface_StaticConf{},
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "4.4.4.4"},
		Label:            yourNetworkDeviceLabelName,
		InterfaceName:    yourNetworkDeviceInterfaceName,
	}

	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	// Prepare label to interface map file.
	labelMap := make(map[string]string)
	labelMap[yourNetworkDeviceLabelName] = yourNetworkDeviceInterfaceName
	WriteMapToFile(labelMap, LabelMapFileName)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()
	if err != nil {
		t.Fail()
	}
	newOne := sut.GetInterfaceWithLabel(testData.Label)
	log.Println("Read new configuration from system :")
	log.Println("GetInterfaceWithLabel RPC result : ", newOne)

	if !cmp.Equal(newOne.Label, strings.ToUpper(testData.Label)) {
		t.Logf("Expected Label: %s , got: %s", testData.Label, newOne.Label)
		t.Fail()
	}

	if !cmp.Equal(strings.ToUpper(newOne.InterfaceName), strings.ToUpper(testData.InterfaceName)) {
		t.Logf("Expected Label: %s , got: %s", testData.InterfaceName, newOne.InterfaceName)
		t.Fail()
	}

	if !cmp.Equal(newOne.DHCP, testData.DHCP) {
		t.Logf("Expected Label: %s , got: %s", testData.DHCP, newOne.DHCP)
		t.Fail()
	}

}

func Test_ApplyNewNetworkSettingsAutoWithoutMacAndInterfaceName(t *testing.T) {
	mutex.Lock()

	testData := &v1.Interface{
		GatewayInterface: true,
		DHCP:             "enabled",
		Static:           &v1.Interface_StaticConf{},
		DNSConfig:        &v1.Interface_Dns{PrimaryDNS: "8.8.8.8", SecondaryDNS: "4.4.4.4"},
	}
	newSettings := &v1.NetworkSettings{Interfaces: []*v1.Interface{testData}}

	goNm, _ := gonetworkmanager.NewNetworkManager()
	sut := NewNetworkConfiguratorWithNM(goNm)

	err := sut.Apply(newSettings)
	time.Sleep(5 * time.Second)
	defer mutex.Unlock()
	if err == nil {
		t.Fail()
	}
}
