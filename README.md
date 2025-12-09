# onvif protocol

Simple management of onvif IP-devices cameras. onvif is an implementation of  ONVIF protocol for managing onvif IP devices. The purpose of this library is convenient and easy management of IP cameras and other devices that support ONVIF standard.

## Installation

To install the library,  use **go get**:

```go
go get github.com/0x524a/onvif

```

## Supported services

The following services are implemented:

- Device
- Media
- PTZ
- Imaging
- Event
- Discovery
- Auth(More Options)
- Soap

## Using

### General concept

1) Connecting to the device
2) Authentication (if necessary)
3) Defining Data Types
4) Carrying out the required method

#### Connecting to the device

If there is a device on the network at the address *192.168.13.42*, and its ONVIF services use the *1234* port, then you can connect to the device in the following way:

```go
dev, err := onvif.NewDevice(onvif.DeviceParams{Xaddr: "192.168.13.42:1234"})
```

*The ONVIF port may differ depending on the device , to find out which port to use, you can go to the web interface of the device. **Usually this is 80 port.***

#### Authentication

If any function of the ONVIF services requires authentication, you must use the `Authenticate` method.

```go
device := onvif.NewDevice(onvif.DeviceParams{Xaddr: "192.168.13.42:1234", Username: "username", Password: password})
```

#### Defining Data Types

Each ONVIF service in this library has its own package, in which all data types of this service are defined, and the package name is identical to the service name and begins with a capital letter. onvif defines the structures for each function of each ONVIF service supported by this library. Define the data type of the `GetCapabilities` function of the Device service. This is done as follows:

```go
capabilities := device.GetCapabilities{Category:"All"}
```

Why does the `GetCapabilities` structure have the Category field and why is the value of this field `All`?

The figure below shows the documentation for the [GetCapabilities](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl). It can be seen that the function takes one Category parameter and its value should be one of the following: 'All', 'Analytics',' Device ',' Events', 'Imaging', 'Media' or 'PTZ'`.

![Device GetCapabilities](docs/img/exmp_GetCapabilities.png)

An example of defining the data type of `GetServiceCapabilities` function in [PTZ](https://www.onvif.org/ver20/ptz/wsdl/ptz.wsdl):

```go
ptzCapabilities := ptz.GetServiceCapabilities{}
```

The figure below shows that `GetServiceCapabilities` does not accept any arguments.

![PTZ GetServiceCapabilities](docs/img/GetServiceCapabilities.png)

*Common data types are in the xsd/onvif package. The types of data (structures) that can be shared by all services are defined in the onvif package.*

An example of how to define the data type of the CreateUsers function in [Devicemgmt](https://www.onvif.org/ver10/device/wsdl/devicemgmt.wsdl):

```go
createUsers := device.CreateUsers{User: onvif.User{Username:"admin", Password:"qwerty", UserLevel:"User"}}
```

The figure below shows that ,in this example, the `CreateUsers` structure field must be a User whose data type is the User structure containing the Username, Password, UserLevel, and optional Extension fields. The User structure is in the onvif package.

![Device CreateUsers](docs/img/exmp_CreateUsers.png)

#### Carrying out the required method

To perform any function of one of the ONVIF services whose structure has been defined, you must use the `CallMethod` of the device object.

```go
createUsers := device.CreateUsers{User: onvif.User{Username:"admin", Password:"qwerty", UserLevel:"User"}}
device := onvif.NewDevice(onvif.DeviceParams{Xaddr: "192.168.13.42:1234", Username: "username", Password: password})
device.Authenticate("username", "password")
resp, err := dev.CallMethod(createUsers)
```

## Automatic Localhost Address Fixing

Some ONVIF cameras incorrectly return `localhost` (127.0.0.1) instead of their actual IP addresses in responses. This library automatically detects and fixes these addresses at multiple levels:

### Architecture

The library implements a **multi-layer approach** to ensure all addresses are correct:

#### 1. Endpoint Caching (Initialization)
When you create a new device with `NewDevice()`, the library:
- Calls `GetCapabilities` to discover all available services
- Caches each service endpoint (Media, PTZ, Events, etc.)
- **Automatically fixes localhost/127.0.0.1** in cached endpoints

This happens in `addEndpoint()` - so your cached endpoints are always correct.

#### 2. Capabilities Response Fixing
When you call `GetCapabilities` directly:
- The response contains `XAddr` fields for each service
- **Automatically fixed** before the response is returned to you
- Ensures consistency between cached endpoints and response data

#### 3. Dynamic URI Responses
For operations that return dynamic URIs (not cached):
- `GetStreamUri` - Returns RTSP/HTTP stream URLs
- `GetSnapshotUri` - Returns snapshot image URLs  
- **Automatically fixed** before responses are returned

### Conservative Approach

The fix only replaces addresses when the camera returns:
- `localhost`
- `127.0.0.1`
- Empty/missing host

**If the camera returns a different IP address, it is preserved as-is.** This ensures cameras configured to use specific network interfaces, proxy servers, or multi-homed setups continue to work correctly.

### What Gets Fixed

| Response Type | Field | Auto-Fixed | Cached |
|--------------|-------|------------|--------|
| GetCapabilities | Analytics.XAddr | ✓ | ✓ |
| GetCapabilities | Device.XAddr | ✓ | ✓ |
| GetCapabilities | Events.XAddr | ✓ | ✓ |
| GetCapabilities | Imaging.XAddr | ✓ | ✓ |
| GetCapabilities | Media.XAddr | ✓ | ✓ |
| GetCapabilities | PTZ.XAddr | ✓ | ✓ |
| GetStreamUri | MediaUri.Uri | ✓ | ✗ |
| GetSnapshotUri | MediaUri.Uri | ✓ | ✗ |

### Manual Fixing (Advanced)

If you need to manually fix addresses:

```go
// Fix a single endpoint address
fixedAddress := device.FixEndpointAddress("http://localhost/onvif/media")
// Returns: "http://192.168.13.42:1234/onvif/media"

// Fix an XAddr field  
fixedXAddr := device.FixXAddr("http://127.0.0.1/onvif/ptz")

// Fix a MediaUri struct
device.FixMediaUriResponse(&mediaUri)
```

### Implementation Details

The localhost detection logic:
```go
// Considered localhost:
"localhost"           -> replaced
"127.0.0.1"          -> replaced  
"localhost:554"      -> replaced
"127.0.0.1:8080"     -> replaced
""                   -> replaced (empty host)

// Preserved (not localhost):
"192.168.1.100"      -> preserved
"10.0.0.5:554"       -> preserved
"camera.local"       -> preserved
```

## Great Thanks

Enhanced and Improved from: [goonvif](https://github.com/yakovlevdmv/goonvif)
