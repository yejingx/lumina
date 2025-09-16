package agent

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type TritonConfig struct {
	ServerAddr   string `yaml:"serverAddr"`
	ModelRepoDir string `yaml:"modelRepoDir"`
}

type Config struct {
	LuminaServerAddr string       `yaml:"luminaServerAddr"`
	WorkDir          string       `yaml:"workDir"`
	Triton           TritonConfig `yaml:"triton"`
}

func (c Config) ModelDir() string {
	return path.Join(c.WorkDir, "models")
}

func (c Config) DataDir() string {
	return path.Join(c.WorkDir, "data")
}

func DefaultConfig() *Config {
	return &Config{
		LuminaServerAddr: "localhost:8080",
		WorkDir:    "./agent_dir",
		Triton: TritonConfig{
			ServerAddr:   "localhost:8001",
			ModelRepoDir: "./model_repo",
		},
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
