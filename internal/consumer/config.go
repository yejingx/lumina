package consumer

import (
	"fmt"
	"lumina/internal/model"
	"os"

	"gopkg.in/yaml.v2"
)

type NSQConfig struct {
	NSQDAddrs []string `yaml:"nsqdAddrs"` // 支持多个NSQ地址
	Topic     string   `yaml:"topic"`
}

type S3Config struct {
	Bucket   string `yaml:"bucket"`
	Endpoint string `yaml:"endpoint"`
	UseSSL   bool   `yaml:"useSSL,omitempty"`
	Region   string `yaml:"region,omitempty"`
}

func (s3 *S3Config) UrlPrefix() string {
	if s3.UseSSL {
		return fmt.Sprintf("https://%s/%s", s3.Endpoint, s3.Bucket)
	}
	return fmt.Sprintf("http://%s/%s", s3.Endpoint, s3.Bucket)
}

type InfluxDBConfig struct {
	URL     string `yaml:"url"`
	Org     string `yaml:"org"`
	Bucket  string `yaml:"bucket"`
	Token   string `yaml:"token"`
	Enabled bool   `yaml:"enabled"`
}

type Config struct {
	NSQ      NSQConfig      `yaml:"nsq"`
	S3       S3Config       `yaml:"s3"`
	DB       model.DBConfig `yaml:"db"`
	InfluxDB InfluxDBConfig `yaml:"influxdb"`
}

func DefaultConfig() *Config {
	return &Config{
		NSQ: NSQConfig{
			NSQDAddrs: []string{"localhost:4150"},
			Topic:     "detection_results",
		},
		DB: *model.DefaultDBConfig(),
		S3: S3Config{
			Bucket:   "lumina",
			Endpoint: "localhost:9000",
			UseSSL:   false,
			Region:   "us-east-1",
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
