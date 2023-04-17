/*
 * Copyright (c) 2022 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package app

import (
	"context"
	"log"
	"net"
	v1 "onboardservice/api/siemens_iedge_dmapi_v1"
	edgecore "onboardservice/internal/restclient"
	"onboardservice/internal/system"

	"google.golang.org/grpc"
)

const (
	socketBasePath       = "/var/run/devicemodel/"
	networkServerAddress = socketBasePath + "network.sock"
	ntpServerAddress     = socketBasePath + "ntp.sock"
	systemServerAddress  = socketBasePath + "system.sock"
	restAPI              = "https://localhost/device/edge/b.service/api"
)

// ClientPack type contains IEDK Services and edgeCoreRuntime Rest API
type ClientPack struct {
	NetworkClient v1.NetworkServiceClient
	NtpClient     v1.NtpServiceClient
	SystemClient  v1.SystemServiceClient
	RestClient    edgecore.EdgeCoreRuntimeRest
}

// ClientFactory Interface for service clients
type ClientFactory interface {
	CreateClients() *ClientPack
	CreateHelper() system.Configurable
}

// ClientFactoryImpl implements ClientFactory interface
type ClientFactoryImpl struct {
}

func createConnection(address string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}))
	if err != nil {
		log.Printf("Could not connect to: %v", err)
	}
	return conn
}

// CreateClients creates a ClientPack that contains all IEDK Services that onboard service needs.
func (o ClientFactoryImpl) CreateClients() *ClientPack {
	ntpConn := createConnection(ntpServerAddress)
	networkConn := createConnection(networkServerAddress)
	cli := edgecore.NewClient(restAPI)
	nc := v1.NewNetworkServiceClient(networkConn)
	ntc := v1.NewNtpServiceClient(ntpConn)

	pack := ClientPack{
		NetworkClient: nc,
		NtpClient:     ntc,
		RestClient:    cli,
	}
	return &pack
}

// CreateHelper creates helper class that generates and Nginx Configurations.
func (o ClientFactoryImpl) CreateHelper() system.Configurable {

	return &system.Helper{}
}
