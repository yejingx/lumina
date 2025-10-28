package dao

import (
	"lumina/pkg/str"
	"time"

	"lumina/internal/model"
)

type CameraSpec struct {
	Id         int                  `json:"id"`
	Uuid       string               `json:"uuid" binding:"required"`
	Name       string               `json:"name" binding:"required"`
	Protocol   model.CameraProtocol `json:"protocol" binding:"required"`
	Ip         string               `json:"ip" binding:"required"`
	Port       int                  `json:"port"`
	Path       string               `json:"path"`
	Username   string               `json:"username"`
	Password   string               `json:"password"`
	CreateTime string               `json:"createTime" binding:"required,datetime=2006-01-02T15:04:05Z07:00"`
	UpdateTime string               `json:"updateTime" binding:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

func FromCameraModel(m *model.Camera) *CameraSpec {
	if m == nil {
		return nil
	}
	c := &CameraSpec{}
	c.Id = m.Id
	c.Uuid = m.Uuid
	c.Name = m.Name
	c.Protocol = m.Protocol
	c.Ip = m.Ip
	c.Port = int(m.Port)
	c.Path = m.Path
	c.Username = m.Username
	c.CreateTime = m.CreateTime.Format(time.RFC3339)
	c.UpdateTime = m.UpdateTime.Format(time.RFC3339)
	return c
}

type ListCamerasRequest struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

type ListCamerasResponse struct {
	Items []CameraSpec `json:"items"`
	Total int64        `json:"total"`
}

type CreateCameraRequest struct {
	Name     string               `json:"name" binding:"required"`
	Protocol model.CameraProtocol `json:"protocol" binding:"required"`
	Ip       string               `json:"ip" binding:"required"`
	Port     int                  `json:"port"`
	Path     string               `json:"path"`
	Username string               `json:"username"`
	Password string               `json:"password"`
}

func (c *CreateCameraRequest) ToModel() *model.Camera {
	return &model.Camera{
		Uuid:     str.GenDeviceId(16),
		Name:     c.Name,
		Protocol: c.Protocol,
		Ip:       c.Ip,
		Port:     c.Port,
		Path:     c.Path,
		Username: c.Username,
		Password: c.Password,
	}
}

type CreateCameraResponse struct {
	Uuid string `json:"uuid"`
}

type UpdateCameraRequest struct {
	Name     *string               `json:"name"`
	Protocol *model.CameraProtocol `json:"protocol"`
	Ip       *string               `json:"ip"`
	Port     *int                  `json:"port"`
	Path     *string               `json:"path"`
	Username *string               `json:"username"`
	Password *string               `json:"password"`
}

func (req *UpdateCameraRequest) UpdateModel(c *model.Camera) {
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Protocol != nil {
		c.Protocol = *req.Protocol
	}
	if req.Ip != nil {
		c.Ip = *req.Ip
	}
	if req.Port != nil {
		c.Port = *req.Port
	}
	if req.Path != nil {
		c.Path = *req.Path
	}
	if req.Username != nil {
		c.Username = *req.Username
	}
	if req.Password != nil {
		c.Password = *req.Password
	}
}
