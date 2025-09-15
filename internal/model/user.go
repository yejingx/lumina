package model

import (
	"time"
)

type User struct {
	Id          int       `gorm:"primarykey"`
	Username    string    `json:"username" gorm:"type:char(96);uniqueIndex"`
	Nickname    string    `json:"nickname" gorm:"type:char(96)"`
	Password    string    `json:"password" gorm:"type:char(96)"`
	AccessToken string    `json:"access_token" gorm:"type:char(96);uniqueIndex"`
	IsAdmin     bool      `json:"is_admin" gorm:"default:false"`
	CreatedTime time.Time `json:"created_time" gorm:"datetime;autoCreateTime"`
}

func CreateUser(user *User) error {
	return DB.Create(user).Error
}

func GetUserById(id int) (*User, error) {
	var user User
	err := DB.First(&user, id).Error
	return &user, err
}

func GetUserByToken(token string) (*User, error) {
	var user User
	err := DB.Where("access_token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(user *User) error {
	return DB.Save(user).Error
}

func DeleteUser(id int) error {
	return DB.Delete(&User{}, id).Error
}

func CountUsers() (int, error) {
	var count int64
	err := DB.Model(&User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func GetUsers(start, limit int) ([]*User, error) {
	var users []*User
	err := DB.Offset(start).Limit(limit).Find(&users).Order("id desc").Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
