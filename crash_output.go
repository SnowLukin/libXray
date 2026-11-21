package libXray

import (
	"errors"
	"os"
	"runtime/debug"
)

// SetCrashOutput makes the Go runtime copy its fatal report to the file at
// path, truncating it. On Android the runtime's stderr goes nowhere, so this
// file is the only place where the text of a panic or fatal error survives
// the process.
func SetCrashOutput(path string) error {
	if path == "" {
		return errors.New("crash output path is empty")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	// SetCrashOutput keeps its own duplicate of the descriptor.
	defer file.Close()
	if err := debug.SetCrashOutput(file, debug.CrashOptions{}); err != nil {
		return err
	}
	// Android starts app processes with AT_SECURE set, and in that mode the
	// runtime forces GOTRACEBACK=none: a panic would leave only its first
	// line, without the goroutine that panicked.
	debug.SetTraceback("single")
	return nil
}
