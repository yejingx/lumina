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

type NSQConfig struct {
	NSQDAddr string `yaml:"nsqdAddr"`
	Topic    string `yaml:"topic"`
}

type S3Config struct {
	Bucket   string `json:"bucket"`
	Endpoint string `json:"endpoint"`
	UseSSL   bool   `json:"useSSL,omitempty"`
	Region   string `json:"region,omitempty"`
}

type Config struct {
	LuminaServerAddr string       `yaml:"luminaServerAddr"`
	WorkDir          string       `yaml:"workDir"`
	Triton           TritonConfig `yaml:"triton"`
	NSQ              NSQConfig    `yaml:"nsq"`
	S3               S3Config     `yaml:"s3"`
}

func (c Config) ModelDir() string {
	return path.Join(c.WorkDir, "models")
}

func (c Config) DataDir() string {
	return path.Join(c.WorkDir, "data")
}

func (c Config) JobDir() string {
	return path.Join(c.WorkDir, "job")
}

func DefaultConfig() *Config {
	return &Config{
		LuminaServerAddr: "localhost:8080",
		WorkDir:          "./agent_dir",
		Triton: TritonConfig{
			ServerAddr:   "localhost:8001",
			ModelRepoDir: "./model_repo",
		},
		NSQ: NSQConfig{
			NSQDAddr: "localhost:4150",
			Topic:    "detection_results",
		},
		S3: S3Config{
			Bucket:   "lumina",
			Endpoint: "localhost:9000",
			UseSSL:   false,
			Region:   "us-east-1",
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
