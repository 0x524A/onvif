package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	goonvif "github.com/0x524a/onvif"
	"github.com/0x524a/onvif/device"
	sdk "github.com/0x524a/onvif/sdk/device"
)

const (
	// Camera credentials
	username = "admin"
	password = "password"
	// Camera IP address and port
	// Use the actual camera IP (e.g., 192.168.1.164:80), not localhost
	cameraIP = "192.168.1.164:80"
)

func main() {
	ctx := context.Background()

	// Create a new ONVIF device connection
	// IMPORTANT: Use the actual camera IP address, not 127.0.0.1 or localhost
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{
		Xaddr:      cameraIP,
		Username:   username,
		Password:   password,
		HttpClient: new(http.Client),
	})
	if err != nil {
		log.Fatalf("Failed to create device: %v", err)
	}

	fmt.Println("Successfully connected to camera at:", cameraIP)
	fmt.Println()

	// Get device capabilities
	// Even if the camera returns localhost addresses in its response,
	// the library will automatically replace them with the actual camera IP
	getCapabilities := device.GetCapabilities{Category: "All"}
	capabilitiesResponse, err := sdk.Call_GetCapabilities(ctx, dev, getCapabilities)
	if err != nil {
		log.Fatalf("Failed to get capabilities: %v", err)
	}

	// Display the service endpoints
	// These will all use the correct camera IP, not localhost
	fmt.Println("Device Service Endpoints:")
	services := dev.GetServices()
	for name, endpoint := range services {
		fmt.Printf("  %s: %s\n", name, endpoint)
	}
	fmt.Println()

	// Display capability information
	fmt.Println("Device Capabilities:")
	if capabilitiesResponse.Capabilities.Device.XAddr != "" {
		fmt.Printf("  Device Service: %s\n", capabilitiesResponse.Capabilities.Device.XAddr)
	}
	if capabilitiesResponse.Capabilities.Media.XAddr != "" {
		fmt.Printf("  Media Service: %s\n", capabilitiesResponse.Capabilities.Media.XAddr)
	}
	if capabilitiesResponse.Capabilities.PTZ.XAddr != "" {
		fmt.Printf("  PTZ Service: %s\n", capabilitiesResponse.Capabilities.PTZ.XAddr)
	}
	if capabilitiesResponse.Capabilities.Events.XAddr != "" {
		fmt.Printf("  Events Service: %s\n", capabilitiesResponse.Capabilities.Events.XAddr)
	}
	if capabilitiesResponse.Capabilities.Imaging.XAddr != "" {
		fmt.Printf("  Imaging Service: %s\n", capabilitiesResponse.Capabilities.Imaging.XAddr)
	}

	fmt.Println()
	fmt.Println("Note: All endpoints have been automatically corrected to use the actual camera IP,")
	fmt.Println("even if the camera originally returned localhost (127.0.0.1) addresses.")
}
