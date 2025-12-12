package onvif

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/0x524a/onvif/device"
	"github.com/0x524a/onvif/gosoap"
	"github.com/0x524a/onvif/networking"
	wsdiscovery "github.com/0x524a/onvif/ws-discovery"
	xsdonvif "github.com/0x524a/onvif/xsd/onvif"
)

// Xlmns XML Schema
var Xlmns = map[string]string{
	"onvif":   "http://www.onvif.org/ver10/schema",
	"tds":     "http://www.onvif.org/ver10/device/wsdl",
	"trt":     "http://www.onvif.org/ver10/media/wsdl",
	"tev":     "http://www.onvif.org/ver10/events/wsdl",
	"tptz":    "http://www.onvif.org/ver20/ptz/wsdl",
	"timg":    "http://www.onvif.org/ver20/imaging/wsdl",
	"tan":     "http://www.onvif.org/ver20/analytics/wsdl",
	"xmime":   "http://www.w3.org/2005/05/xmlmime",
	"wsnt":    "http://docs.oasis-open.org/wsn/b-2",
	"xop":     "http://www.w3.org/2004/08/xop/include",
	"wsa":     "http://www.w3.org/2005/08/addressing",
	"wstop":   "http://docs.oasis-open.org/wsn/t-1",
	"wsntw":   "http://docs.oasis-open.org/wsn/bw-2",
	"wsrf-rw": "http://docs.oasis-open.org/wsrf/rw-2",
	"wsaw":    "http://www.w3.org/2006/05/addressing/wsdl",
}

// DeviceType alias for int
type DeviceType int

// Onvif Device Tyoe
const (
	NVD DeviceType = iota
	NVS
	NVA
	NVT
)

func (devType DeviceType) String() string {
	stringRepresentation := []string{
		"NetworkVideoDisplay",
		"NetworkVideoStorage",
		"NetworkVideoAnalytics",
		"NetworkVideoTransmitter",
	}
	i := uint8(devType)
	switch {
	case i <= uint8(NVT):
		return stringRepresentation[i]
	default:
		return strconv.Itoa(int(i))
	}
}

// DeviceInfo struct contains general information about ONVIF device
type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

// Device for a new device of onvif and DeviceInfo
// struct represents an abstract ONVIF device.
// It contains methods, which helps to communicate with ONVIF device
type Device struct {
	params    DeviceParams
	endpoints map[string]string
	info      DeviceInfo
}

type DeviceParams struct {
	Xaddr      string
	Username   string
	Password   string
	HttpClient *http.Client
}

// GetServices return available endpoints
func (dev *Device) GetServices() map[string]string {
	return dev.endpoints
}

// GetServices return available endpoints
func (dev *Device) GetDeviceInfo() DeviceInfo {
	return dev.info
}

// GetDeviceParams return available endpoints
func (dev *Device) GetDeviceParams() DeviceParams {
	return dev.params
}

// FixXAddr fixes localhost/127.0.0.1 addresses in XAddr fields
// This is a helper to ensure consistent URI fixing across all responses
func (dev *Device) FixXAddr(addr string) string {
	return dev.FixEndpointAddress(addr)
}

// FixMediaUriResponse fixes localhost/127.0.0.1 addresses in MediaUri responses
// This should be called on GetStreamUri and GetSnapshotUri responses
func (dev *Device) FixMediaUriResponse(mediaUri *xsdonvif.MediaUri) {
	if mediaUri != nil {
		mediaUri.FixMediaUri(dev.params.Xaddr)
	}
}

func readResponse(resp *http.Response) string {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

//GetAvailableDevicesAtSpecificEthernetInterface ...
func GetAvailableDevicesAtSpecificEthernetInterface(interfaceName string) ([]Device, error) {
	// Call a ws-discovery Probe Message to Discover NVT type Devices
	devices, err := wsdiscovery.SendProbe(interfaceName, nil, []string{"dn:" + NVT.String()}, map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})
	if err != nil {
		return nil, err
	}

	nvtDevicesSeen := make(map[string]bool)
	nvtDevices := make([]Device, 0)

	for _, j := range devices {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			return nil, err
		}

		for _, xaddr := range doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch/XAddrs") {
			xaddr := strings.Split(strings.Split(xaddr.Text(), " ")[0], "/")[2]
			if !nvtDevicesSeen[xaddr] {
				dev, err := NewDevice(DeviceParams{Xaddr: strings.Split(xaddr, " ")[0]})
				if err != nil {
					// TODO(jfsmig) print a warning
				} else {
					nvtDevicesSeen[xaddr] = true
					nvtDevices = append(nvtDevices, *dev)
				}
			}
		}
	}

	return nvtDevices, nil
}

func (dev *Device) getSupportedServices(resp *http.Response) error {
	doc := etree.NewDocument()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	resp.Body.Close()

	if err := doc.ReadFromBytes(data); err != nil {
		//log.Println(err.Error())
		return err
	}

	services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/*/XAddr")
	for _, j := range services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	extension_services := doc.FindElements("./Envelope/Body/GetCapabilitiesResponse/Capabilities/Extension/*/XAddr")
	for _, j := range extension_services {
		dev.addEndpoint(j.Parent().Tag, j.Text())
	}

	return nil
}

