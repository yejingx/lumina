package config

const defaultSqlDsn = "root:123456@tcp(127.0.0.1:3306)/lumina?charset=utf8mb4&parseTime=True&loc=Local"

type DBConfig struct {
	DSN          string `yaml:"dsn"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxLifetime  int    `yaml:"maxLifetime"`
}

type S3Config struct {
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	UseSSL          bool   `yaml:"useSSL"`
	Region          string `yaml:"region"`
}

type Config struct {
	Addr          string   `yaml:"addr"`
	SSLCert       string   `yaml:"sslCert"`
	SSLKey        string   `yaml:"sslKey"`
	JwtSecret     string   `yaml:"jwtSecret"`
	DB            DBConfig `yaml:"db"`
	S3            S3Config `yaml:"s3"`
	ModelRepoAddr string   `yaml:"modelRepoAddr"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr: "127.0.0.1:8081",
		DB: DBConfig{
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
