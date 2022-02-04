package main

import (
	"bytes"
	"context"
	"deploy-agent/pkg/config"
	"deploy-agent/pkg/omnikeeper"
	"deploy-agent/pkg/processors"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	version    = "0.0.0-src"
	configFile = flag.String("config", "config.json", "Config file location")
	log        logrus.Logger
)

var cfg = config.Configuration{}

func init() {
	log = *logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.TraceLevel) // is overwritten by configuration below
}

func main() {
	log.Infof("omnikeeper-deploy-agent (Version: %s)", version)
	flag.Parse()

	log.Infof("Loading config from file: %s", *configFile)
	err := config.ReadConfigFromFile(*configFile, &cfg)
	if err != nil {
		log.Fatalf("Error opening config file: %s", err)
	}

	parsedLogLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error parsing loglevel in config file: %s", err)
	}
	log.SetLevel(parsedLogLevel)

	processor := processors.TelegrafProcessor{}

	runOnce(processor)
	for range time.Tick(time.Duration(cfg.CollectIntervalSeconds * int(time.Second))) {
		runOnce(processor)
	}

	log.Infof("Stopping omnikeeper-deploy-agent (Version: %s)", version)
}

func runOnce(processor processors.Processor) error {
	ctx := context.Background()

	log.Debugf("Starting processing...")

	okClient, err := omnikeeper.BuildGraphQLClient(ctx, cfg.OmnikeeperBackendUrl, cfg.KeycloakClientId, cfg.Username, cfg.Password)
	if err != nil {
		return fmt.Errorf("Error building omnikeeper GraphQL client: %w", err)
	}

	log.Debugf("Starting fetch from omnikeeper...")
	outputItems, err := processor.Process(ctx, okClient, &log)
	if err != nil {
		return fmt.Errorf("Processing error: %w", err)
	}
	log.Debugf("Finished fetch from omnikeeper")

	log.Debugf("Updating output files...")
	_, err = updateOutputFiles(outputItems, &log)
	if err != nil {
		return fmt.Errorf("Error updating output files: %w", err)
	}
	log.Debugf("Finished updating output files")

	// TODO: trigger ansible for each updated item
	// updatedItems

	log.Debugf("Finished processing")

	return nil
}

func updateOutputFiles(outputItems map[string]interface{}, log *logrus.Logger) (map[string]bool, error) {
	if _, err := os.Stat(cfg.OutputDirectory); os.IsNotExist(err) {
		err = os.Mkdir(cfg.OutputDirectory, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("Error creating output directory: %w", err)
		}
	}
	processedFiles := make(map[string]bool, len(outputItems))
	updatedItems := make(map[string]bool, len(outputItems))
	for id, output := range outputItems {
		newJsonOutput, err := json.MarshalIndent(output, "", " ")
		if err != nil {
			log.Errorf("Error marshalling output JSON for ID %s: %w", id, err)
			continue
		}
		outputFilename := id + ".json"
		fullOutputFilename := filepath.Join(cfg.OutputDirectory, outputFilename)

		oldFile, err := os.Open(fullOutputFilename)
		defer oldFile.Close()
		var oldJsonOutput []byte = nil
		if err == nil {
			// an old file exists, read it
			readOldJsonOutput, err := ioutil.ReadAll(oldFile)
			if err != nil {
				// if we cannot read it, log a warning, but otherwise continue
				log.Warningf("Error reading existing output file %s: %w", outputFilename, err)
			} else {
				oldJsonOutput = readOldJsonOutput
			}
		}

		// compare old and new data, byte-by-byte, write/update output file only if a difference was detected
		if bytes.Compare(oldJsonOutput, newJsonOutput) != 0 {
			err = ioutil.WriteFile(fullOutputFilename, newJsonOutput, os.ModePerm)
			if err != nil {
				log.Errorf("Error writing output JSON for ID %s: %w", id, err)
				continue
			}
			updatedItems[id] = true
			log.Tracef("Updated output file %s", outputFilename)
		}

		processedFiles[outputFilename] = true
	}
	// delete old files (i.e. files that have not been written)
	dirRead, _ := os.Open(cfg.OutputDirectory)
	dirFiles, _ := dirRead.Readdir(0)
	for index := range dirFiles {
		fileHere := dirFiles[index]
		filename := fileHere.Name()

		if !processedFiles[filename] {
			fullFilename := filepath.Join(cfg.OutputDirectory, filename)
			err := os.Remove(fullFilename)
			if err != nil {
				log.Errorf("Error deleting old file %s: %w", filename, err)
				continue
			}
		}
	}
	return updatedItems, nil
}
