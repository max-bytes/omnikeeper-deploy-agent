package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return fmt.Errorf("can't parse config file: %w", err)
	}
	return nil
}

type Configuration struct {
	LogLevel               string `json:"log_level"`
	Username               string
	Password               string
	ServerUrl              string   `json:"server_url"`
	AuthURL                string   `json:"auth_url"`
	TokenURL               string   `json:"token_url"`
	ClientId               string   `json:"client_id"`
	CollectInterval        int      `json:"collect_interval"`
	TraitId                string   `json:"trait_id"`
	LayerIds               []string `json:"layer_ids"`
	ConfigType             string   `json:"config_type"`
	NaemonNameIdentifier   string   `json:"naemon_name_identifier"`
	NaemonConfigIdentifier string   `json:"naemon_config_identifier"`
	NaemonConfigDirectory  string   `json:"naemon_config_directory"`
}
