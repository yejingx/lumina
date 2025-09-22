package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Device struct {
	Id           int       `gorm:"primaryKey"`
	Uuid         string    `gorm:"type:char(96);unique"`
	Token        string    `gorm:"type:char(96);unique"`
	RegisterTime time.Time `gorm:"datetime;autoCreateTime"`
	LastPingTime time.Time `gorm:"datetime;"`
}

func (d *Device) IsRegistered() bool {
	return d.RegisterTime != time.Time{}
}

func (d *Device) Unregister() error {
	d.RegisterTime = time.Time{}
	return DB.Save(d).Error
}

func CreateDevice(d *Device) error {
	return DB.Create(d).Error
}

func DeleteDevice(id uint) error {
	return DB.Delete(&Device{}, id).Error
}

func GetDeviceById(id int) (*Device, error) {
	var d Device
	err := DB.First(&d, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func GetDeviceByUuid(uuid string) (*Device, error) {
	var d Device
	err := DB.Where("uuid = ?", uuid).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func GetDeviceByToken(token string) (*Device, error) {
	var d Device
	err := DB.Where("token = ?", token).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &d, err
}

func ListDevices(start, limit int) ([]Device, int64, error) {
	var devices []Device
	var total int64
	if err := DB.Model(&Device{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&Device{}).Offset(start).Limit(limit).Find(&devices).Error; err != nil {
		return nil, 0, err
	}
	return devices, total, nil
}

type AccessToken struct {
	Id          int       `gorm:"primaryKey"`
	AccessToken string    `gorm:"type:char(96);unique"`
	CreateTime  time.Time `gorm:"datetime;autoCreateTime"`
	ExpireTime  time.Time `gorm:"datetime;autoCreateTime"`
	DeviceUuid  string    `gorm:"type:char(96);index"`
}

func (t *AccessToken) IsExpired() bool {
	return t.ExpireTime.Before(time.Now())
}

func (t *AccessToken) IsBound() bool {
	return t.DeviceUuid != ""
}

func (t *AccessToken) BindDevice(d *Device) error {
	t.DeviceUuid = d.Uuid
	d.RegisterTime = time.Now()

	// 使用事务确保数据一致性
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(d).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Save(t).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func CreateAccessToken(t *AccessToken) error {
	return DB.Create(t).Error
}

func DeleteAccessToken(id uint) error {
	return DB.Delete(&AccessToken{}, id).Error
}

func GetAccessToken(id int) (*AccessToken, error) {
	var t AccessToken
	err := DB.First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func GetAccessTokenByToken(token string) (*AccessToken, error) {
	var t AccessToken
	err := DB.Where("token = ?", token).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func ListAccessToken(start, limit int) ([]AccessToken, int64, error) {
	var accessTokens []AccessToken
	var total int64
	if err := DB.Model(&AccessToken{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&AccessToken{}).Offset(start).Limit(limit).Find(&accessTokens).Error; err != nil {
		return nil, 0, err
	}
	return accessTokens, total, nil
}
