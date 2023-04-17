# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [Onboard.proto](#Onboard.proto)
    - [Agent](#siemens.iedge.dmapi.onboard.v1.Agent)
    - [Device](#siemens.iedge.dmapi.onboard.v1.Device)
    - [DeviceConfiguration](#siemens.iedge.dmapi.onboard.v1.DeviceConfiguration)
    - [Onboarding](#siemens.iedge.dmapi.onboard.v1.Onboarding)
    - [OnboardingStatus](#siemens.iedge.dmapi.onboard.v1.OnboardingStatus)
    - [ProxySettings](#siemens.iedge.dmapi.onboard.v1.ProxySettings)
    - [Security](#siemens.iedge.dmapi.onboard.v1.Security)
  
    - [OnboardService](#siemens.iedge.dmapi.onboard.v1.OnboardService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="Onboard.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## Onboard.proto



<a name="siemens.iedge.dmapi.onboard.v1.Agent"></a>

### Agent
Agent section in IEM Onboarding JSON file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| proxy | [string](#string) | repeated |  |
| agentId | [string](#string) |  |  |
| security | [Security](#siemens.iedge.dmapi.onboard.v1.Security) |  |  |






<a name="siemens.iedge.dmapi.onboard.v1.Device"></a>

### Device
Security section in IEM Onboarding JSON file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Network | [siemens.iedge.dmapi.network.v1.NetworkSettings](#siemens.iedge.dmapi.network.v1.NetworkSettings) |  |  |
| ntpServer | [string](#string) | repeated |  |
| dockerIP | [string](#string) |  |  |
| customConfiguration | [google.protobuf.Any](#google.protobuf.Any) |  | device builder specific custom configuraiton. |






<a name="siemens.iedge.dmapi.onboard.v1.DeviceConfiguration"></a>

### DeviceConfiguration
Corresponding proto message for JSON file that had downloaded from IEM. Onboarding JSON file needs to be converted to this type first.
Compatible with IEM Onboarding JSON schema, since all field names are same.
For more information please refer to IEM Onboarding JSON scheme.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| Device | [Device](#siemens.iedge.dmapi.onboard.v1.Device) |  | Device seciton in IEM Onboarding JSON file. |
| onboarding | [Onboarding](#siemens.iedge.dmapi.onboard.v1.Onboarding) |  | Onboarding section in IEM Onboarding JSON file. |
| agents | [Agent](#siemens.iedge.dmapi.onboard.v1.Agent) | repeated | Agent section in IEM Onboarding JSON file. |
| proxies | [ProxySettings](#siemens.iedge.dmapi.onboard.v1.ProxySettings) | repeated | Proxy section in IEM Onboarding JSON file. |






<a name="siemens.iedge.dmapi.onboard.v1.Onboarding"></a>

### Onboarding
Onboarding section in IEM Onboarding JSON file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| localUserName | [string](#string) |  |  |
| localPassword | [string](#string) |  |  |
| deviceName | [string](#string) |  |  |
| deviceId | [string](#string) |  |  |
| platformActualName | [string](#string) |  |  |
| softwarePlatformName | [string](#string) |  |  |
| potalUrl | [string](#string) |  |  |
| nodeId | [string](#string) |  |  |
| userId | [string](#string) |  |  |
| deviceType | [string](#string) |  |  |
| swPlatformId | [string](#string) |  |  |
| platformId | [string](#string) |  |  |
| isActivationConfirmed | [bool](#bool) |  |  |
| deviceRole | [string](#string) |  |  |






<a name="siemens.iedge.dmapi.onboard.v1.OnboardingStatus"></a>

### OnboardingStatus
Type indicates the Onboarding Status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| isOnboarded | [bool](#bool) |  | true if onboarding is successful,otherwise false. |
| message | [string](#string) |  | Additonal information for onboarding status. |






<a name="siemens.iedge.dmapi.onboard.v1.ProxySettings"></a>

### ProxySettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proxyId | [string](#string) |  |  |
| proxyType | [string](#string) |  |  |
| host | [string](#string) |  |  |
| protocol | [string](#string) |  |  |
| authenticationType | [string](#string) |  |  |
| user | [string](#string) |  |  |
| password | [string](#string) |  |  |
| noProxy | [string](#string) |  |  |
| customPorts | [int32](#int32) | repeated |  |






<a name="siemens.iedge.dmapi.onboard.v1.Security"></a>

### Security
Security section in IEM Onboarding JSON file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| baseUrl | [string](#string) |  |  |
| clientCredentialProfile | [string](#string) | repeated |  |
| ca_chain | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="siemens.iedge.dmapi.onboard.v1.OnboardService"></a>

### OnboardService
Onboard service ,uses a UNIX Domain Socket "/var/run/devicemodel/onboard.sock" for GRPC communication.
protoc  generates both client and server instance for this Service.
GRPC Status codes : https://developers.google.com/maps-booking/reference/grpc-api/status_codes .

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| OnboardWithUSB | [DeviceConfiguration](#siemens.iedge.dmapi.onboard.v1.DeviceConfiguration) | [.google.protobuf.Empty](#google.protobuf.Empty) | Starts onboarding sequence,applies all settings via ApplyConfiguration() and finally calls EdgeCoreRuntime REST API to onboard the device. |
| ApplyConfiguration | [DeviceConfiguration](#siemens.iedge.dmapi.onboard.v1.DeviceConfiguration) | [.google.protobuf.Empty](#google.protobuf.Empty) | Applies all settings. Does not call EdgeCoreRuntime REST API.Applicable for UI Onboarding. |
| SetOnboardingStatus | [OnboardingStatus](#siemens.iedge.dmapi.onboard.v1.OnboardingStatus) | [.google.protobuf.Empty](#google.protobuf.Empty) | Sets state for device onboard status.In Onboarding from UI, EdgeCoreRuntime calls this method to set the onboarding status. |
| ListenOnboardState | [.google.protobuf.Empty](#google.protobuf.Empty) | [OnboardingStatus](#siemens.iedge.dmapi.onboard.v1.OnboardingStatus) stream | Server stream for onboarding status changes. |

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
