package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"lumina/internal/dao"
	"lumina/internal/model"
)

const cameraKey = "camera"

func SetCameraToContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		cameraIdStr := c.Param("camera_id")
		if cameraIdStr == "" {
			c.Next()
			return
		}

		cameraId, err := strconv.Atoi(cameraIdStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid camera_id",
			})
			return
		}

		camera, err := model.GetCameraById(cameraId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		} else if camera == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "camera not found",
			})
			return
		}
		c.Set(cameraKey, camera)
		c.Next()
	}
}

// handleCreateCamera 创建摄像头
// @Summary 创建摄像头
// @Description 创建摄像头
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param req body dao.CreateCameraRequest true "创建摄像头请求"
// @Success 200 {object} dao.CreateCameraResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera [post]
func (s *Server) handleCreateCamera(c *gin.Context) {
	var req dao.CreateCameraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	cam := req.ToModel()
	if err := model.CreateCamera(cam); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.CreateCameraResponse{Uuid: cam.Uuid}
	c.JSON(http.StatusOK, resp)
}

// handleGetCamera 获取摄像头
// @Summary 获取摄像头
// @Description 获取摄像头
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param camera_id path int true "摄像头ID"
// @Success 200 {object} dao.CameraSpec "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "摄像头不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera/{camera_id} [get]
func (s *Server) handleGetCamera(c *gin.Context) {
	cam := c.MustGet(cameraKey).(*model.Camera)

	spec, err := dao.FromCameraModel(cam)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, spec)
}

// handleUpdateCamera 更新摄像头
// @Summary 更新摄像头
// @Description 更新摄像头
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param camera_id path int true "摄像头ID"
// @Param req body dao.UpdateCameraRequest true "更新摄像头请求"
// @Success 200 "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "摄像头不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera/{camera_id} [put]
func (s *Server) handleUpdateCamera(c *gin.Context) {
	var req dao.UpdateCameraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	cam := c.MustGet(cameraKey).(*model.Camera)

	req.UpdateModel(cam)
	if err := model.UpdateCamera(cam); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// handleDeleteCamera 删除摄像头
// @Summary 删除摄像头
// @Description 删除摄像头
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param camera_id path int true "摄像头ID"
// @Success 200 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera/{camera_id} [delete]
func (s *Server) handleDeleteCamera(c *gin.Context) {
	cam := c.MustGet(cameraKey).(*model.Camera)

	if err := model.DeleteCamera(cam); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// handleListCameras 列出摄像头
// @Summary 列出摄像头
// @Description 列出摄像头
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param start query int true "分页起始位置"
// @Param limit query int true "分页每页数量"
// @Success 200 {object} dao.ListCamerasResponse "列出成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera [get]
func (s *Server) handleListCameras(c *gin.Context) {
	var req dao.ListCamerasRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	items, total, err := model.ListCameras(req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.ListCamerasResponse{
		Items: make([]dao.CameraSpec, 0, len(items)),
		Total: total,
	}
	for _, cam := range items {
		camSpec, err := dao.FromCameraModel(&cam)
		if err != nil {
			s.writeError(c, http.StatusInternalServerError, err)
			return
		}
		resp.Items = append(resp.Items, *camSpec)
	}
	c.JSON(http.StatusOK, resp)
}

func genPreviewAddr(serverIp string, serverPort int, taskUuid string) string {
	// 新的预览地址，需要添加.live.flv后缀
	// 老的预览地址，添加 .flv 后缀
	return fmt.Sprintf("http://%s:%d/preview/%s.live.flv", serverIp, serverPort, taskUuid)
}

func genPushAddr(serverIp string, serverPort int, taskUuid string) string {
	return fmt.Sprintf("rtmp://%s:%d/preview/%s", serverIp, serverPort, taskUuid)
}

// handleStartCameraPreview 开始摄像头预览
// @Summary 开始摄像头预览
// @Description 开始摄像头预览
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param camera_id path int true "摄像头ID"
// @Success 200 {object} dao.PreviewTask "预览任务"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "摄像头不存在"
// @Failure 409 {object} ErrorResponse "预览任务已存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera/{camera_id}/preview [post]
func (s *Server) handleStartCameraPreview(c *gin.Context) {
	cam := c.MustGet(cameraKey).(*model.Camera)
	camSpec, err := dao.FromCameraModel(cam)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	device, err := cam.BindDevice()
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	task, err := model.GetPreviewTask(c, device.Uuid, cam.Uuid)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if task != nil {
		previewAddr := genPreviewAddr(s.conf.MediaServer.Ip, s.conf.MediaServer.HttpPort, task.TaskUuid)
		task.ExpireTime = time.Now().Add(15 * time.Minute)
		if err := model.AddPreviewTask(c, device.Uuid, cam.Uuid, task); err != nil {
			s.writeError(c, http.StatusInternalServerError, err)
			return
		}
		resp := dao.FromPreviewTaskModel(task)
		resp.PreviewAddr = previewAddr
		c.JSON(http.StatusOK, resp)
		return
	}

	taskUuid := uuid.New().String()
	pushAddr := genPushAddr(s.conf.MediaServer.Ip, s.conf.MediaServer.RtmpPort, taskUuid)
	task = &model.PreviewTask{
		TaskUuid:   taskUuid,
		ExpireTime: time.Now().Add(15 * time.Minute),
		PullAddr:   camSpec.Url(),
		PushAddr:   pushAddr,
	}

	if err := model.AddPreviewTask(c, device.Uuid, cam.Uuid, task); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	previewAddr := genPreviewAddr(s.conf.MediaServer.Ip, s.conf.MediaServer.HttpPort, taskUuid)
	resp := dao.FromPreviewTaskModel(task)
	resp.PreviewAddr = previewAddr
	c.JSON(http.StatusOK, resp)
}

// handleTouchCameraPreview 刷新摄像头预览任务过期时间
// @Summary 刷新摄像头预览任务过期时间
// @Description 刷新摄像头预览任务过期时间
// @Tags 摄像头
// @Accept json
// @Produce json
// @Param camera_id path int true "摄像头ID"
// @Success 200 "刷新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "摄像头不存在"
// @Failure 409 {object} ErrorResponse "预览任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/camera/{camera_id}/preview [put]
func (s *Server) handleTouchCameraPreview(c *gin.Context) {
	cam := c.MustGet(cameraKey).(*model.Camera)
	device, err := cam.BindDevice()
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	if err := model.TouchPreviewTask(c, device.Uuid, cam.Uuid); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
