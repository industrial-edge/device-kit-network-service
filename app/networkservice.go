/*
 * Copyright © Siemens 2020 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking"
	"networkservice/internal/networking/common"

	"os"
	"sync"
	"time"

	nm "github.com/Wifx/gonetworkmanager/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrNoGSMDevice = errors.New("no gsm device found")

// IEDKService typical IEDK service should implement this.
type IEDKService interface {
	StartGRPC(args []string)
	StartApp()
}

// CreateServiceApp creates main app
func CreateServiceApp() *MainApp {
	nm, err := nm.NewNetworkManager()
	if err != nil {
		log.Printf("failed to connect to NetworkManager: %v", err)
		return nil
	}
	return &MainApp{
		serverInstance: &networkServer{
			configurator: networking.NewNetworkConfiguratorWithNM(nm),
		},
	}
}

type networkServer struct {
	v1.UnimplementedNetworkServiceServer
	configurator *networking.NetworkConfigurator
	sync.Mutex
}

// MainApp type for Network
type MainApp struct {
	serverInstance *networkServer
}

const errMsgInterfaceNotFound = "Interface does not exist on this device!"
const errMsgCreateConnectionFailed = "Failed to create connection!"
const errMsgRemoveConnectionFailed = "Failed to remove connection!"

// StartGRPC starts GPRC listen server.
func (app *MainApp) StartGRPC(args []string) error {
	const message string = "ERROR: Could not start monitor with bad arguments! \n " +
		"Sample usage:\n  ./networkservice unix /tmp/devicemodel/ntp.socket \n" +
		"  ./networkservice tcp localhost:50006"

	if len(args) != 3 {
		fmt.Println(message)
		return errors.New("parameter not supported")
	}
	typeOfConnection := args[1]
	address := args[2]
	if typeOfConnection != "unix" && typeOfConnection != "tcp" {
		fmt.Println(message)
		return errors.New("parameter not supported: " + typeOfConnection)
	}

	if typeOfConnection == "unix" {

		if err := os.RemoveAll(os.Args[2]); err != nil {
			return errors.New("socket could not removed: " + typeOfConnection)
		}

	}

	lis, err := net.Listen(typeOfConnection, address)

	if err != nil {
		log.Println("Failed to listen: ", err.Error())
		return errors.New("Failed to listen: " + err.Error())

	}
	if typeOfConnection == "unix" {
		err = chownSocket(address, "root", "docker")
		if err != nil {
			return err
		}
	}

	log.Print("Started listening on : ", typeOfConnection, " - ", address)
	s := grpc.NewServer()

	v1.RegisterNetworkServiceServer(s, app.serverInstance)
	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve: %v", err)
		return errors.New("Failed to serve: " + err.Error())
	}

	return nil
}

// StartApp starts additional tasks during start stage.
func (app *MainApp) StartApp() {
	// No need for additional start up .
}

// GetAllInterfaces Returns all ETHERNET Typed network interface settings.
func (n *networkServer) GetAllInterfaces(ctx context.Context, e *emptypb.Empty) (*v1.NetworkSettings, error) {

	log.Println("GetAllInterfaces() called")
	n.Lock()

	retVal := &v1.NetworkSettings{Interfaces: n.configurator.GetNetworkInterfaces()}

	n.Unlock()
	log.Println("GetAllInterfaces() done")

	return retVal, status.New(codes.OK, "Get All Interfaces Done!").Err()
}

// GetInterfaceWithMac returns the device mathcing with given mac, else returns error.
func (n *networkServer) GetInterfaceWithMac(ctx context.Context,
	request *v1.NetworkInterfaceRequest) (*v1.Interface, error) {

	log.Println("GetInterfaceWithMac() called")
	n.Lock()
	defer n.Unlock()

	retVal := n.configurator.GetInterfaceWithMac(request.Mac)
	if retVal == nil {
		log.Println(errMsgInterfaceNotFound)
		return nil, status.New(codes.NotFound, errMsgInterfaceNotFound).Err()
	}

	// Check if it is the gateway interface.
	if n.configurator.IsGatewayInterface(request.Mac) {
		retVal.GatewayInterface = true
	}

	log.Println("GetInterfaceWithMac() done")

	return retVal, status.New(codes.OK, "GetInterfaceWithMac Done!").Err()
}

// ApplySettings applies given network configurations via NetworkManager
func (n *networkServer) ApplySettings(ctx context.Context, newSettings *v1.NetworkSettings) (*emptypb.Empty, error) {
	result := status.New(codes.OK, "Apply Settings Done!").Err()

	_, err := n.configurator.ArePreconditionsOk(newSettings)

	if err != nil {
		result = status.New(codes.FailedPrecondition, fmt.Sprintf("Wrong input for this method, %v", err)).Err()

	} else {
		//APPLY THE NEW SETTINGS
		n.Lock()

		if (nil != newSettings.LabelMap) && (0 != len(newSettings.LabelMap)) {
			err = networking.WriteMapToFile(newSettings.LabelMap, common.LabelMapFileName)
		}

		if err == nil {
			err = n.configurator.Apply(newSettings)
		}

		defer n.Unlock()
		if err != nil {
			result = status.New(codes.Internal,
				fmt.Sprintf("Errors occured while applying new settings,  %v", err)).Err()
		}
		// After the device's network configurations are set, a 5 second sleep is set because it takes time to apply to the device.
		time.Sleep(5 * time.Second)
	}

	return &emptypb.Empty{}, result

}

func (n *networkServer) GetInterfaceWithLabel(ctx context.Context, request *v1.NetworkInterfaceRequestWithLabel) (*v1.Interface, error) {

	log.Println("GetInterfaceWithLabel() called")
	n.Lock()
	state := status.New(codes.OK, "GetInterfaceWithLabel Done!").Err()
	retVal := n.configurator.GetInterfaceWithLabel(request.Label)
	if retVal == nil {
		state = status.New(codes.NotFound, errMsgInterfaceNotFound).Err()
	}

	n.Unlock()
	log.Println("GetInterfaceWithLabel() done")

	return retVal, state
}

// CreateConnection creates a new GSM network connection with given settings.
func (n *networkServer) CreateConnection(ctx context.Context, connSettings *v1.ConnectionSettings) (*emptypb.Empty, error) {

	log.Println("CreateConnection() called")
	n.Lock()
	state := status.New(codes.OK, "CreateConnection Done!").Err()

	err := n.configurator.CreateConnection(connSettings)

	if err != nil {
		log.Println("Failed to create  connection:", err)
		if errors.Is(err, ErrNoGSMDevice) {
			state = status.New(codes.NotFound, err.Error()).Err()
		} else {
			state = status.New(codes.Internal, errMsgCreateConnectionFailed).Err()
		}
	}
	n.Unlock()
	log.Println("CreateConnection() done")
	return &emptypb.Empty{}, state
}

// RemoveConnection removes an existing GSM network connection with given settings.
func (n *networkServer) RemoveConnection(ctx context.Context, connSettings *v1.ConnectionSettings) (*emptypb.Empty, error) {

	log.Println("RemoveConnection() called")
	n.Lock()
	state := status.New(codes.OK, "RemoveConnection  Done!").Err()

	err := n.configurator.RemoveConnection(connSettings)
	if err != nil {
		state = status.New(codes.Internal, errMsgRemoveConnectionFailed).Err()
	}
	n.Unlock()
	log.Println("RemoveConnection() done")
	return &emptypb.Empty{}, state
}
