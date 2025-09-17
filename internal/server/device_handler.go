package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"lumina/internal/dao"
	"lumina/internal/model"
	"lumina/pkg/str"
)

const agentKey = "agent"

func genDeviceToken() string {
	return "agent-" + str.GenToken(20)
}

func AgentAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			auth := c.GetHeader("Authorization")
			if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
				tokenStr = auth[7:]
			}
		}
		if tokenStr == "" || !strings.HasPrefix(tokenStr, "agent-") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}
		agent, err := model.GetDeviceByToken(tokenStr)
		if err != nil || agent == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}
		c.Set(agentKey, agent)
		c.Next()
	}
}

func (s *Server) handleRegister(c *gin.Context) {
	var req dao.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	accessToken, err := model.GetAccessTokenByToken(req.AccessToken)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if accessToken == nil {
		s.writeError(c, http.StatusNotFound, errors.New("token not found"))
		return
	} else if accessToken.IsExpired() {
		s.writeError(c, http.StatusUnauthorized, errors.New("token expired"))
		return
	} else if accessToken.IsBound() {
		s.writeError(c, http.StatusConflict, errors.New("token already bound"))
		return
	}

	var device *model.Device
	if req.Uuid != "" {
		device, err = model.GetDeviceByUuid(req.Uuid)
		if err != nil {
			s.writeError(c, http.StatusInternalServerError, err)
			return
		} else if device == nil {
			s.writeError(c, http.StatusNotFound, errors.New("device not found"))
			return
		} else if device.IsRegistered() {
			s.writeError(c, http.StatusConflict, errors.New("device already registered"))
			return
		}
	} else {
		device = &model.Device{
			Uuid: str.GenDeviceId(16),
		}
	}
	device.Token = genDeviceToken()
	if err := accessToken.BindDevice(device); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.RegisterResponse{
		Uuid:  device.Uuid,
		Token: device.Token,
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleUnregister(c *gin.Context) {
	agent := c.MustGet(agentKey).(*model.Device)
	if err := agent.Unregister(); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
