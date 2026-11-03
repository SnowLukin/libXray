package libXray

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/xtls/libxray/xray"
	"github.com/xtls/xray-core/common/platform"
)

// The functions in this file keep the base64 request/response ABI used by the
// BlancVPN Android host. Invoke is the upstream entry point; these wrappers only
// translate the old request shape and do not add behavior of their own.

type base64CallResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Err     string `json:"error,omitempty"`
}

func encodeBase64CallResponse(data any, err error) string {
	response := base64CallResponse{Data: data, Success: err == nil}
	if err != nil {
		response.Err = err.Error()
	}
	raw, err := json.Marshal(&response)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// RunXrayFromJSONRequest is the decoded request of RunXrayFromJSON.
type RunXrayFromJSONRequest struct {
	DatDir     string `json:"datDir,omitempty"`
	ConfigJSON string `json:"configJSON,omitempty"`
}

// RunXrayFromJSON starts Xray from a base64 JSON RunXrayFromJSONRequest.
// DatDir is where geoip.dat and geosite.dat are looked up. The call returns
// after inbounds are bound, so start errors are part of the response.
func RunXrayFromJSON(base64Text string) string {
	raw, err := base64.StdEncoding.DecodeString(base64Text)
	if err != nil {
		return encodeBase64CallResponse(nil, err)
	}
	var request RunXrayFromJSONRequest
	if err := json.Unmarshal(raw, &request); err != nil {
		return encodeBase64CallResponse(nil, err)
	}
	if request.DatDir != "" {
		os.Setenv(platform.AssetLocation, request.DatDir)
		os.Setenv(platform.CertLocation, request.DatDir)
	}
	return encodeBase64CallResponse(nil, xray.RunXray(request.ConfigJSON))
}

// StopXray stops the running instance and returns a base64 call response.
func StopXray() string {
	return encodeBase64CallResponse(nil, xray.StopXray())
}

// GetXrayState reports whether an Xray instance is running.
func GetXrayState() bool {
	return xray.GetXrayState()
}
