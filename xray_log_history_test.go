package libXray

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/xtls/libxray/nodep"
)

type logHistoryStateResponse struct {
	Success bool `json:"success"`
	Data    *struct {
		Status string `json:"status"`
	} `json:"data"`
	Error string `json:"error"`
}

func TestConfigureLogHistoryUsesCallResponseABI(t *testing.T) {
	t.Cleanup(func() { _ = ConfigureLogHistory("") })
	config, err := json.Marshal(map[string]any{
		"directory":    filepath.Join(t.TempDir(), "history"),
		"connectionId": "connection-a",
	})
	if err != nil {
		t.Fatal(err)
	}

	var configured nodep.CallResponse[string]
	decodeCallResponse(t, ConfigureLogHistory(base64.StdEncoding.EncodeToString(config)), &configured)
	if !configured.Success {
		t.Fatalf("configure failed: %s", configured.Err)
	}

	var state logHistoryStateResponse
	decodeCallResponse(t, GetLogHistoryState(), &state)
	if !state.Success || state.Data != nil {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestConfigureLogHistoryRejectsInvalidBase64(t *testing.T) {
	var response nodep.CallResponse[string]
	decodeCallResponse(t, ConfigureLogHistory("%%%"), &response)
	if response.Success || response.Err == "" {
		t.Fatalf("invalid request was accepted: %#v", response)
	}
}

func decodeCallResponse(t *testing.T, encoded string, response any) {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, response); err != nil {
		t.Fatal(err)
	}
}
