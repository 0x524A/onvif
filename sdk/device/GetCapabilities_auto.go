// Code generated : DO NOT EDIT.
// Copyright (c) 2022 Jean-Francois SMIGIELSKI
// Distributed under the MIT License

package device

import (
	"context"
	"github.com/juju/errors"
	"github.com/ritj/onvif"
	"github.com/ritj/onvif/sdk"
	"github.com/ritj/onvif/device"
)

// Call_GetCapabilities forwards the call to dev.CallMethod() then parses the payload of the reply as a GetCapabilitiesResponse.
func Call_GetCapabilities(ctx context.Context, dev *onvif.Device, request device.GetCapabilities) (device.GetCapabilitiesResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			GetCapabilitiesResponse device.GetCapabilitiesResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.GetCapabilitiesResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "GetCapabilities")
		if err == nil {
			// Fix localhost addresses in the capabilities response
			deviceParams := dev.GetDeviceParams()
			reply.Body.GetCapabilitiesResponse.Capabilities.FixEndpointAddresses(deviceParams.Xaddr)
		}
		return reply.Body.GetCapabilitiesResponse, errors.Annotate(err, "reply")
	}
}
