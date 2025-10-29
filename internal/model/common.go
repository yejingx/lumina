package model

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const defaultSqlDsn = "root:123456@tcp(127.0.0.1:3306)/lumina?charset=utf8mb4&parseTime=True&loc=Local"

var DB *gorm.DB
var Redis *redis.Client

type DBConfig struct {
	DSN          string `yaml:"dsn"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxLifetime  int    `yaml:"maxLifetime"`
}

func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		DSN:          defaultSqlDsn,
		MaxIdleConns: 100,
		MaxOpenConns: 1000,
		MaxLifetime:  60,
	}
}

type RedisConfig struct {
	MasterName    string   `yaml:"masterName"`
	SentinelAddrs []string `yaml:"sentinelAddrs"`
	Password      string   `yaml:"password,omitempty"`
	DB            int      `yaml:"db,omitempty"`
}

func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"127.0.0.1:26379"},
		Password:      "",
		DB:            0,
	}
}

func InitRedis(cfg RedisConfig) (*redis.Client, error) {
	cli := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: cfg.SentinelAddrs,
		Password:      cfg.Password,
		DB:            cfg.DB,
	})
	// Ping to validate connection at init time
	if err := cli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	Redis = cli
	return cli, nil
}

func InitDB(dbConfig DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dbConfig.DSN), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(dbConfig.MaxLifetime))

	DB = db

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	for _, model := range []any{
		&User{},
		&Job{},
		&Device{},
		&Workflow{},
		&Message{},
		&AccessToken{},
		&Conversation{},
		&ChatMessage{},
		&AlertMessage{},
		&Camera{},
	} {
		err := db.AutoMigrate(model)
		if err != nil {
			return err
		}
	}

	// Ensure ChatMessage.answer uses a large text type to avoid overflow errors
	// MySQL TEXT/LONGTEXT columns cannot have default values; tag has been updated.
	// AutoMigrate does not always change existing column types, so we enforce it here.
	_ = db.Exec("ALTER TABLE chat_messages MODIFY COLUMN answer LONGTEXT").Error

	return nil
}

func InsertTestData(db *gorm.DB) error {
	return nil
}
