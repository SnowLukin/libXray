package loghistory

import (
	"errors"
	"path/filepath"
	"sync"

	applog "github.com/xtls/xray-core/app/log"
	commonlog "github.com/xtls/xray-core/common/log"
)

type Manager struct {
	mu             sync.Mutex
	config         *Config
	consoleFactory func() closableHandler
	// A prepared VPN config must not reopen a plain file after collection is revoked.
	markerPath string
	handler    *Handler
	failure    *State
}

type closableHandler interface {
	commonlog.Handler
	Close() error
}

type teeHandler struct {
	console closableHandler
	history *Handler
}

func (h *teeHandler) Handle(message commonlog.Message) {
	h.console.Handle(message)
	if h.history != nil {
		h.history.Handle(message)
	}
}

func (h *teeHandler) Close() error {
	err := error(nil)
	if h.history != nil {
		err = h.history.Close()
	}
	if consoleErr := h.console.Close(); err == nil {
		err = consoleErr
	}
	return err
}

func (m *Manager) Register() error {
	return applog.RegisterHandlerCreator(applog.LogType_File, m.create)
}

func (m *Manager) Configure(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if config == nil {
		m.config = nil
		if m.handler != nil {
			_ = m.handler.Close()
		}
		return nil
	}
	copy := *config
	if copy.Policy == (Policy{}) {
		copy.Policy = DefaultPolicy()
	}
	if err := copy.Policy.Validate(); err != nil {
		return err
	}
	if !filepath.IsAbs(copy.Directory) || len(copy.ConnectionID) > 128 {
		return errors.New("invalid_log_history_config")
	}
	if m.config != nil && *m.config == copy {
		return nil
	}
	if m.handler != nil && !m.handler.finished() {
		return errors.New("log_history_busy")
	}
	m.config = &copy
	m.markerPath = filepath.Join(copy.Directory, "xray-core.log")

	m.handler = nil
	m.failure = nil
	return nil
}

func (m *Manager) State() *State {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.handler != nil {
		state := m.handler.State()
		return &state
	}
	if m.failure != nil {
		state := *m.failure
		return &state
	}
	return nil
}

func (m *Manager) create(_ applog.LogType, options applog.HandlerCreatorOptions) (commonlog.Handler, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if filepath.Clean(options.Path) != m.markerPath {
		creator, err := commonlog.CreateFileLogWriter(options.Path)
		if err != nil {
			return nil, err
		}
		return commonlog.NewLogger(creator), nil
	}
	console := m.newConsoleHandler()
	if m.config == nil {
		return &teeHandler{console: console}, nil
	}

	if m.handler != nil && !m.handler.finished() {
		return &teeHandler{console: console}, nil
	}
	writer, err := New(*m.config)
	if err != nil {
		m.failure = &State{Version: 1, Status: "failed", ConnectionID: connectionID(m.config.ConnectionID), ErrorCode: SafeErrorCode(err)}
		return &teeHandler{console: console}, nil
	}
	m.handler = newHandler(*m.config, writer)
	return &teeHandler{console: console, history: m.handler}, nil
}

func (m *Manager) newConsoleHandler() closableHandler {
	if m.consoleFactory != nil {
		return m.consoleFactory()
	}
	return commonlog.NewLogger(commonlog.CreateStdoutLogWriter()).(closableHandler)
}
