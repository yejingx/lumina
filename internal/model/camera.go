package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type CameraProtocol string

const (
	CameraProtocolRtmp CameraProtocol = "rtmp"
	CameraProtocolRtsp CameraProtocol = "rtsp"
)

type Camera struct {
	Id           int            `gorm:"primaryKey"`
	Uuid         string         `gorm:"type:char(96);unique"`
	Name         string         `gorm:"type:char(96)"`
	Protocol     CameraProtocol `gorm:"type:char(96)"`
	Ip           string         `gorm:"type:char(96)"`
	Port         int            `gorm:"type:int"`
	Path         string         `gorm:"type:char(96)"`
	Username     string         `gorm:"type:char(96)"`
	Password     string         `gorm:"type:char(96)"`
	CreateTime   time.Time      `gorm:"datetime;autoCreateTime"`
	UpdateTime   time.Time      `gorm:"datetime;autoCreateTime;autoUpdateTime"`
	BindDeviceId int            `gorm:"type:int"`
}

func (c *Camera) BindDevice() (*Device, error) {
	if c.BindDeviceId == 0 {
		return nil, nil
	}
	dev, err := GetDeviceById(c.BindDeviceId)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func CreateCamera(camera *Camera) error {
	return DB.Create(camera).Error
}

func GetCameraById(id int) (*Camera, error) {
	var camera Camera
	err := DB.Where("id = ?", id).First(&camera).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &camera, nil
}

func GetCameraByUuid(uuid string) (*Camera, error) {
	var camera Camera
	err := DB.Where("uuid = ?", uuid).First(&camera).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &camera, nil
}

func DeleteCamera(camera *Camera) error {
	return DB.Delete(camera).Error
}

func UpdateCamera(camera *Camera) error {
	return DB.Save(camera).Error
}

func ListCameras(start, limit int) ([]Camera, int64, error) {
	var cameras []Camera
	var total int64
	if err := DB.Model(&Camera{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := DB.Model(&Camera{}).Offset(start).Limit(limit).Find(&cameras).Error; err != nil {
		return nil, 0, err
	}
	return cameras, total, nil
}
