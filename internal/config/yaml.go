package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// LoadYAMLConfig load config from filename in YAML format
func LoadYAMLConfig(filename string, cfg interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ReadFile: %v", err)
	}
	err = yaml.Unmarshal(data, cfg)
	return err
}

func InitConfig(configPath string) (*Config, error) {
	conf := DefaultConfig()

	err := LoadYAMLConfig(configPath, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
