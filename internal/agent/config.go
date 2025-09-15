package agent

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddr string `yaml:"server_addr"`
	WorkDir    string `yaml:"work_dir"`
}

func (c Config) ModelDir() string {
	return path.Join(c.WorkDir, "models")
}

func (c Config) DataDir() string {
	return path.Join(c.WorkDir, "data")
}

func DefaultConfig() *Config {
	return &Config{
		ServerAddr: "localhost:8080",
		WorkDir:    "./agent_dir",
	}
}

func LoadConfig(configPath string) (*Config, error) {
	conf := DefaultConfig()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %v", err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config file: %v", err)
	}

	return conf, nil
}
