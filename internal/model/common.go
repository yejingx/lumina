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
