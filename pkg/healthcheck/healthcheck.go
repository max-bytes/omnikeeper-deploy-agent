package healthcheck

import (
	"os"
	"time"
)

var lastSuccess time.Time = time.Now()

func Check(threshold time.Duration) {
	isTooOld := time.Now().Sub(lastSuccess) > threshold
	if isTooOld {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func SetLastSuccess() {
	lastSuccess = time.Now()
}
