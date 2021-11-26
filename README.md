# IEDK Network Service
_Network Service_ is gRPC based network settings configurator microservice. It also can gather current network settings.

## Overview

_Network Service_ is developed in the Go programming language and gRPC. More information can be found [here](https://grpc.io/docs/). The _Network Service_ runs as a systemd service within the device that has a debian-based operating system.






fasdfasdf





asdgasdg
asdgasdg
asdg





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


## Development

> gRPC can be used with many different programming languages. For this reason, if you want to develop an IEDK service using golang or different programming languages, the following documentations will be useful. [Golang Tutorial](https://industrial-edge-device-builders.code.siemens.io/documentation/howtos/go-tutorial/),
[Python Tutorial](https://industrial-edge-device-builders.code.siemens.io/documentation/howtos/python-tutorial/),
[C++ Tutorial](https://industrial-edge-device-builders.code.siemens.io/documentation/howtos/cpp-tutorial/)
