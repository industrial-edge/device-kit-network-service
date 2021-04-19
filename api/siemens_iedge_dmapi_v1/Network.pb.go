// Industrial Edge Device Kit API
// Copyright Siemens AG. 2020, All rights reserved.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.0
// source: Network.proto

package siemens_iedge_dmapi_v1

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Contains MAC address, used for retrieving specified Network Interface settings.
type NetworkInterfaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mac string `protobuf:"bytes,1,opt,name=mac,proto3" json:"mac,omitempty"`
}

func (x *NetworkInterfaceRequest) Reset() {
	*x = NetworkInterfaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Network_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkInterfaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkInterfaceRequest) ProtoMessage() {}

func (x *NetworkInterfaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_Network_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkInterfaceRequest.ProtoReflect.Descriptor instead.
func (*NetworkInterfaceRequest) Descriptor() ([]byte, []int) {
	return file_Network_proto_rawDescGZIP(), []int{0}
}

func (x *NetworkInterfaceRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

// Interface type holds settings for a Network Interface.
type Interface struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GatewayInterface bool                  `protobuf:"varint,1,opt,name=GatewayInterface,proto3" json:"GatewayInterface,omitempty"` // if true, route metric will be set to 1. Otherwise route metric is -1. Similarly, when the interface is requested,return value will be true if route metric is 1.
	MacAddress       string                `protobuf:"bytes,2,opt,name=MacAddress,proto3" json:"MacAddress,omitempty"`              // "20:87:56:b5:ed:e0"
	DHCP             string                `protobuf:"bytes,3,opt,name=DHCP,proto3" json:"DHCP,omitempty"`                          // values can be 'enabled' or 'disabled'. for compatiblity reasons it is not boolean.
	Static           *Interface_StaticConf `protobuf:"bytes,4,opt,name=Static,proto3" json:"Static,omitempty"`                      // Static field is StaticConf type instance.
	DNSConfig        *Interface_Dns        `protobuf:"bytes,5,opt,name=DNSConfig,proto3" json:"DNSConfig,omitempty"`                // DNSConfig is dns type instance.
	L2Conf           *Interface_L2         `protobuf:"bytes,6,opt,name=L2Conf,proto3" json:"L2Conf,omitempty"`
	InterfaceName    string                `protobuf:"bytes,7,opt,name=InterfaceName,proto3" json:"InterfaceName,omitempty"` // ens2p - read only.
}

func (x *Interface) Reset() {
	*x = Interface{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Network_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Interface) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interface) ProtoMessage() {}

func (x *Interface) ProtoReflect() protoreflect.Message {
	mi := &file_Network_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interface.ProtoReflect.Descriptor instead.
func (*Interface) Descriptor() ([]byte, []int) {
	return file_Network_proto_rawDescGZIP(), []int{1}
}

func (x *Interface) GetGatewayInterface() bool {
	if x != nil {
		return x.GatewayInterface
	}
	return false
}

func (x *Interface) GetMacAddress() string {
	if x != nil {
		return x.MacAddress
	}
	return ""
}

func (x *Interface) GetDHCP() string {
	if x != nil {
		return x.DHCP
	}
	return ""
}

func (x *Interface) GetStatic() *Interface_StaticConf {
	if x != nil {
		return x.Static
	}
	return nil
}

func (x *Interface) GetDNSConfig() *Interface_Dns {
	if x != nil {
		return x.DNSConfig
	}
	return nil
}

func (x *Interface) GetL2Conf() *Interface_L2 {
	if x != nil {
		return x.L2Conf
	}
	return nil
}

func (x *Interface) GetInterfaceName() string {
	if x != nil {
		return x.InterfaceName
	}
	return ""
}

// Contains multiple network interface settings. It can be used to apply or get the settings.
type NetworkSettings struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Interfaces []*Interface `protobuf:"bytes,1,rep,name=Interfaces,proto3" json:"Interfaces,omitempty"` // Network settings contains an array of Interfaces.Applying new settings or receiving current settings is supported for multiple ethernet typed network interfaces supported.On
}

func (x *NetworkSettings) Reset() {
	*x = NetworkSettings{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Network_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkSettings) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkSettings) ProtoMessage() {}

func (x *NetworkSettings) ProtoReflect() protoreflect.Message {
	mi := &file_Network_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkSettings.ProtoReflect.Descriptor instead.
func (*NetworkSettings) Descriptor() ([]byte, []int) {
	return file_Network_proto_rawDescGZIP(), []int{2}
}

func (x *NetworkSettings) GetInterfaces() []*Interface {
	if x != nil {
		return x.Interfaces
	}
	return nil
}

// StaticConf type holds IP Netmask and Gateway information
type Interface_StaticConf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IPv4    string `protobuf:"bytes,1,opt,name=IPv4,proto3" json:"IPv4,omitempty"`       // e.g: 192.168.0.2
	NetMask string `protobuf:"bytes,2,opt,name=NetMask,proto3" json:"NetMask,omitempty"` // e.g: 255.255.255.0
	Gateway string `protobuf:"bytes,3,opt,name=Gateway,proto3" json:"Gateway,omitempty"` // e.g: 192.168.0.1
}

func (x *Interface_StaticConf) Reset() {
	*x = Interface_StaticConf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Network_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Interface_StaticConf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interface_StaticConf) ProtoMessage() {}

func (x *Interface_StaticConf) ProtoReflect() protoreflect.Message {
	mi := &file_Network_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interface_StaticConf.ProtoReflect.Descriptor instead.
func (*Interface_StaticConf) Descriptor() ([]byte, []int) {
	return file_Network_proto_rawDescGZIP(), []int{1, 0}
}

func (x *Interface_StaticConf) GetIPv4() string {
	if x != nil {
		return x.IPv4
	}
	return ""
}

func (x *Interface_StaticConf) GetNetMask() string {
	if x != nil {
		return x.NetMask
	}
	return ""
}

func (x *Interface_StaticConf) GetGateway() string {
	if x != nil {
		return x.Gateway
	}
	return ""
}

// Type that contains Primary and Secondary DNS.
type Interface_Dns struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrimaryDNS   string `protobuf:"bytes,1,opt,name=PrimaryDNS,proto3" json:"PrimaryDNS,omitempty"`     // e.g:  "1.1.1.2"
	SecondaryDNS string `protobuf:"bytes,2,opt,name=SecondaryDNS,proto3" json:"SecondaryDNS,omitempty"` // e.g: "1.1.1.1"
}

func (x *Interface_Dns) Reset() {
	*x = Interface_Dns{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Network_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Interface_Dns) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interface_Dns) ProtoMessage() {}

func (x *Interface_Dns) ProtoReflect() protoreflect.Message {
	mi := &file_Network_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interface_Dns.ProtoReflect.Descriptor instead.
func (*Interface_Dns) Descriptor() ([]byte, []int) {
	return file_Network_proto_rawDescGZIP(), []int{1, 1}
}

func (x *Interface_Dns) GetPrimaryDNS() string {
	if x != nil {
		return x.PrimaryDNS
	}
	return ""
}

func (x *Interface_Dns) GetSecondaryDNS() string {
	if x != nil {
		return x.SecondaryDNS
	}
	return ""
}

type Interface_L2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StartingAddressIPv4 string            `protobuf:"bytes,1,opt,name=StartingAddressIPv4,proto3" json:"StartingAddressIPv4,omitempty"`                                                                                       // e.g: 192.168.0.2
	NetMask             string            `protobuf:"bytes,2,opt,name=NetMask,proto3" json:"NetMask,omitempty"`                                                                                                               // e.g: 255.255.255.0
	Range               string            `protobuf:"bytes,3,opt,name=Range,proto3" json:"Range,omitempty"`                                                                                                                   // e.g: 16
	Gateway             string            `protobuf:"bytes,4,opt,name=Gateway,proto3" json:"Gateway,omitempty"`                                                                                                               //e.g: 192.168.2.1
	AuxiliaryAddresses  map[string]string `protobuf:"bytes,5,rep,name=AuxiliaryAddresses,proto3" json:"AuxiliaryAddresses,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` //Preserved addresses for other devices. These addresses won't be assigned to containers. e.g  my_plc,  192.168.0.5
}

func (x *Interface_L2) Reset() {
	*x = Interface_L2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Network_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Interface_L2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interface_L2) ProtoMessage() {}

func (x *Interface_L2) ProtoReflect() protoreflect.Message {
	mi := &file_Network_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interface_L2.ProtoReflect.Descriptor instead.
func (*Interface_L2) Descriptor() ([]byte, []int) {
	return file_Network_proto_rawDescGZIP(), []int{1, 2}
}

func (x *Interface_L2) GetStartingAddressIPv4() string {
	if x != nil {
		return x.StartingAddressIPv4
	}
	return ""
}

func (x *Interface_L2) GetNetMask() string {
	if x != nil {
		return x.NetMask
	}
	return ""
}

func (x *Interface_L2) GetRange() string {
	if x != nil {
		return x.Range
	}
	return ""
}

func (x *Interface_L2) GetGateway() string {
	if x != nil {
		return x.Gateway
	}
	return ""
}

func (x *Interface_L2) GetAuxiliaryAddresses() map[string]string {
	if x != nil {
		return x.AuxiliaryAddresses
	}
	return nil
}

var File_Network_proto protoreflect.FileDescriptor

var file_Network_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x1e, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64,
	0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x1a,
	0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2b, 0x0a, 0x17,
	0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x22, 0xd3, 0x06, 0x0a, 0x09, 0x49, 0x6e,
	0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x47, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x10, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x4d, 0x61, 0x63, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x4d, 0x61, 0x63, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x44, 0x48, 0x43, 0x50, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x44, 0x48, 0x43, 0x50, 0x12, 0x4c, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x69,
	0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e,
	0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61,
	0x63, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x69, 0x63, 0x43, 0x6f, 0x6e, 0x66, 0x52, 0x06, 0x53,
	0x74, 0x61, 0x74, 0x69, 0x63, 0x12, 0x4b, 0x0a, 0x09, 0x44, 0x4e, 0x53, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x73, 0x69, 0x65, 0x6d, 0x65,
	0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x2e, 0x44, 0x6e, 0x73, 0x52, 0x09, 0x44, 0x4e, 0x53, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x44, 0x0a, 0x06, 0x4c, 0x32, 0x43, 0x6f, 0x6e, 0x66, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64,
	0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x2e, 0x4c, 0x32,
	0x52, 0x06, 0x4c, 0x32, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x24, 0x0a, 0x0d, 0x49, 0x6e, 0x74, 0x65,
	0x72, 0x66, 0x61, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x1a, 0x54,
	0x0a, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x69, 0x63, 0x43, 0x6f, 0x6e, 0x66, 0x12, 0x12, 0x0a, 0x04,
	0x49, 0x50, 0x76, 0x34, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x49, 0x50, 0x76, 0x34,
	0x12, 0x18, 0x0a, 0x07, 0x4e, 0x65, 0x74, 0x4d, 0x61, 0x73, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x4e, 0x65, 0x74, 0x4d, 0x61, 0x73, 0x6b, 0x12, 0x18, 0x0a, 0x07, 0x47, 0x61,
	0x74, 0x65, 0x77, 0x61, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x47, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x1a, 0x49, 0x0a, 0x03, 0x44, 0x6e, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x50,
	0x72, 0x69, 0x6d, 0x61, 0x72, 0x79, 0x44, 0x4e, 0x53, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x50, 0x72, 0x69, 0x6d, 0x61, 0x72, 0x79, 0x44, 0x4e, 0x53, 0x12, 0x22, 0x0a, 0x0c, 0x53,
	0x65, 0x63, 0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x44, 0x4e, 0x53, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x44, 0x4e, 0x53, 0x1a,
	0xbd, 0x02, 0x0a, 0x02, 0x4c, 0x32, 0x12, 0x30, 0x0a, 0x13, 0x53, 0x74, 0x61, 0x72, 0x74, 0x69,
	0x6e, 0x67, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x49, 0x50, 0x76, 0x34, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x13, 0x53, 0x74, 0x61, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x49, 0x50, 0x76, 0x34, 0x12, 0x18, 0x0a, 0x07, 0x4e, 0x65, 0x74, 0x4d,
	0x61, 0x73, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4e, 0x65, 0x74, 0x4d, 0x61,
	0x73, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x47, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x47, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x12, 0x74, 0x0a, 0x12, 0x41, 0x75, 0x78, 0x69, 0x6c, 0x69, 0x61, 0x72, 0x79, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x44,
	0x2e, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64,
	0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x2e, 0x4c, 0x32, 0x2e, 0x41, 0x75, 0x78,
	0x69, 0x6c, 0x69, 0x61, 0x72, 0x79, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x12, 0x41, 0x75, 0x78, 0x69, 0x6c, 0x69, 0x61, 0x72, 0x79, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x1a, 0x45, 0x0a, 0x17, 0x41, 0x75, 0x78, 0x69,
	0x6c, 0x69, 0x61, 0x72, 0x79, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x5c, 0x0a, 0x0f, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e,
	0x67, 0x73, 0x12, 0x49, 0x0a, 0x0a, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73,
	0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63,
	0x65, 0x52, 0x0a, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x32, 0xc2, 0x02,
	0x0a, 0x0e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x5b, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x41, 0x6c, 0x6c, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x2f, 0x2e, 0x73,
	0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61,
	0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x79, 0x0a,
	0x13, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x57, 0x69, 0x74,
	0x68, 0x4d, 0x61, 0x63, 0x12, 0x37, 0x2e, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x2e, 0x69,
	0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e,
	0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d,
	0x61, 0x70, 0x69, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x49,
	0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x58, 0x0a, 0x0d, 0x41, 0x70, 0x70, 0x6c,
	0x79, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x2f, 0x2e, 0x73, 0x69, 0x65, 0x6d,
	0x65, 0x6e, 0x73, 0x2e, 0x69, 0x65, 0x64, 0x67, 0x65, 0x2e, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x2e,
	0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x42, 0x1a, 0x5a, 0x18, 0x2e, 0x3b, 0x73, 0x69, 0x65, 0x6d, 0x65, 0x6e, 0x73, 0x5f,
	0x69, 0x65, 0x64, 0x67, 0x65, 0x5f, 0x64, 0x6d, 0x61, 0x70, 0x69, 0x5f, 0x76, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_Network_proto_rawDescOnce sync.Once
	file_Network_proto_rawDescData = file_Network_proto_rawDesc
)

func file_Network_proto_rawDescGZIP() []byte {
	file_Network_proto_rawDescOnce.Do(func() {
		file_Network_proto_rawDescData = protoimpl.X.CompressGZIP(file_Network_proto_rawDescData)
	})
	return file_Network_proto_rawDescData
}

var file_Network_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_Network_proto_goTypes = []interface{}{
	(*NetworkInterfaceRequest)(nil), // 0: siemens.iedge.dmapi.network.v1.NetworkInterfaceRequest
	(*Interface)(nil),               // 1: siemens.iedge.dmapi.network.v1.Interface
	(*NetworkSettings)(nil),         // 2: siemens.iedge.dmapi.network.v1.NetworkSettings
	(*Interface_StaticConf)(nil),    // 3: siemens.iedge.dmapi.network.v1.Interface.StaticConf
	(*Interface_Dns)(nil),           // 4: siemens.iedge.dmapi.network.v1.Interface.Dns
	(*Interface_L2)(nil),            // 5: siemens.iedge.dmapi.network.v1.Interface.L2
	nil,                             // 6: siemens.iedge.dmapi.network.v1.Interface.L2.AuxiliaryAddressesEntry
	(*empty.Empty)(nil),             // 7: google.protobuf.Empty
}
var file_Network_proto_depIdxs = []int32{
	3, // 0: siemens.iedge.dmapi.network.v1.Interface.Static:type_name -> siemens.iedge.dmapi.network.v1.Interface.StaticConf
	4, // 1: siemens.iedge.dmapi.network.v1.Interface.DNSConfig:type_name -> siemens.iedge.dmapi.network.v1.Interface.Dns
	5, // 2: siemens.iedge.dmapi.network.v1.Interface.L2Conf:type_name -> siemens.iedge.dmapi.network.v1.Interface.L2
	1, // 3: siemens.iedge.dmapi.network.v1.NetworkSettings.Interfaces:type_name -> siemens.iedge.dmapi.network.v1.Interface
	6, // 4: siemens.iedge.dmapi.network.v1.Interface.L2.AuxiliaryAddresses:type_name -> siemens.iedge.dmapi.network.v1.Interface.L2.AuxiliaryAddressesEntry
	7, // 5: siemens.iedge.dmapi.network.v1.NetworkService.GetAllInterfaces:input_type -> google.protobuf.Empty
	0, // 6: siemens.iedge.dmapi.network.v1.NetworkService.GetInterfaceWithMac:input_type -> siemens.iedge.dmapi.network.v1.NetworkInterfaceRequest
	2, // 7: siemens.iedge.dmapi.network.v1.NetworkService.ApplySettings:input_type -> siemens.iedge.dmapi.network.v1.NetworkSettings
	2, // 8: siemens.iedge.dmapi.network.v1.NetworkService.GetAllInterfaces:output_type -> siemens.iedge.dmapi.network.v1.NetworkSettings
	1, // 9: siemens.iedge.dmapi.network.v1.NetworkService.GetInterfaceWithMac:output_type -> siemens.iedge.dmapi.network.v1.Interface
	7, // 10: siemens.iedge.dmapi.network.v1.NetworkService.ApplySettings:output_type -> google.protobuf.Empty
	8, // [8:11] is the sub-list for method output_type
	5, // [5:8] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_Network_proto_init() }
func file_Network_proto_init() {
	if File_Network_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_Network_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkInterfaceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_Network_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Interface); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_Network_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkSettings); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_Network_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Interface_StaticConf); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_Network_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Interface_Dns); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_Network_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Interface_L2); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_Network_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_Network_proto_goTypes,
		DependencyIndexes: file_Network_proto_depIdxs,
		MessageInfos:      file_Network_proto_msgTypes,
	}.Build()
	File_Network_proto = out.File
	file_Network_proto_rawDesc = nil
	file_Network_proto_goTypes = nil
	file_Network_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// NetworkServiceClient is the client API for NetworkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NetworkServiceClient interface {
	//Returns the settings of all ethernet typed network interfaces
	GetAllInterfaces(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*NetworkSettings, error)
	//Returns the current setting with given MAC address.
	GetInterfaceWithMac(ctx context.Context, in *NetworkInterfaceRequest, opts ...grpc.CallOption) (*Interface, error)
	//Applies given configurations to Network Interfaces.
	ApplySettings(ctx context.Context, in *NetworkSettings, opts ...grpc.CallOption) (*empty.Empty, error)
}

type networkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNetworkServiceClient(cc grpc.ClientConnInterface) NetworkServiceClient {
	return &networkServiceClient{cc}
}

func (c *networkServiceClient) GetAllInterfaces(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*NetworkSettings, error) {
	out := new(NetworkSettings)
	err := c.cc.Invoke(ctx, "/siemens.iedge.dmapi.network.v1.NetworkService/GetAllInterfaces", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *networkServiceClient) GetInterfaceWithMac(ctx context.Context, in *NetworkInterfaceRequest, opts ...grpc.CallOption) (*Interface, error) {
	out := new(Interface)
	err := c.cc.Invoke(ctx, "/siemens.iedge.dmapi.network.v1.NetworkService/GetInterfaceWithMac", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *networkServiceClient) ApplySettings(ctx context.Context, in *NetworkSettings, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/siemens.iedge.dmapi.network.v1.NetworkService/ApplySettings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NetworkServiceServer is the server API for NetworkService service.
type NetworkServiceServer interface {
	//Returns the settings of all ethernet typed network interfaces
	GetAllInterfaces(context.Context, *empty.Empty) (*NetworkSettings, error)
	//Returns the current setting with given MAC address.
	GetInterfaceWithMac(context.Context, *NetworkInterfaceRequest) (*Interface, error)
	//Applies given configurations to Network Interfaces.
	ApplySettings(context.Context, *NetworkSettings) (*empty.Empty, error)
}

// UnimplementedNetworkServiceServer can be embedded to have forward compatible implementations.
type UnimplementedNetworkServiceServer struct {
}

func (*UnimplementedNetworkServiceServer) GetAllInterfaces(context.Context, *empty.Empty) (*NetworkSettings, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllInterfaces not implemented")
}
func (*UnimplementedNetworkServiceServer) GetInterfaceWithMac(context.Context, *NetworkInterfaceRequest) (*Interface, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInterfaceWithMac not implemented")
}
func (*UnimplementedNetworkServiceServer) ApplySettings(context.Context, *NetworkSettings) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplySettings not implemented")
}

func RegisterNetworkServiceServer(s *grpc.Server, srv NetworkServiceServer) {
	s.RegisterService(&_NetworkService_serviceDesc, srv)
}

func _NetworkService_GetAllInterfaces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkServiceServer).GetAllInterfaces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/siemens.iedge.dmapi.network.v1.NetworkService/GetAllInterfaces",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkServiceServer).GetAllInterfaces(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _NetworkService_GetInterfaceWithMac_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkInterfaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkServiceServer).GetInterfaceWithMac(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/siemens.iedge.dmapi.network.v1.NetworkService/GetInterfaceWithMac",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkServiceServer).GetInterfaceWithMac(ctx, req.(*NetworkInterfaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NetworkService_ApplySettings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkSettings)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkServiceServer).ApplySettings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/siemens.iedge.dmapi.network.v1.NetworkService/ApplySettings",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkServiceServer).ApplySettings(ctx, req.(*NetworkSettings))
	}
	return interceptor(ctx, in, info, handler)
}

var _NetworkService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "siemens.iedge.dmapi.network.v1.NetworkService",
	HandlerType: (*NetworkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAllInterfaces",
			Handler:    _NetworkService_GetAllInterfaces_Handler,
		},
		{
			MethodName: "GetInterfaceWithMac",
			Handler:    _NetworkService_GetInterfaceWithMac_Handler,
		},
		{
			MethodName: "ApplySettings",
			Handler:    _NetworkService_ApplySettings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Network.proto",
}
