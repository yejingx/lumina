package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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

type PreviewTask struct {
	TaskUuid   string    `json:"taskUuid"`
	PullAddr   string    `json:"pullAddr"`
	PushAddr   string    `json:"pushAddr"`
	ExpireTime time.Time `json:"expireTime,omitempty"`
}

const previewKeyTemplate = "preview:%s:%s"
const previewExpire = time.Minute

func previewKey(deviceUuid, cameraUuid string) string {
	return fmt.Sprintf(previewKeyTemplate, deviceUuid, cameraUuid)
}

func AddPreviewTask(ctx context.Context, deviceUuid, cameraUuid string, args *PreviewTask) error {
	data, _ := json.Marshal(args)
	return Redis.Set(ctx, previewKey(deviceUuid, cameraUuid), data, previewExpire).Err()
}

func GetPreviewTask(ctx context.Context, deviceUuid, cameraUuid string) (*PreviewTask, error) {
	var data []byte
	if err := Redis.Get(ctx, previewKey(deviceUuid, cameraUuid)).Scan(&data); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var task PreviewTask
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func TouchPreviewTask(ctx context.Context, deviceUuid, cameraUuid string) error {
	return Redis.Expire(ctx, previewKey(deviceUuid, cameraUuid), previewExpire).Err()
}

func GetPreviewTasksByDeviceUuid(ctx context.Context, deviceUuid string) ([]*PreviewTask, error) {
	keys, err := Redis.Keys(ctx, fmt.Sprintf(previewKeyTemplate, deviceUuid, "*")).Result()
	if err != nil {
		return nil, err
	}
	var tasks []*PreviewTask
	for _, key := range keys {
		var data []byte
		if err := Redis.Get(ctx, key).Scan(&data); err != nil {
			return nil, err
		}
		var task PreviewTask
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	return tasks, nil
}
