package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func ReadConfigFromFile(configFile string, cfg *Configuration) error {
	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("can't open config file: %w", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can't read config file: %w", err)
	}
	err = yaml.Unmarshal(byteValue, &cfg)
	if err != nil {
		return fmt.Errorf("can't parse config file: %w", err)
	}
	return nil
}

type Configuration struct {
	LogLevel               string `yaml:"log_level"`
	Username               string `yaml:"username"`
	Password               string `yaml:"password"`
	OmnikeeperBackendUrl   string `yaml:"omnikeeper_backend_url"`
	KeycloakClientId       string `yaml:"keycloak_client_id"`
	CollectIntervalSeconds int    `yaml:"collect_interval_seconds"`
	OutputDirectory        string `yaml:"output_directory"`
}
