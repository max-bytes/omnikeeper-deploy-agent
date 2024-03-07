package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/max-bytes/omnikeeper-deploy-agent/v2/pkg/ansible"
	"github.com/max-bytes/omnikeeper-deploy-agent/v2/pkg/config"
	"github.com/max-bytes/omnikeeper-deploy-agent/v2/pkg/healthcheck"
	"github.com/max-bytes/omnikeeper-deploy-agent/v2/pkg/omnikeeper"

	"github.com/sirupsen/logrus"
)

var cfg = config.Configuration{}
var logCollector = NewLogCollectorHook()

func Run(processor Processor, configFile string, log *logrus.Logger) {
	log.Infof("Loading config from file: %s", configFile)
	err := config.ReadConfigFromFilename(configFile, &cfg)
	if err != nil {
		log.Fatalf("Error opening config file: %s", err)
	}

	parsedLogLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error parsing loglevel in config file: %s", err)
	}
	log.SetLevel(parsedLogLevel)

	// NOTE: touch stats file at the beginning
	healthcheck.TouchStatFile()

	log.AddHook(logCollector)

	ticker := time.NewTicker(time.Duration(cfg.CollectIntervalSeconds * int(time.Second)))
	for ; true; <-ticker.C {
		runOnce(processor, configFile, cfg, log)
		ticker.Reset(time.Duration(cfg.CollectIntervalSeconds * int(time.Second)))
	}
}

func runOnce(processor Processor, configFile string, cfg config.Configuration, log *logrus.Logger) {
	ctx := context.Background()

	log.Debugf("Starting processing...")

	okClient, err := omnikeeper.BuildGraphQLClient(ctx, cfg.OmnikeeperBackendUrl, cfg.KeycloakClientId, cfg.Username, cfg.Password, cfg.OmnikeeperInsecureSkipVerify)
	if err != nil {
		log.Errorf("Error building omnikeeper GraphQL client: %v", err)
		return
	}

	log.Debugf("Starting fetch from omnikeeper and processing...")
	outputItems, err := processor.Process(configFile, ctx, okClient, log)
	if err != nil {
		log.Errorf("Processing error: %v", err)
		return
	}
	log.Debugf("Finished fetch from omnikeeper and processing")

	log.Debugf("Creating variables files...")
	updatedItems, err := createVariablesFiles(outputItems, cfg.OutputDirectory, log)
	if err != nil {
		log.Errorf("Error creating variables files: %v", err)
		return
	}
	log.Debugf("Finished creating variables files")

	logCollector.ClearLogs()

	itemErr := make(map[string][]error)
	itemErrMutex := &sync.Mutex{}
	if len(updatedItems) > 0 {
		log.Debugf("Running ansible for updated items...")

		if cfg.Ansible.ParallelProcessing {
			log.Debugf("Running in parallel...")
			// parallel processing of playbooks
			var wg sync.WaitGroup
			wg.Add(len(updatedItems))

			for id := range updatedItems {
				itemLog := log.WithField("item", id)
				go func(id string) {
					defer wg.Done()

					err := runItem(id, ctx, itemLog)
					if err != nil {
						itemErrMutex.Lock()
						itemErr[id] = append(itemErr[id], err)
						itemErrMutex.Unlock()
					}
				}(id)
			}
			wg.Wait()
		} else {
			log.Debugf("Running in series...")
			// serial processing of playbooks
			for id := range updatedItems {
				itemLog := log.WithField("item", id)
				err := runItem(id, ctx, itemLog)
				if err != nil {
					itemErr[id] = append(itemErr[id], err)
				}
			}
		}

		log.Debugf("Finished running ansible for updated items...")
	} else {
		log.Debugf("Skipping running ansible because no items were updated")
	}

	if len(itemErr) == 0 {
		healthcheck.TouchStatFile()
	} else {
		log.Errorf("Encountered errors in %d items... items with errors will be re-run", len(itemErr))
	}

	// post-process
	results := make(map[string]ProcessResultItem)
	itemLogs := logCollector.GetLogs()
	for id, logs := range itemLogs {
		results[id] = ProcessResultItem{
			Logs:    logs,
			Success: len(itemErr[id]) <= 0,
		}
	}
	err = processor.PostProcess(configFile, ctx, okClient, results)
	if err != nil {
		log.Errorf("Error post-processing: %v", err)
		return
	}

	log.Debugf("Finished processing")
}

func runItem(id string, ctx context.Context, itemLog *logrus.Entry) error {
	fullOutputFilename := buildFullOutputFilename(id, cfg.OutputDirectory)
	ansibleItemErr := ansible.Callout(ctx, cfg.Ansible, id, fullOutputFilename, cfg.Ansible.Disabled, itemLog)

	fullProcessedFilename := buildFullProcessedFilename(id, cfg.OutputDirectory)
	if ansibleItemErr != nil {
		itemLog.Errorf("Error running ansible for item %s: %v", id, ansibleItemErr)
		// delete the .processed file, if present
		_ = os.Remove(fullProcessedFilename)
		return ansibleItemErr
	} else {
		// place a .processed file to indicate that ansible successfully processed the host
		_, err := os.OpenFile(fullProcessedFilename, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			// can't do much else other than report the error
			itemLog.Errorf("Error writing .processed file for item %s: %v", id, err)
			return err
		}
	}
	return nil
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
			log.Errorf("Error marshalling output JSON for ID %s: %v", id, err)
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
				log.Warningf("Error reading existing output file %s: %v", outputFilename, err)
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
				log.Errorf("Error writing output JSON for ID %s: %v", id, err)
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
	defer dirRead.Close()
	dirFiles, _ := dirRead.Readdir(0)
	for index := range dirFiles {
		fileHere := dirFiles[index]
		filename := fileHere.Name()

		if !processedFiles[filename] {
			fullFilename := filepath.Join(outputDirectory, filename)
			err := os.Remove(fullFilename)
			if err != nil {
				log.Errorf("Error deleting old file %s: %v", filename, err)
				continue
			}
		}
	}
	return updatedItems, nil
}
