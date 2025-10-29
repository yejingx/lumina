package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lumina/internal/dao"
	"lumina/internal/model"
)

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
	id, err := strconv.Atoi(c.Param("camera_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	cam, err := model.GetCameraById(id)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if cam == nil {
		s.writeError(c, http.StatusNotFound, errors.New("camera not found"))
		return
	}

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
	id, err := strconv.Atoi(c.Param("camera_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	var req dao.UpdateCameraRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	cam, err := model.GetCameraById(id)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if cam == nil {
		s.writeError(c, http.StatusNotFound, errors.New("camera not found"))
		return
	}

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
	id, err := strconv.Atoi(c.Param("camera_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	cam, err := model.GetCameraById(id)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if cam == nil {
		s.writeError(c, http.StatusNotFound, errors.New("camera not found"))
		return
	}

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
