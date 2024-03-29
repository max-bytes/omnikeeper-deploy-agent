package healthcheck

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/max-bytes/omnikeeper-deploy-agent/v2/pkg/config"
)

var (
	statFilename = "/tmp/healthcheck_stat"
)

func Check(configFile string) {
	var cfg = config.Configuration{}
	err := config.ReadConfigFromFilename(configFile, &cfg)
	if err != nil {
		fmt.Printf("Error opening config file: %s\n", err)
		os.Exit(1)
	}

	file, err := os.Stat(statFilename)
	if err != nil {
		fmt.Printf("Error reading stats file: %s\n", err)
		os.Exit(1)
	}
	modifiedtime := file.ModTime()

	isTooOld := time.Now().Sub(modifiedtime) > time.Duration(cfg.HealthcheckThresholdSeconds*int64(time.Second))
	if isTooOld {
		fmt.Println("stats file too old")
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func TouchStatFile() {
	touchFile(statFilename)
}

func touchFile(fileName string) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(fileName, currentTime, currentTime)
		if err != nil {
			fmt.Println(err)
		}
	}
}
