//go:build ios

package memory

import (
	"math"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func countFrees(t *testing.T) *atomic.Int64 {
	t.Helper()
	var frees atomic.Int64
	reset := func() {
		loopMu.Lock()
		stopLoopLocked()
		loopInterval = defaultInterval
		loopStarted = false
		loopMu.Unlock()
		initForceFreeOnce = sync.Once{}
	}
	reset()
	freeOSMemory = func() { frees.Add(1) }
	t.Cleanup(func() {
		reset()
		freeOSMemory = debug.FreeOSMemory
		debug.SetGCPercent(100)
		debug.SetMemoryLimit(math.MaxInt64)
	})
	return &frees
}

func waitForFrees(t *testing.T, frees *atomic.Int64, want int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for frees.Load() < want {
		if time.Now().After(deadline) {
			t.Fatalf("got %d frees, want at least %d", frees.Load(), want)
		}
		time.Sleep(time.Millisecond)
	}
}

func assertNoMoreFrees(t *testing.T, frees *atomic.Int64) {
	t.Helper()
	time.Sleep(20 * time.Millisecond)
	before := frees.Load()
	time.Sleep(50 * time.Millisecond)
	if after := frees.Load(); after != before {
		t.Fatalf("loop kept running: %d frees, then %d", before, after)
	}
}

func TestInitForceFreeRunsLoopWithInterval(t *testing.T) {
	frees := countFrees(t)
	SetForceFreeInterval(5 * time.Millisecond)

	InitForceFree()

	waitForFrees(t, frees, 3)
}

func TestSetForceFreeIntervalZeroStopsLoop(t *testing.T) {
	frees := countFrees(t)
	SetForceFreeInterval(5 * time.Millisecond)
	InitForceFree()
	waitForFrees(t, frees, 1)

	SetForceFreeInterval(0)

	assertNoMoreFrees(t, frees)
}

func TestZeroIntervalBeforeStartKeepsLoopOff(t *testing.T) {
	frees := countFrees(t)
	SetForceFreeInterval(0)

	InitForceFree()

	assertNoMoreFrees(t, frees)
	SetForceFreeInterval(5 * time.Millisecond)
	waitForFrees(t, frees, 1)
}

func TestNewIntervalReplacesRunningLoop(t *testing.T) {
	frees := countFrees(t)
	SetForceFreeInterval(5 * time.Millisecond)
	InitForceFree()
	waitForFrees(t, frees, 1)

	SetForceFreeInterval(time.Hour)

	assertNoMoreFrees(t, frees)
}

func TestInitForceFreeKeepsLibXrayCollectorSettings(t *testing.T) {
	countFrees(t)

	InitForceFree()

	gcPercent := debug.SetGCPercent(-1)
	debug.SetGCPercent(gcPercent)
	if limit := debug.SetMemoryLimit(-1); gcPercent != 10 || limit != maxMemory {
		t.Fatalf("got GOGC %d, limit %d", gcPercent, limit)
	}
}
