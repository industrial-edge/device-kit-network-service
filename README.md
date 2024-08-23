# Introduction

The IE Device Kit API provides the abstraction layer that decouples the Industrial Edge Runtime from the underlying Linux systems. This allows to adapt the runtime and its behavior to serve for the specific needs of different Industrial Edge products. 

The IE Device Kit API is based on gRPC which provides a modern intermediate process communication style for building distributed applications and microservices. The Industrial Edge platform provides and maintains the protobuf specification files for the APIs contained in the IE Device Kit. These protobuf specifications can be used to create stub implementations for both client and server in various programming languages. The Industrial Edge Runtime ships with a client side implementation of these APIs and expects the host system to provide a server side implementation.

Purpose of these repositories is to share reference implementation of IE Device Kit APIs. You can use existing implementation or adapt it based on your needs.

# IEDK Network Service
_Network Service_ is a gRPC & Go based network configuration microservice. The network settings of the Edge Devices are configured through this service.

```bash
    //Returns the settings of all ethernet typed network interfaces
    rpc GetAllInterfaces(google.protobuf.Empty) returns(NetworkSettings);
   
    //Returns the current setting for the interface, with given MAC address.
    rpc GetInterfaceWithMac(NetworkInterfaceRequest) returns(Interface);

    //Returns the current setting for the interface,  with given Label.
    rpc GetInterfaceWithLabel(NetworkInterfaceRequestWithLabel) returns(Interface);
       
    //Applies given configurations to Network Interfaces.
    rpc ApplySettings(NetworkSettings) returns(google.protobuf.Empty);

```

## Overview

_Network Service_ is developed in the Go programming language and gRPC. More information can be found [here](https://grpc.io/docs/). The _Network Service_ runs as a systemd service within the device that has a debian-based operating system.


## Getting Started

### Prerequisities

> - Setting up Go (For build)
> - Additional requirements to _run_ or _develop_ the project


### Building the service and know-how about other features

> Instructions how to build the deb package:
>
> - For generating a deb package, [goreleaser](https://goreleaser.com/intro/) tool is heavily used. 
>   For proper execution, goreleaser needs a TAG identifier which indicates version of the deb package. __TAG__ must obey the semantic versioning rules and must include ```${major}.${minor}.${hotfix}``` release identifiers. 
>   Only two commands are needed for generating deb package: `cd build/package`, `TAG=X.Y.Z make deb`. After running these commands, a __dist__ directory will be created under the __build/package__ directory and the deb package will be in dist directory. 
>   To install this generated deb package on the device as a daemon `cd dist`, `sudo apt install ./dm-network_X.Y.Z_linux_amd64.deb`
>
> Instructions how to use the make command:
>
> - There is a Makefile file under the __build/package__ directory. Unit tests, code coverage and many other similar features can be used via this Makefile. The following commands are used to view all features: `cd build/package`, `make help`


### Running the service

> To see the status and logs of the deb package running as daemon(systemd service) directly from the command line, the following commands can be run: `systemctl status dm-network`, `journalctl -fu dm-network`

# Contributing IE Device Kit Repository
Please check our [contribution guideline](CONTRIBUTING.md). 

# Contribution License Agreement

If you haven't previously signed the [Siemens Contributor License Agreement](https://cla-assistant.io/industrial-edge/) (CLA), the system will automatically prompt you to do so when you submit your Pull Request. This can be conveniently done through the CLA Assistant's online platform.
Once the CLA is signed, your Pull Request will automatically be cleared and made ready for merging if all other test stages succeed.


# How to be part of Siemens Industrial Edge Ecosystem
Please check [this](https://new.siemens.com/global/en/products/automation/topic-areas/industrial-edge.html) page to learn more information about Industrial Edge.
