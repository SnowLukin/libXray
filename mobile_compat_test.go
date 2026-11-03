package libXray

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/xtls/xray-core/common/platform"
)

func encodeRunRequest(t *testing.T, request RunXrayFromJSONRequest) string {
	t.Helper()
	raw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestRunXrayFromJSONStartsAndStopsCore(t *testing.T) {
	datDir := t.TempDir()
	t.Setenv(platform.AssetLocation, "")
	t.Setenv(platform.CertLocation, "")
	t.Cleanup(func() { _ = StopXray() })

	config := `{"outbounds":[{"protocol":"freedom"}]}`
	var started base64CallResponse
	decodeCallResponse(t, RunXrayFromJSON(encodeRunRequest(t, RunXrayFromJSONRequest{DatDir: datDir, ConfigJSON: config})), &started)
	if !started.Success {
		t.Fatalf("start failed: %s", started.Err)
	}
	if !GetXrayState() {
		t.Fatal("core is not running after a successful start")
	}
	if got := os.Getenv(platform.AssetLocation); got != datDir {
		t.Fatalf("asset location = %q, want %q", got, datDir)
	}

	var stopped base64CallResponse
	decodeCallResponse(t, StopXray(), &stopped)
	if !stopped.Success {
		t.Fatalf("stop failed: %s", stopped.Err)
	}
	if GetXrayState() {
		t.Fatal("core is still running after stop")
	}
}

func TestRunXrayFromJSONReportsStartError(t *testing.T) {
	t.Cleanup(func() { _ = StopXray() })
	var response base64CallResponse
	decodeCallResponse(t, RunXrayFromJSON(encodeRunRequest(t, RunXrayFromJSONRequest{ConfigJSON: `{"outbounds":[{"protocol":"no-such-protocol"}]}`})), &response)
	if response.Success || response.Err == "" {
		t.Fatalf("broken config was accepted: %#v", response)
	}
	if GetXrayState() {
		t.Fatal("core is running after a failed start")
	}
}

func TestRunXrayFromJSONRejectsInvalidBase64(t *testing.T) {
	var response base64CallResponse
	decodeCallResponse(t, RunXrayFromJSON("%%%"), &response)
	if response.Success || !strings.Contains(response.Err, "illegal base64") {
		t.Fatalf("invalid request was accepted: %#v", response)
	}
}
