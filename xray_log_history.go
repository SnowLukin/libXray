package libXray

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/xtls/libxray/internal/loghistory"
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
	if base64Text == "" {
		return encodeBase64CallResponse(nil, xrayLogHistory.Configure(nil))
	}
	if base64.StdEncoding.DecodedLen(len(base64Text)) > maxLogHistoryConfigBytes {
		return encodeBase64CallResponse(nil, errors.New("log_history_config_too_large"))
	}
	raw, err := base64.StdEncoding.DecodeString(base64Text)
	if err != nil {
		return encodeBase64CallResponse(nil, err)
	}
	if len(raw) > maxLogHistoryConfigBytes {
		return encodeBase64CallResponse(nil, errors.New("log_history_config_too_large"))
	}
	var config loghistory.Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return encodeBase64CallResponse(nil, err)
	}
	return encodeBase64CallResponse(nil, xrayLogHistory.Configure(&config))
}

// GetLogHistoryState returns the writer state through the standard call-response ABI.
func GetLogHistoryState() string {
	return encodeBase64CallResponse(xrayLogHistory.State(), nil)
}
