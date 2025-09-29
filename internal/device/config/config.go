package config

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

func (s3 *S3Config) UrlPrefix() string {
	if s3.UseSSL {
		return fmt.Sprintf("https://%s/%s", s3.Endpoint, s3.Bucket)
	}
	return fmt.Sprintf("http://%s/%s", s3.Endpoint, s3.Bucket)
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
	cfg := &Config{
		LuminaServerAddr: "http://localhost:8080",
		Triton: TritonConfig{
			ServerAddr: "localhost:8001",
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

	dataDir := os.Getenv("LUMINA_DATA")
	if dataDir != "" {
		cfg.WorkDir = path.Join(dataDir, "device_dir")
		cfg.Triton.ModelRepoDir = path.Join(dataDir, "model_repo")
	} else {
		cfg.WorkDir = "./device_dir"
		cfg.Triton.ModelRepoDir = "./model_repo"
	}

	return cfg
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
