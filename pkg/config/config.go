package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"gopkg.in/yaml.v2"
)

func ReadConfigFromBytes(byteValue []byte, cfg *Configuration) error {
	err := yaml.Unmarshal(byteValue, &cfg)
	if err != nil {
		return fmt.Errorf("can't parse config file: %w", err)
	}
	return nil
}

func ReadConfigFromFilename(configFile string, cfg *Configuration) error {
	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("can't open config file: %w", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can't read config file: %w", err)
	}

	return ReadConfigFromBytes(byteValue, cfg)
}

type Configuration struct {
	LogLevel               string               `yaml:"log_level"`
	Username               string               `yaml:"username"`
	Password               string               `yaml:"password"`
	OmnikeeperBackendUrl   string               `yaml:"omnikeeper_backend_url"`
	KeycloakClientId       string               `yaml:"keycloak_client_id"`
	CollectIntervalSeconds int                  `yaml:"collect_interval_seconds"`
	OutputDirectory        string               `yaml:"output_directory"`
	Ansible                AnsibleCalloutConfig `yaml:"ansible"`
}

type AnsibleCalloutConfig struct {
	Playbooks         []string
	Options           *playbook.AnsiblePlaybookOptions
	ConnectionOptions *options.AnsibleConnectionOptions
	AnsibleBinary     string
}
