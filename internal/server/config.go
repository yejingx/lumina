package server

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"lumina/internal/model"
)

const defaultSqlDsn = "root:123456@tcp(127.0.0.1:3306)/lumina?charset=utf8mb4&parseTime=True&loc=Local"

type S3Config struct {
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	UseSSL          bool   `yaml:"useSSL"`
	Region          string `yaml:"region"`
}

type Config struct {
	Addr      string         `yaml:"addr"`
	SSLCert   string         `yaml:"sslCert"`
	SSLKey    string         `yaml:"sslKey"`
	JwtSecret string         `yaml:"jwtSecret"`
	DB        model.DBConfig `yaml:"db"`
	S3        S3Config       `yaml:"s3"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr: "127.0.0.1:8081",
		DB: model.DBConfig{
			DSN:          defaultSqlDsn,
			MaxIdleConns: 100,
			MaxOpenConns: 1000,
			MaxLifetime:  60,
		},
		S3: S3Config{
			Bucket:   "lumina",
			Endpoint: "127.0.0.1:9000",
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
