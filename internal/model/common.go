package model

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const defaultSqlDsn = "root:123456@tcp(127.0.0.1:3306)/lumina?charset=utf8mb4&parseTime=True&loc=Local"

var DB *gorm.DB

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
	err := db.AutoMigrate(&User{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Job{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Device{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Workflow{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Message{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&AccessToken{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Conversation{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&LLMMessage{})
	if err != nil {
		return err
	}

	return nil
}

func InsertTestData(db *gorm.DB) error {
	return nil
}
