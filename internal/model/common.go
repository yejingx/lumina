package model

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"lumina/internal/config"
)

var DB *gorm.DB

func InitDB(dbConfig config.DBConfig) (*gorm.DB, error) {
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
	return nil
}

func InsertTestData(db *gorm.DB) error {
	return nil
}
