package server

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"lumina/internal/agent"
	"lumina/internal/model"
)

type S3Config struct {
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	UseSSL          bool   `yaml:"useSSL"`
	Region          string `yaml:"region"`
	VisitEndpoint   string `yaml:"visitEndpoint"`
}

func (c S3Config) VisitPrefix() string {
	if c.VisitEndpoint == "" {
		if c.UseSSL {
			return fmt.Sprintf("https://%s/%s", c.Endpoint, c.Bucket)
		}
		return fmt.Sprintf("http://%s/%s", c.Endpoint, c.Bucket)
	}
	return c.VisitEndpoint + "/" + c.Bucket
}

type InfluxDBConfig struct {
	URL     string `yaml:"url"`
	Org     string `yaml:"org"`
	Bucket  string `yaml:"bucket"`
	Token   string `yaml:"token"`
	Enabled bool   `yaml:"enabled"`
}

type Config struct {
	Addr      string          `yaml:"addr"`
	SSLCert   string          `yaml:"sslCert"`
	SSLKey    string          `yaml:"sslKey"`
	JwtSecret string          `yaml:"jwtSecret"`
	DB        model.DBConfig  `yaml:"db"`
	S3        S3Config        `yaml:"s3"`
	LLM       agent.LLMConfig `yaml:"llm"`
	InfluxDB  InfluxDBConfig  `yaml:"influxdb"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr: "127.0.0.1:8081",
		DB:   *model.DefaultDBConfig(),
		S3: S3Config{
			Bucket:   "lumina",
			Endpoint: "127.0.0.1:9000",
			UseSSL:   false,
			Region:   "us-east-1",
		},
		LLM: agent.LLMConfig{
			Model:   "gpt-3.5-turbo",
			BaseUrl: "https://api.openai.com/v1",
			Timeout: 300 * time.Second,
		},
		InfluxDB: InfluxDBConfig{
			URL:     "http://127.0.0.1:48086",
			Org:     "lumina",
			Bucket:  "lumina",
			Token:   "",
			Enabled: false,
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
