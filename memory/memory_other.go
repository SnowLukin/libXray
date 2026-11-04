//go:build !ios

package memory

import "time"

func InitForceFree() {}

func SetForceFreeInterval(time.Duration) {}