// NewDevice function construct a ONVIF Device entity
func NewDevice(params DeviceParams) (*Device, error) {
	dev := new(Device)
	dev.params = params
	dev.endpoints = make(map[string]string)
	dev.addEndpoint("Device", "http://"+dev.params.Xaddr+"/onvif/device_service")

	if dev.params.HttpClient == nil {
		dev.params.HttpClient = new(http.Client)
	}

	getCapabilities := device.GetCapabilities{Category: "All"}

	resp, err := dev.CallMethod(getCapabilities)

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("camera is not available at " + dev.params.Xaddr + " or it does not support ONVIF services")
	}

	err = dev.getSupportedServices(resp)
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (dev *Device) addEndpoint(Key, Value string) {
	//use lowCaseKey
	//make key having ability to handle Mixed Case for Different vendor devcie (e.g. Events EVENTS, events)
	lowCaseKey := strings.ToLower(Key)

	// Replace host with host from device params only if it's localhost or empty.
	if u, err := url.Parse(Value); err == nil {
		if isLocalhostOrEmpty(u.Host) {
			u.Host = dev.params.Xaddr
			Value = u.String()
		}
	}

	dev.endpoints[lowCaseKey] = Value
}

// isLocalhostOrEmpty checks if a host is localhost, 127.0.0.1, or empty
func isLocalhostOrEmpty(host string) bool {
	if host == "" {
		return true
	}
	// Remove port if present
	hostname := host
	if idx := strings.Index(host, ":"); idx != -1 {
		hostname = host[:idx]
	}
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == ""
}

// FixEndpointAddress replaces the host in a URL with the device's actual address
// only if the host is localhost, 127.0.0.1, or empty.
// This is used to fix localhost addresses that cameras sometimes return.
func (dev *Device) FixEndpointAddress(address string) string {
	if address == "" {
		return address
	}
	if u, err := url.Parse(address); err == nil {
		if isLocalhostOrEmpty(u.Host) {
			u.Host = dev.params.Xaddr
			return u.String()
		}
	}
	return address
}

// DebugDeviceEndpoints returns all cached endpoints with their status
// Returns a map where key is service name and value is endpoint URL with status
func (dev *Device) DebugDeviceEndpoints() map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for service, endpoint := range dev.endpoints {
		status := "OK"
		isLocalhost := false
		if u, err := url.Parse(endpoint); err == nil {
			if isLocalhostOrEmpty(u.Host) {
				status = "LOCALHOST"
				isLocalhost = true
			}
		} else {
			status = "PARSE_ERROR"
		}
		result[service] = map[string]interface{}{
			"endpoint":    endpoint,
			"status":      status,
			"is_localhost": isLocalhost,
		}
	}
	return result
}

// GetEndpointForService returns the endpoint URL for a specific service name (for debugging)
// For example: GetEndpointForService("media") returns the Media service endpoint
func (dev *Device) GetEndpointForService(serviceName string) (string, error) {
	return dev.getEndpoint(serviceName)
}

// GetEndpoint returns specific ONVIF service endpoint address
func (dev *Device) GetEndpoint(name string) string {
	return dev.endpoints[name]
}

func (dev Device) buildMethodSOAP(msg string) (gosoap.SoapMessage, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(msg); err != nil {
		//log.Println("Got error")

		return "", err
	}
	element := doc.Root()

	soap := gosoap.NewEmptySOAP()
	soap.AddBodyContent(element)

	return soap, nil
}

// getEndpoint functions get the target service endpoint in a better way
func (dev Device) getEndpoint(endpoint string) (string, error) {

	// common condition, endpointMark in map we use this.
	if endpointURL, bFound := dev.endpoints[endpoint]; bFound {
		return endpointURL, nil
	}

	//but ,if we have endpoint like event、analytic
	//and sametime the Targetkey like : events、analytics
	//we use fuzzy way to find the best match url
	var endpointURL string
	for targetKey := range dev.endpoints {
		if strings.Contains(targetKey, endpoint) {
			endpointURL = dev.endpoints[targetKey]
			return endpointURL, nil
		}
	}
	return endpointURL, errors.New("target endpoint service not found")
}

// CallMethod functions call an method, defined <method> struct.
// You should use Authenticate method to call authorized requests.
func (dev Device) CallMethod(method interface{}) (*http.Response, error) {
	pkgPath := strings.Split(reflect.TypeOf(method).PkgPath(), "/")
	pkg := strings.ToLower(pkgPath[len(pkgPath)-1])

	endpoint, err := dev.getEndpoint(pkg)
	if err != nil {
		return nil, err
	}
	
	// Validate endpoint before calling
	if u, parseErr := url.Parse(endpoint); parseErr == nil && isLocalhostOrEmpty(u.Host) {
		return nil, errors.New("endpoint for service '" + pkg + "' has localhost: " + endpoint + 
			". Use DebugDeviceEndpoints() to see all cached endpoints")
	}
	
	return dev.callMethodDo(endpoint, method)
}

// GetLastEndpointUsed returns the last endpoint that would be used for a given service (for debugging)
func (dev *Device) GetLastEndpointUsed(serviceName string) (string, error) {
	return dev.getEndpoint(serviceName)
}

// CallMethod functions call an method, defined <method> struct with authentication data
func (dev Device) callMethodDo(endpoint string, method interface{}) (*http.Response, error) {
	output, err := xml.MarshalIndent(method, "  ", "    ")
	if err != nil {
		return nil, err
	}

	soap, err := dev.buildMethodSOAP(string(output))
	if err != nil {
		return nil, err
	}

	soap.AddRootNamespaces(Xlmns)
	soap.AddAction()

	//Auth Handling
	if dev.params.Username != "" && dev.params.Password != "" {
		soap.AddWSSecurity(dev.params.Username, dev.params.Password)
	}

	return networking.SendSoap(dev.params.HttpClient, endpoint, soap.String())
}
