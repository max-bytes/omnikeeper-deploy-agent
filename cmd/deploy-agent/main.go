package main

import (
	"context"
	"deploy-agent/pkg/config"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shurcooL/graphql"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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
	log.Infof("deploy-agent (Version: %s)", version)
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

	var waitGrp sync.WaitGroup
	waitGrp.Add(1)

	go func() {
		for range time.Tick(time.Duration(cfg.CollectInterval * int(time.Second))) {
			oauth2cfg := &oauth2.Config{
				ClientID: cfg.ClientId,
				Endpoint: oauth2.Endpoint{
					AuthURL:  cfg.AuthURL,
					TokenURL: cfg.TokenURL,
				},
			}

			ctx := context.Background()
			token, tokenErr := oauth2cfg.PasswordCredentialsToken(ctx, cfg.Username, cfg.Password)

			if tokenErr != nil {
				// return nil, tokenErr
				// log the error here
				continue
			}

			tokenSource := oauth2cfg.TokenSource(ctx, token)
			httpClient := oauth2.NewClient(ctx, tokenSource)
			client := graphql.NewClient(cfg.ServerUrl, httpClient)

			fmt.Print(httpClient)

			var query = ETQuery{}
			layers := make([]graphql.String, len(cfg.LayerIds))
			for i, v := range cfg.LayerIds {
				layers[i] = graphql.String(v)
			}
			variables := map[string]interface{}{
				"traitID": graphql.String(cfg.TraitId),
				"layers":  layers,
			}
			err := client.Query(context.Background(), &query, variables)
			if err != nil {
				// return nil, err
			}

			items := make(map[string]string)
			for _, value := range query.EffectiveTraitsForTrait {
				// fmt.Print(value)
				item := make(map[string]string)

				naemonName := ""
				naemonConfig := ""

				for _, v := range value.TraitAttributes {
					values := v.MergedAttribute.Attribute.Value.Values

					// if string(v.Identifier) == cfg.NaemonNameIdentifier {
					// 	naemonName =
					// }

					identifier := string(v.Identifier)
					if !v.MergedAttribute.Attribute.Value.IsArray {
						item[identifier] = string(values[0])
					} else {
						valuesStr := make([]string, len(values))
						for i, v := range values {
							valuesStr[i] = string(v)
						}
						item[identifier] = strings.Join(valuesStr, ",")
					}
				}

				if n, ok := item[cfg.NaemonNameIdentifier]; ok {
					naemonName = n
				}

				if c, ok := item[cfg.NaemonConfigIdentifier]; ok {
					naemonConfig = c
				}

				items[naemonName] = naemonConfig

			}

			// now we need to process foreach item
			if _, err := os.Stat(cfg.NaemonConfigDirectory); os.IsNotExist(err) {
				mkDirErr := os.Mkdir(cfg.NaemonConfigDirectory, os.ModePerm)
				if mkDirErr != nil {
					// log.Fatal(err)
				}
			}

			for naemon, config := range items {
				nPath := filepath.Join(cfg.NaemonConfigDirectory, "/", naemon)
				if _, err := os.Stat(nPath); os.IsNotExist(err) {
					nDirErr := os.Mkdir(nPath, os.ModePerm)
					if nDirErr != nil {
						// log error here
					}
				}

				wErr := ioutil.WriteFile(filepath.Join(nPath, "/", (naemon+".json")), []byte(config), os.ModePerm)
				if wErr != nil {

				}
				// fmt.Print(config[0])
			}
		}
	}()

	waitGrp.Wait()

	log.Infof("Stopping deploy-agent (Version: %s)", version)
}

type ETQuery struct {
	EffectiveTraitsForTrait []struct {
		TraitAttributes []struct {
			Identifier      graphql.String
			MergedAttribute struct {
				Attribute struct {
					Value struct {
						IsArray graphql.Boolean
						Values  []graphql.String
					}
				}
			}
		}
	} `graphql:"effectiveTraitsForTrait(traitID: $traitID, layers: $layers)"`
}

type NaemonItem struct {
	Name   string
	Config string
}
