# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [Network.proto](#Network.proto)
    - [Interface](#siemens.iedge.dmapi.network.v1.Interface)
    - [Interface.Dns](#siemens.iedge.dmapi.network.v1.Interface.Dns)
    - [Interface.L2](#siemens.iedge.dmapi.network.v1.Interface.L2)
    - [Interface.L2.AuxiliaryAddressesEntry](#siemens.iedge.dmapi.network.v1.Interface.L2.AuxiliaryAddressesEntry)
    - [Interface.StaticConf](#siemens.iedge.dmapi.network.v1.Interface.StaticConf)
    - [NetworkInterfaceRequest](#siemens.iedge.dmapi.network.v1.NetworkInterfaceRequest)
    - [NetworkInterfaceRequestWithLabel](#siemens.iedge.dmapi.network.v1.NetworkInterfaceRequestWithLabel)
    - [NetworkSettings](#siemens.iedge.dmapi.network.v1.NetworkSettings)
    - [NetworkSettings.LabelMapEntry](#siemens.iedge.dmapi.network.v1.NetworkSettings.LabelMapEntry)
  
    - [NetworkService](#siemens.iedge.dmapi.network.v1.NetworkService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="Network.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## Network.proto



<a name="siemens.iedge.dmapi.network.v1.Interface"></a>

### Interface
Interface type holds settings for a Network Interface.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| GatewayInterface | [bool](#bool) |  | if true, route metric will be set to 1. Otherwise route metric is -1. Similarly, when the interface is requested,return value will be true if route metric is 1. |
| MacAddress | [string](#string) |  | "20:87:56:b5:ed:e0" |
| DHCP | [string](#string) |  | values can be 'enabled' or 'disabled'. for compatiblity reasons it is not boolean. |
| Static | [Interface.StaticConf](#siemens.iedge.dmapi.network.v1.Interface.StaticConf) |  | Static field is StaticConf type instance. |
| DNSConfig | [Interface.Dns](#siemens.iedge.dmapi.network.v1.Interface.Dns) |  | DNSConfig is dns type instance. |
| L2Conf | [Interface.L2](#siemens.iedge.dmapi.network.v1.Interface.L2) |  |  |
| InterfaceName | [string](#string) |  | ens2p |
| Label | [string](#string) |  | x1 |






<a name="siemens.iedge.dmapi.network.v1.Interface.Dns"></a>

### Interface.Dns
Type that contains Primary and Secondary DNS.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| PrimaryDNS | [string](#string) |  | e.g: "1.1.1.2" |
| SecondaryDNS | [string](#string) |  | e.g: "1.1.1.1" |






<a name="siemens.iedge.dmapi.network.v1.Interface.L2"></a>

### Interface.L2



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| StartingAddressIPv4 | [string](#string) |  | e.g: 192.168.0.2 |
| NetMask | [string](#string) |  | e.g: 255.255.255.0 |
| Range | [string](#string) |  | e.g: 16 |
| Gateway | [string](#string) |  | e.g: 192.168.2.1 |
| AuxiliaryAddresses | [Interface.L2.AuxiliaryAddressesEntry](#siemens.iedge.dmapi.network.v1.Interface.L2.AuxiliaryAddressesEntry) | repeated | Preserved addresses for other devices. These addresses won't be assigned to containers. e.g my_plc, 192.168.0.5 |






<a name="siemens.iedge.dmapi.network.v1.Interface.L2.AuxiliaryAddressesEntry"></a>

### Interface.L2.AuxiliaryAddressesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="siemens.iedge.dmapi.network.v1.Interface.StaticConf"></a>

### Interface.StaticConf
StaticConf type holds IP Netmask and Gateway information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| IPv4 | [string](#string) |  | e.g: 192.168.0.2 |
| NetMask | [string](#string) |  | e.g: 255.255.255.0 |
| Gateway | [string](#string) |  | e.g: 192.168.0.1 |






<a name="siemens.iedge.dmapi.network.v1.NetworkInterfaceRequest"></a>

### NetworkInterfaceRequest
Contains MAC address, used for retrieving specified Network Interface settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mac | [string](#string) |  |  |






<a name="siemens.iedge.dmapi.network.v1.NetworkInterfaceRequestWithLabel"></a>

### NetworkInterfaceRequestWithLabel
Returns an Network Interface with


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| label | [string](#string) |  |  |






<a name="siemens.iedge.dmapi.network.v1.NetworkSettings"></a>

### NetworkSettings
Contains multiple network interface settings. It can be used to apply or get the settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Interfaces | [Interface](#siemens.iedge.dmapi.network.v1.Interface) | repeated | Network settings contains an array of Interfaces.Applying new settings or receiving current settings is supported for multiple ethernet typed network interfaces supported. |
| LabelMap | [NetworkSettings.LabelMapEntry](#siemens.iedge.dmapi.network.v1.NetworkSettings.LabelMapEntry) | repeated | LabelMap contains port label and corresponding interface-name. e.g key : x1 value: enp2s0 |






<a name="siemens.iedge.dmapi.network.v1.NetworkSettings.LabelMapEntry"></a>

### NetworkSettings.LabelMapEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="siemens.iedge.dmapi.network.v1.NetworkService"></a>

### NetworkService
Network service ,uses a UNIX Domain Socket "/var/run/devicemodel/network.sock" for GRPC communication.
protoc  generates both client and server instance for this Service.
GRPC Status codes : https://developers.google.com/maps-booking/reference/grpc-api/status_codes .

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetAllInterfaces | [.google.protobuf.Empty](#google.protobuf.Empty) | [NetworkSettings](#siemens.iedge.dmapi.network.v1.NetworkSettings) | Returns the settings of all ethernet typed network interfaces |
| GetInterfaceWithMac | [NetworkInterfaceRequest](#siemens.iedge.dmapi.network.v1.NetworkInterfaceRequest) | [Interface](#siemens.iedge.dmapi.network.v1.Interface) | Returns the current setting for the interface, with given MAC address. |
| GetInterfaceWithLabel | [NetworkInterfaceRequestWithLabel](#siemens.iedge.dmapi.network.v1.NetworkInterfaceRequestWithLabel) | [Interface](#siemens.iedge.dmapi.network.v1.Interface) | Returns the current setting for the interface, with given Label. |
| ApplySettings | [NetworkSettings](#siemens.iedge.dmapi.network.v1.NetworkSettings) | [.google.protobuf.Empty](#google.protobuf.Empty) | Applies given configurations to Network Interfaces. |

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
