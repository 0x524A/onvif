package onvif

import (
	"testing"

	"github.com/ritj/onvif/xsd"
	"github.com/ritj/onvif/xsd/onvif"
)

func TestCapabilities_FixEndpointAddresses(t *testing.T) {
	// Create a Capabilities structure with localhost addresses
	caps := onvif.Capabilities{
		Analytics: onvif.AnalyticsCapabilities{
			XAddr: xsd.AnyURI("http://127.0.0.1/onvif/analytics"),
		},
		Device: onvif.DeviceCapabilities{
			XAddr: xsd.AnyURI("http://localhost/onvif/device_service"),
		},
		Events: onvif.EventCapabilities{
			XAddr: xsd.AnyURI("http://127.0.0.1:80/onvif/events"),
		},
		Imaging: onvif.ImagingCapabilities{
			XAddr: xsd.AnyURI("http://localhost:80/onvif/imaging"),
		},
		Media: onvif.MediaCapabilities{
			XAddr: xsd.AnyURI("http://127.0.0.1/onvif/media"),
		},
		PTZ: onvif.PTZCapabilities{
			XAddr: xsd.AnyURI("http://127.0.0.1/onvif/ptz"),
		},
	}

	// Fix the addresses
	actualCameraIP := "192.168.1.164:80"
	caps.FixEndpointAddresses(actualCameraIP)

	// Verify all addresses were fixed
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Analytics", string(caps.Analytics.XAddr), "http://192.168.1.164:80/onvif/analytics"},
		{"Device", string(caps.Device.XAddr), "http://192.168.1.164:80/onvif/device_service"},
		{"Events", string(caps.Events.XAddr), "http://192.168.1.164:80/onvif/events"},
		{"Imaging", string(caps.Imaging.XAddr), "http://192.168.1.164:80/onvif/imaging"},
		{"Media", string(caps.Media.XAddr), "http://192.168.1.164:80/onvif/media"},
		{"PTZ", string(caps.PTZ.XAddr), "http://192.168.1.164:80/onvif/ptz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %s, expected %s", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestDevice_FixEndpointAddress(t *testing.T) {
	dev := &Device{
		params: DeviceParams{
			Xaddr: "192.168.1.164:80",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Localhost without port",
			input:    "http://127.0.0.1/onvif/services",
			expected: "http://192.168.1.164:80/onvif/services",
		},
		{
			name:     "Localhost with port",
			input:    "http://127.0.0.1:80/onvif/services",
			expected: "http://192.168.1.164:80/onvif/services",
		},
		{
			name:     "Localhost name",
			input:    "http://localhost/onvif/media",
			expected: "http://192.168.1.164:80/onvif/media",
		},
		{
			name:     "Localhost name with port",
			input:    "http://localhost:8080/onvif/media",
			expected: "http://192.168.1.164:80/onvif/media",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Different IP - should NOT be replaced",
			input:    "http://10.0.0.1/onvif/ptz",
			expected: "http://10.0.0.1/onvif/ptz",
		},
		{
			name:     "Different IP with port - should NOT be replaced",
			input:    "http://10.0.0.1:8080/onvif/ptz",
			expected: "http://10.0.0.1:8080/onvif/ptz",
		},
		{
			name:     "Valid camera IP - should NOT be replaced",
			input:    "http://192.168.1.100/onvif/services",
			expected: "http://192.168.1.100/onvif/services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dev.FixEndpointAddress(tt.input)
			if result != tt.expected {
				t.Errorf("FixEndpointAddress() = %v, want %v", result, tt.expected)
			}
		})
	}
}
