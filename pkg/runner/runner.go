package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/ansible"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/config"
	"github.com/max-bytes/omnikeeper-deploy-agent/pkg/omnikeeper"

	"github.com/sirupsen/logrus"
)

var (
	version    = "0.0.0-src"
	configFile = flag.String("config", "config.yml", "Config file location")
	log        logrus.Logger
)

var cfg = config.Configuration{}

func init() {
	log = *logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.TraceLevel) // is overwritten by configuration below
}

func Run(processor Processor) {
	log.Infof("omnikeeper-deploy-agent (Version: %s)", version)
	flag.Parse()

	log.Infof("Loading config from file: %s", *configFile)
	err := config.ReadConfigFromFilename(*configFile, &cfg)
	if err != nil {
		log.Fatalf("Error opening config file: %s", err)
	}

	parsedLogLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error parsing loglevel in config file: %s", err)
	}
	log.SetLevel(parsedLogLevel)

	runOnce(processor, cfg, &log)
	for range time.Tick(time.Duration(cfg.CollectIntervalSeconds * int(time.Second))) {
		runOnce(processor, cfg, &log)
	}

	log.Infof("Stopping omnikeeper-deploy-agent (Version: %s)", version)
}

func runOnce(processor Processor, cfg config.Configuration, log *logrus.Logger) {
	ctx := context.Background()

	log.Debugf("Starting processing...")

	okClient, err := omnikeeper.BuildGraphQLClient(ctx, cfg.OmnikeeperBackendUrl, cfg.KeycloakClientId, cfg.Username, cfg.Password)
	if err != nil {
		log.Errorf("Error building omnikeeper GraphQL client: %w", err)
		return
	}

	log.Debugf("Starting fetch from omnikeeper...")
	outputItems, err := processor.Process(ctx, okClient, log)
	if err != nil {
		log.Errorf("Processing error: %w", err)
		return
	}
	log.Debugf("Finished fetch from omnikeeper")

	log.Debugf("Creating variables files...")
	updatedItems, err := createVariablesFiles(outputItems, cfg.OutputDirectory, log)
	if err != nil {
		log.Errorf("Error creating variables files: %w", err)
		return
	}
	log.Debugf("Finished creating variables files")

	if len(updatedItems) > 0 {

		log.Debugf("Running ansible for updated items...")
		for id := range updatedItems {
			ansibleItemErr := ansible.Callout(ctx, cfg.Ansible, id)

			fullProcessedFilename := buildFullProcessedFilename(id, cfg.OutputDirectory)
			if ansibleItemErr != nil {
				log.Errorf("Error running ansible for item %s: %v", id, ansibleItemErr)

				// delete the .processed file, if present
				_ = os.Remove(fullProcessedFilename)
			} else {
				// place a .processed file to indicate that ansible successfully processed the host
				_, err := os.OpenFile(fullProcessedFilename, os.O_RDONLY|os.O_CREATE, 0666)
				if err != nil {
					// can't do much else other than report the error
					log.Errorf("Error writing .processed file for item %s: %v", id, err)
				}
			}
		}
		log.Debugf("Finished running ansible for updated items...")
	} else {
		log.Debugf("Skipping running ansible because no items were updated")
	}

	log.Debugf("Finished processing")
}

func buildProcessedFilename(id string) string {
	return id + ".processed"
}
func buildFullProcessedFilename(id string, outputDirectory string) string {
	return filepath.Join(outputDirectory, buildProcessedFilename(id))
}
func buildOutputFilename(id string) string {
	return id + ".json"
}
func buildFullOutputFilename(id string, outputDirectory string) string {
	return filepath.Join(outputDirectory, buildOutputFilename(id))
}

func createVariablesFiles(outputItems map[string]interface{}, outputDirectory string, log *logrus.Logger) (map[string]bool, error) {
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		err = os.Mkdir(outputDirectory, os.ModePerm)
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
		outputFilename := buildOutputFilename(id)
		fullOutputFilename := buildFullOutputFilename(id, outputDirectory)

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

		processedFilename := buildProcessedFilename(id)
		fullProcessedFilename := buildFullProcessedFilename(id, outputDirectory)
		_, errFPCheck := os.Stat(fullProcessedFilename)
		processedFileExists := !os.IsNotExist(errFPCheck)

		// first check if a processed file exists
		// then, compare old and new data, byte-by-byte, write/update output file only if a difference was detected
		if !processedFileExists || bytes.Compare(oldJsonOutput, newJsonOutput) != 0 {
			err = ioutil.WriteFile(fullOutputFilename, newJsonOutput, os.ModePerm)
			if err != nil {
				log.Errorf("Error writing output JSON for ID %s: %w", id, err)
				continue
			}
			updatedItems[id] = true
			log.Tracef("Updated variable file %s", outputFilename)
		}

		processedFiles[outputFilename] = true
		processedFiles[processedFilename] = true
	}
	// delete old items (i.e. files that have not been processed)
	dirRead, _ := os.Open(outputDirectory)
	dirFiles, _ := dirRead.Readdir(0)
	for index := range dirFiles {
		fileHere := dirFiles[index]
		filename := fileHere.Name()

		if !processedFiles[filename] {
			fullFilename := filepath.Join(outputDirectory, filename)
			err := os.Remove(fullFilename)
			if err != nil {
				log.Errorf("Error deleting old file %s: %w", filename, err)
				continue
			}
		}
	}
	return updatedItems, nil
}
