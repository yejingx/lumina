package consumer

import (
	"fmt"
	"os"
	"time"

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

type VLMConfig struct {
	APIKey  string        `yaml:"apiKey"`
	BaseURL string        `yaml:"baseURL"`
	Model   string        `yaml:"model"`
	Timeout time.Duration `yaml:"timeout"` // 超时时间(秒)
}

type DifyConfig struct {
	APIKey  string        `yaml:"apiKey"`
	BaseURL string        `yaml:"baseURL"`
	Timeout time.Duration `yaml:"timeout"`
}

type Config struct {
	NSQ  NSQConfig  `yaml:"nsq"`
	S3   S3Config   `yaml:"s3"`
	VLM  VLMConfig  `yaml:"vlm"`
	Dify DifyConfig `yaml:"dify"`
}

func DefaultConfig() *Config {
	return &Config{
		NSQ: NSQConfig{
			NSQDAddrs: []string{"localhost:4150"},
			Topic:     "detection_results",
		},
		S3: S3Config{
			Bucket:   "lumina",
			Endpoint: "localhost:9000",
			UseSSL:   false,
			Region:   "us-east-1",
		},
		VLM: VLMConfig{
			APIKey:  "",
			BaseURL: "https://api.openai.com/v1",
			Model:   "gpt-4-vision-preview",
			Timeout: 30 * time.Second,
		},
		Dify: DifyConfig{
			APIKey:  "",
			BaseURL: "https://api.dify.cn/v1",
			Timeout: 30 * time.Second,
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
