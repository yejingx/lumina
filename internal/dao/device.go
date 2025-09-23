package dao

import (
	"time"

	"lumina/internal/model"
)

type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	AccessToken string `json:"accessToken" binding:"required"`
	Uuid        string `json:"uuid"`
}

type S3Config struct {
	AccessKeyID     *string `json:"accessKeyID"`
	SecretAccessKey *string `json:"secretAccessKey"`
}

type RegisterResponse struct {
	Uuid              string `json:"uuid"`
	Token             string `json:"token"`
	S3AccessKeyID     string `json:"s3AccessKeyID"`
	S3SecretAccessKey string `json:"s3SecretAccessKey"`
}

type AccessTokenSpec struct {
	Id          int    `json:"id"`
	AccessToken string `json:"accessToken" binding:"required"`
	CreateTime  string `json:"createTime" binding:"datetime=2006-01-02T15:04:05Z07:00"`
	ExpireTime  string `json:"expireTime" binding:"datetime=2006-01-02T15:04:05Z07:00"`
	DeviceUuid  string `json:"deviceUuid"`
}

func FromAccessTokenModel(m *model.AccessToken) *AccessTokenSpec {
	if m == nil {
		return nil
	}
	t := &AccessTokenSpec{}
	t.Id = int(m.Id)
	t.AccessToken = m.AccessToken
	t.CreateTime = m.CreateTime.Format(time.RFC3339)
	t.ExpireTime = m.ExpireTime.Format(time.RFC3339)
	if m.DeviceUuid != "" {
		t.DeviceUuid = m.DeviceUuid
	}
	return t
}

type CreateAccessTokenRequest struct {
	ExpireTime string `json:"expireTime" binding:"datetime=2006-01-02T15:04:05Z07:00"`
}

type CreateAccessTokenResponse struct {
	AccessToken string `json:"token"`
}

type ListAccessTokenRequest struct {
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=50"`
}

type ListAccessTokenResponse struct {
	AccessTokens []AccessTokenSpec `json:"accessTokens"`
	Total        int64             `json:"total"`
}

type DeviceSpec struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Token        string `json:"token"`
	Uuid         string `json:"uuid"`
	RegisterTime string `json:"registerTime"`
	LastPingTime string `json:"lastPingTime"`
}

func FromDeviceModel(m *model.Device) *DeviceSpec {
	if m == nil {
		return nil
	}
	t := &DeviceSpec{}
	t.Id = m.Id
	t.Token = m.Token
	t.Uuid = m.Uuid
	t.Name = m.Name
	t.RegisterTime = m.RegisterTime.Time.Format(time.RFC3339)
	if !m.LastPingTime.Time.IsZero() {
		t.LastPingTime = m.LastPingTime.Time.Format(time.RFC3339)
	}
	return t
}

type ListDeviceRequest struct {
	Start int `json:"start" form:"start" binding:"min=0"`
	Limit int `json:"limit" form:"limit" binding:"min=0,max=100"`
}

type ListDeviceResponse struct {
	Devices []DeviceSpec `json:"devices"`
	Total   int64        `json:"total"`
}
