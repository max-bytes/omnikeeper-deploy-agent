package healthcheck

import (
	"flag"
	"os"
	"time"

	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/config"
)

var (
	configFile = flag.String("config", "config.yml", "Config file location")
)

var lastSuccess time.Time = time.Now()

func Check() {
	flag.Parse()

	var cfg = config.Configuration{}
	err := config.ReadConfigFromFilename(*configFile, &cfg)
	if err != nil {
		os.Exit(1)
	}

	isTooOld := time.Now().Sub(lastSuccess) > time.Duration(cfg.HealthcheckThresholdSeconds*int64(time.Second))
	if isTooOld {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func SetLastSuccess() {
	lastSuccess = time.Now()
}
