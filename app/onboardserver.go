/*
 * Copyright (c) 2022 Siemens AG
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package app

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	v1 "onboardservice/api/siemens_iedge_dmapi_v1"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OnboardServer as GRPC Server
type OnboardServer struct {
	v1.UnimplementedOnboardServiceServer
	listeners map[int32]listenerInfo
	status    *v1.OnboardingStatus
	parentApp *MainApp
	sync.Mutex
	isContainerUp bool
}

const (
	leaveMessage        = "ApplyConfiguration leave."
	daemonReloadCommand = "/usr/bin/systemctl daemon-reload"
)

// GRPC method implementations ################################################################################
// ############################################################################################################

// OnboardWithUSB  Applies all settings then registers the device.
func (o *OnboardServer) OnboardWithUSB(ctx context.Context,
	configuration *v1.DeviceConfiguration) (*emptypb.Empty, error) {
	log.Println("Checking activation status of  device  ...")

	if o.isContainerUp == false {
		return &emptypb.Empty{}, status.New(codes.Internal, "Try again later, since edge core runtime is not ready !: ").Err()
	}

	val, _ := o.parentApp.Clients.RestClient.Onboarded()

	if val == true {
		log.Println("Device is already onboarded, skipping given config!")
		return &emptypb.Empty{}, status.New(codes.Internal, "Device is already onboarded !: ").Err()
	}

	_, err := o.ApplyConfiguration(ctx, configuration)
	if err != nil {
		return &emptypb.Empty{}, status.New(codes.Internal, "Applying Configurations failed: "+err.Error()).Err()
	}

	log.Println("Activating device  ...")
	//REST CALL TO EDGECORE
	res, err := o.parentApp.Clients.RestClient.Activate(configuration)
	if res == false {
		log.Println("Device Activation failed ...")
		return &emptypb.Empty{}, status.New(codes.Internal, "Activation failed: "+err.Error()).Err()
	}
	log.Println("Device Activation successful...")
	o.SetOnboardingStatus(ctx, &v1.OnboardingStatus{IsOnboarded: true})
	return &emptypb.Empty{}, status.New(codes.OK, "fine").Err()
}

// ApplyConfiguration calls all required IEDK services to perform configurations for onboarding.
func (o *OnboardServer) ApplyConfiguration(ctx context.Context, configuration *v1.DeviceConfiguration) (
	*emptypb.Empty, error) {
	void := &emptypb.Empty{}
	log.Println("ApplyConfiguration enter.")
	o.lock()
	defer o.unLock()

	if err := o.configureNTP(ctx, configuration); err != nil {
		log.Println(leaveMessage)
		return void, err
	}
	if err := o.configureNetwork(ctx, configuration); err != nil {
		log.Println(leaveMessage)
		return void, err
	}
	if err := o.parentApp.helper.DaemonReload(daemonReloadCommand); err == false {
		log.Println("Daemon Reload Failed")
	}

	if err := o.configureCustomSettings(ctx, configuration.Device.CustomConfiguration); err != nil {
		log.Println("error configureCustomSettings : ", err)
	}

	if err := o.write2Device(configuration.Onboarding.DeviceId, "/var/device.id"); err != nil {
		log.Println("An error occurred while writing the device ID to the file")
		return void, err
	}

	if err := o.write2Device(configuration.Onboarding.DeviceName, "/var/device.name"); err != nil {
		log.Println("An error occurred while writing the device Name to the file")
		return void, err
	}

	log.Println(leaveMessage)
	return void, status.New(codes.OK, "Applied All Settings").Err()
}
func (o *OnboardServer) lock() {
	log.Println("Lock is starting...")
	o.Mutex.Lock()
	log.Println("Locked")
}

func (o *OnboardServer) unLock() {
	log.Println("Unlock is starting...")
	o.Mutex.Unlock()
	log.Println("Unlocked")
}

// SetOnboardingStatus  sets the information of  device's onboarding status and changes led states.
func (o *OnboardServer) SetOnboardingStatus(ctx context.Context, onboardingStatus *v1.OnboardingStatus) (*emptypb.Empty,
	error) {
	log.Println("SetOnboardingStatus enter.")
	o.lock()
	defer o.unLock()
	if onboardingStatus != nil {

		//this routine updates the channel with new states for streaming
		go func() {
			log.Println("Listener Update routine enter.")
			if len(o.listeners) > 0 {
				for _, k := range o.listeners {
					k.channel <- onboardingStatus
				}
			}
			log.Println("Listener Update routine leave.")
		}()
		log.Println("SetOnboardingStatus leave.")
		return &emptypb.Empty{}, status.New(codes.OK, "Onboard Status has been set successfully").Err()
	}
	log.Println("SetOnboardingStatus leave.")
	return &emptypb.Empty{}, status.New(codes.InvalidArgument, "Setting onboard status failed").Err()
}

// ListenOnboardState Implementation of RPC method given Onboard proto file When a client calls this once, all new
// onboarding states delivered until client is closed.
func (o *OnboardServer) ListenOnboardState(e *emptypb.Empty, listener v1.OnboardService_ListenOnboardStateServer) error {
	listener.Send(o.status) // initial message for the first time connected client (retained like in MQTT )

	linfo := createNewListenerInfo(listener)

	o.Mutex.Lock()
	o.listeners[linfo.id] = linfo
	o.Mutex.Unlock()
	log.Println("New listener added")
	log.Println("Listeners count:", len(o.listeners))

	//main event loop
	for {
		select {
		case <-listener.Context().Done():
			{
				o.Mutex.Lock()
				delete(o.listeners, linfo.id)
				o.Mutex.Unlock()
				log.Println("Client removed: ", linfo.id)
				log.Println("Listeners count: ", len(o.listeners))

				return nil
			}

		case res := <-linfo.channel:
			{
				err := linfo.listener.Send(res)
				if err == nil {
					log.Printf("Message delivered succesfully to client : %v %v", linfo.id, res)
				}
			}
		}
	}
}
