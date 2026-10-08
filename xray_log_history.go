package libXray

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/xtls/libxray/internal/loghistory"
	"github.com/xtls/libxray/nodep"
)

const maxLogHistoryConfigBytes = 8192

var xrayLogHistory loghistory.Manager

func init() {
	if err := xrayLogHistory.Register(); err != nil {
		panic(err)
	}
}

// ConfigureLogHistory configures Xray JSONL diagnostic history from a base64 JSON request.
func ConfigureLogHistory(base64Text string) string {
	var response nodep.CallResponse[string]
	if base64Text == "" {
		return response.EncodeToBase64("", xrayLogHistory.Configure(nil))
	}
	if base64.StdEncoding.DecodedLen(len(base64Text)) > maxLogHistoryConfigBytes {
		return response.EncodeToBase64("", errors.New("log_history_config_too_large"))
	}
	raw, err := base64.StdEncoding.DecodeString(base64Text)
	if err != nil {
		return response.EncodeToBase64("", err)
	}
	if len(raw) > maxLogHistoryConfigBytes {
		return response.EncodeToBase64("", errors.New("log_history_config_too_large"))
	}
	var config loghistory.Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return response.EncodeToBase64("", err)
	}
	return response.EncodeToBase64("", xrayLogHistory.Configure(&config))
}

// GetLogHistoryState returns the writer state through the standard call-response ABI.
func GetLogHistoryState() string {
	var response nodep.CallResponse[*loghistory.State]
	return response.EncodeToBase64(xrayLogHistory.State(), nil)
}
