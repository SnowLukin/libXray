//go:build ios

package memory

import (
	"runtime/debug"
	"sync"
	"time"
)

const (
	defaultInterval = time.Second
	// 30M
	maxMemory = 30 * 1024 * 1024
)

var (
	initForceFreeOnce sync.Once

	loopMu       sync.Mutex
	loopInterval = defaultInterval
	loopStarted  bool
	stopLoop     chan struct{}

	freeOSMemory = debug.FreeOSMemory
)

func startLoopLocked() {
	if !loopStarted || loopInterval <= 0 {
		return
	}
	stop := make(chan struct{})
	stopLoop = stop
	// The pause counts from the end of the previous collection, so a slow
	// collection never makes them run back to back.
	go func(interval time.Duration, free func()) {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-stop:
				return
			case <-timer.C:
				free()
				timer.Reset(interval)
			}
		}
	}(loopInterval, freeOSMemory)
}

func stopLoopLocked() {
	if stopLoop != nil {
		close(stopLoop)
		stopLoop = nil
	}
}

// SetForceFreeInterval sets how often the process returns memory to the OS
// with debug.FreeOSMemory; zero or a negative interval stops the loop. Called
// before the first xray start, it sets the interval the loop starts with.
func SetForceFreeInterval(interval time.Duration) {
	loopMu.Lock()
	defer loopMu.Unlock()
	if interval == loopInterval {
		return
	}
	loopInterval = interval
	stopLoopLocked()
	startLoopLocked()
}

func InitForceFree() {
	initForceFreeOnce.Do(func() {
		debug.SetGCPercent(10)
		debug.SetMemoryLimit(maxMemory)
		loopMu.Lock()
		defer loopMu.Unlock()
		loopStarted = true
		startLoopLocked()
	})
}
