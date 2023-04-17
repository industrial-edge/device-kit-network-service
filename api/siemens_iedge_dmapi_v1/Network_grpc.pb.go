// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: Network.proto

package siemens_iedge_dmapi_v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// NetworkServiceClient is the client API for NetworkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NetworkServiceClient interface {
	// Returns the settings of all ethernet typed network interfaces
	GetAllInterfaces(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*NetworkSettings, error)
	// Returns the current setting for the interface, with given MAC address.
	GetInterfaceWithMac(ctx context.Context, in *NetworkInterfaceRequest, opts ...grpc.CallOption) (*Interface, error)
	// Returns the current setting for the interface,  with given Label.
	GetInterfaceWithLabel(ctx context.Context, in *NetworkInterfaceRequestWithLabel, opts ...grpc.CallOption) (*Interface, error)
	// Applies given configurations to Network Interfaces.
	ApplySettings(ctx context.Context, in *NetworkSettings, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type networkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNetworkServiceClient(cc grpc.ClientConnInterface) NetworkServiceClient {
	return &networkServiceClient{cc}
}

func (c *networkServiceClient) GetAllInterfaces(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*NetworkSettings, error) {
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

func (c *networkServiceClient) GetInterfaceWithLabel(ctx context.Context, in *NetworkInterfaceRequestWithLabel, opts ...grpc.CallOption) (*Interface, error) {
	out := new(Interface)
	err := c.cc.Invoke(ctx, "/siemens.iedge.dmapi.network.v1.NetworkService/GetInterfaceWithLabel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *networkServiceClient) ApplySettings(ctx context.Context, in *NetworkSettings, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/siemens.iedge.dmapi.network.v1.NetworkService/ApplySettings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NetworkServiceServer is the server API for NetworkService service.
// All implementations must embed UnimplementedNetworkServiceServer
// for forward compatibility
type NetworkServiceServer interface {
	// Returns the settings of all ethernet typed network interfaces
	GetAllInterfaces(context.Context, *emptypb.Empty) (*NetworkSettings, error)
	// Returns the current setting for the interface, with given MAC address.
	GetInterfaceWithMac(context.Context, *NetworkInterfaceRequest) (*Interface, error)
	// Returns the current setting for the interface,  with given Label.
	GetInterfaceWithLabel(context.Context, *NetworkInterfaceRequestWithLabel) (*Interface, error)
	// Applies given configurations to Network Interfaces.
	ApplySettings(context.Context, *NetworkSettings) (*emptypb.Empty, error)
	mustEmbedUnimplementedNetworkServiceServer()
}

// UnimplementedNetworkServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNetworkServiceServer struct {
}

func (UnimplementedNetworkServiceServer) GetAllInterfaces(context.Context, *emptypb.Empty) (*NetworkSettings, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllInterfaces not implemented")
}
func (UnimplementedNetworkServiceServer) GetInterfaceWithMac(context.Context, *NetworkInterfaceRequest) (*Interface, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInterfaceWithMac not implemented")
}
func (UnimplementedNetworkServiceServer) GetInterfaceWithLabel(context.Context, *NetworkInterfaceRequestWithLabel) (*Interface, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInterfaceWithLabel not implemented")
}
func (UnimplementedNetworkServiceServer) ApplySettings(context.Context, *NetworkSettings) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplySettings not implemented")
}
func (UnimplementedNetworkServiceServer) mustEmbedUnimplementedNetworkServiceServer() {}

// UnsafeNetworkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NetworkServiceServer will
// result in compilation errors.
type UnsafeNetworkServiceServer interface {
	mustEmbedUnimplementedNetworkServiceServer()
}

func RegisterNetworkServiceServer(s grpc.ServiceRegistrar, srv NetworkServiceServer) {
	s.RegisterService(&NetworkService_ServiceDesc, srv)
}

func _NetworkService_GetAllInterfaces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
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
		return srv.(NetworkServiceServer).GetAllInterfaces(ctx, req.(*emptypb.Empty))
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

func _NetworkService_GetInterfaceWithLabel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkInterfaceRequestWithLabel)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkServiceServer).GetInterfaceWithLabel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/siemens.iedge.dmapi.network.v1.NetworkService/GetInterfaceWithLabel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkServiceServer).GetInterfaceWithLabel(ctx, req.(*NetworkInterfaceRequestWithLabel))
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

// NetworkService_ServiceDesc is the grpc.ServiceDesc for NetworkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NetworkService_ServiceDesc = grpc.ServiceDesc{
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
			MethodName: "GetInterfaceWithLabel",
			Handler:    _NetworkService_GetInterfaceWithLabel_Handler,
		},
		{
			MethodName: "ApplySettings",
			Handler:    _NetworkService_ApplySettings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Network.proto",
}
