package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// handleRegister 注册设备
// @Summary 注册设备
// @Description 注册设备
// @Tags 设备
// @Accept json
// @Produce json
// @Param req body dao.RegisterRequest true "注册请求"
// @Success 200 {object} dao.RegisterResponse "注册成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 409 {object} ErrorResponse "冲突"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/agent/register [post]
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

// handleUnregister 注销设备
// @Summary 注销设备
// @Description 注销设备
// @Tags 设备
// @Accept json
// @Produce json
// @Success 200 "注销成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/agent/unregister [post]
func (s *Server) handleUnregister(c *gin.Context) {
	agent := c.MustGet(agentKey).(*model.Device)
	if err := agent.Unregister(); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// handleCreateAccessToken 创建访问令牌
// @Summary 创建访问令牌
// @Description 创建访问令牌
// @Tags 设备
// @Accept json
// @Produce json
// @Param req body dao.CreateAccessTokenRequest true "创建访问令牌请求"
// @Success 200 {object} dao.CreateAccessTokenResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/access-token [post]
func (s *Server) handleCreateAccessToken(c *gin.Context) {
	var req dao.CreateAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	if req.ExpireTime == "" {
		req.ExpireTime = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	}
	expireTime, err := time.Parse(time.RFC3339, req.ExpireTime)
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	token := str.GenDeviceId(16)
	accessToken := &model.AccessToken{
		AccessToken: str.RandStr(16, str.UpperAlphabet+str.Numerals),
		ExpireTime:  expireTime,
	}
	if err := model.CreateAccessToken(accessToken); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	resp := dao.CreateAccessTokenResponse{
		AccessToken: token,
	}
	c.JSON(http.StatusOK, resp)
}

// handleDeleteAccessToken 删除访问令牌
// @Summary 删除访问令牌
// @Description 删除访问令牌
// @Tags 设备
// @Accept json
// @Produce json
// @Param token_id path int true "访问令牌ID"
// @Success 200 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/access-token/{token_id} [delete]
func (s *Server) handleDeleteAccessToken(c *gin.Context) {
	tokenId, err := strconv.Atoi(c.Param("token_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	if err := model.DeleteAccessToken(uint(tokenId)); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// handleListAccessToken 列出访问令牌
// @Summary 列出访问令牌
// @Description 列出访问令牌
// @Tags 设备
// @Accept json
// @Produce json
// @Param start query int true "分页起始位置"
// @Param limit query int true "分页每页数量"
// @Success 200 {object} dao.ListAccessTokenResponse "列出成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/access-token [get]
func (s *Server) handleListAccessToken(c *gin.Context) {
	var req dao.ListAccessTokenRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	accessTokens, total, err := model.ListAccessToken(req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.ListAccessTokenResponse{
		AccessTokens: make([]dao.AccessTokenSpec, 0, len(accessTokens)),
		Total:        total,
	}
	for _, t := range accessTokens {
		spec := dao.FromAccessTokenModel(&t)
		resp.AccessTokens = append(resp.AccessTokens, *spec)
	}
	c.JSON(http.StatusOK, resp)
}

// handleGetAccessToken 获取访问令牌
// @Summary 获取访问令牌
// @Description 获取访问令牌
// @Tags 设备
// @Accept json
// @Produce json
// @Param token_id path int true "访问令牌ID"
// @Success 200 {object} dao.AccessTokenSpec "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/access-token/{token_id} [get]
func (s *Server) handleGetAccessToken(c *gin.Context) {
	tokenId, err := strconv.Atoi(c.Param("token_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	accessToken, err := model.GetAccessToken(tokenId)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if accessToken == nil {
		s.writeError(c, http.StatusNotFound, errors.New("access token not found"))
		return
	}
	spec := dao.FromAccessTokenModel(accessToken)
	c.JSON(http.StatusOK, spec)
}

// handleGetDevice 获取设备
// @Summary 获取设备
// @Description 获取设备
// @Tags 设备
// @Accept json
// @Produce json
// @Param device_id path int true "设备ID"
// @Success 200 {object} dao.DeviceSpec "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/device/{device_id} [get]
func (s *Server) handleGetDevice(c *gin.Context) {
	deviceId, err := strconv.Atoi(c.Param("device_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	device, err := model.GetDeviceById(deviceId)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if device == nil {
		s.writeError(c, http.StatusNotFound, errors.New("device not found"))
		return
	}
	spec := dao.FromDeviceModel(device)
	c.JSON(http.StatusOK, spec)
}

// handleListDevices 列出设备
// @Summary 列出设备
// @Description 列出设备
// @Tags 设备
// @Accept json
// @Produce json
// @Param start query int true "分页起始位置"
// @Param limit query int true "分页每页数量"
// @Success 200 {object} dao.ListDeviceResponse "列出成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/device [get]
func (s *Server) handleListDevices(c *gin.Context) {
	var req dao.ListDeviceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	devices, total, err := model.ListDevices(req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.ListDeviceResponse{
		Devices: make([]dao.DeviceSpec, 0, len(devices)),
		Total:   total,
	}
	for _, d := range devices {
		spec := dao.FromDeviceModel(&d)
		resp.Devices = append(resp.Devices, *spec)
	}
	c.JSON(http.StatusOK, resp)
}

// handleDeleteDevice 删除设备
// @Summary 删除设备
// @Description 删除设备
// @Tags 设备
// @Accept json
// @Produce json
// @Param device_id path int true "设备ID"
// @Success 200 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/device/{device_id} [delete]
func (s *Server) handleDeleteDevice(c *gin.Context) {
	deviceId, err := strconv.Atoi(c.Param("device_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	if err := model.DeleteDevice(uint(deviceId)); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
