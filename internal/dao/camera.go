package dao

import (
	"fmt"
	"strings"
	"time"

	"lumina/internal/model"
	"lumina/pkg/str"
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
	BindDevice *DeviceSpec          `json:"bindDevice,omitempty"`
}

func (c CameraSpec) Url() string {
	port := c.Port
	if port == 0 {
		switch c.Protocol {
		case model.CameraProtocolRtmp:
			port = 1935
		case model.CameraProtocolRtsp:
			port = 554
		}
	}

	path := c.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if c.Username != "" && c.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d%s", c.Protocol, c.Username, c.Password, c.Ip, port, path)
	}
	return fmt.Sprintf("%s://%s:%d%s", c.Protocol, c.Ip, port, path)
}

func FromCameraModel(m *model.Camera) (*CameraSpec, error) {
	if m == nil {
		return nil, nil
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
	if m.BindDeviceId != 0 {
		dev, err := m.BindDevice()
		if err != nil {
			return nil, err
		}
		c.BindDevice = FromDeviceModel(dev)
	}
	return c, nil
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
	Name         string               `json:"name" binding:"required"`
	Protocol     model.CameraProtocol `json:"protocol" binding:"required"`
	Ip           string               `json:"ip" binding:"required"`
	Port         int                  `json:"port"`
	Path         string               `json:"path"`
	Username     string               `json:"username"`
	Password     string               `json:"password"`
	BindDeviceId int                  `json:"bindDeviceId"`
}

func (c *CreateCameraRequest) ToModel() *model.Camera {
	return &model.Camera{
		Uuid:         str.GenDeviceId(16),
		Name:         c.Name,
		Protocol:     c.Protocol,
		Ip:           c.Ip,
		Port:         c.Port,
		Path:         c.Path,
		Username:     c.Username,
		Password:     c.Password,
		BindDeviceId: c.BindDeviceId,
	}
}

type CreateCameraResponse struct {
	Uuid string `json:"uuid"`
}

type UpdateCameraRequest struct {
	Name         *string               `json:"name"`
	Protocol     *model.CameraProtocol `json:"protocol"`
	Ip           *string               `json:"ip"`
	Port         *int                  `json:"port"`
	Path         *string               `json:"path"`
	Username     *string               `json:"username"`
	Password     *string               `json:"password"`
	BindDeviceId *int                  `json:"bindDeviceId"`
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
	if req.BindDeviceId != nil {
		c.BindDeviceId = *req.BindDeviceId
	}
}
