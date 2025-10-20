package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lumina/internal/dao"
	"lumina/internal/model"
)

const jobKey = "job"

func SetJobToContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobIdStr := c.Param("job_id")
		if jobIdStr == "" {
			c.Next()
			return
		}

		jobId, err := strconv.Atoi(jobIdStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid job_id",
			})
			return
		}

		job, err := model.GetJobById(jobId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		} else if job == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "job not found",
			})
			return
		}
		c.Set(jobKey, job)
		c.Next()
	}
}

// handleCreateJob 创建任务
// @Summary 创建任务
// @Description 创建任务
// @Tags 任务
// @Accept json
// @Produce json
// @Param req body dao.CreateJobRequest true "创建任务请求"
// @Success 200 {object} dao.CreateJobResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job [post]
func (s *Server) handleCreateJob(c *gin.Context) {
	var req dao.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	job := req.ToModel()

	if err := model.AddJob(job); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.CreateJobResponse{
		Uuid: job.Uuid,
	}
	c.JSON(http.StatusOK, resp)
}

// handleGetJob 获取任务
// @Summary 获取任务
// @Description 根据job_id获取任务详情
// @Tags 任务
// @Accept json
// @Produce json
// @Param job_id path string true "任务job_id"
// @Success 200 {object} dao.JobSpec "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job/{job_id} [get]
func (s *Server) handleGetJob(c *gin.Context) {
	job := c.MustGet(jobKey).(*model.Job)

	spec, err := dao.FromJobModel(job)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, spec)
}

// handleUpdateJob 更新任务
// @Summary 更新任务
// @Description 根据job_id更新任务信息
// @Tags 任务
// @Accept json
// @Produce json
// @Param job_id path string true "任务job_id"
// @Param req body dao.UpdateJobRequest true "更新任务请求"
// @Success 200 "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job/{job_id} [put]
func (s *Server) handleUpdateJob(c *gin.Context) {
	var req dao.UpdateJobRequest
	if err2 := c.ShouldBindJSON(&req); err2 != nil {
		s.writeError(c, http.StatusBadRequest, err2)
		return
	}

	job := c.MustGet(jobKey).(*model.Job)

	req.UpdateModel(job)

	if err := model.UpdateJob(job); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// handleDeleteJob 删除任务
// @Summary 删除任务
// @Description 根据job_id删除任务
// @Tags 任务
// @Accept json
// @Produce json
// @Param job_id path string true "任务job_id"
// @Success 200 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job/{job_id} [delete]
func (s *Server) handleDeleteJob(c *gin.Context) {
	job := c.MustGet(jobKey).(*model.Job)

	if err := model.DeleteJob(job); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// handleListJobs 获取任务列表
// @Summary 获取任务列表
// @Description 分页获取任务列表
// @Tags 任务
// @Accept json
// @Produce json
// @Param start query int false "起始位置" default(0)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} dao.ListJobsResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job [get]
func (s *Server) handleListJobs(c *gin.Context) {
	req := &dao.ListJobsRequest{}
	if err := c.ShouldBindQuery(req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	// Set default values if not provided
	if req.Limit == 0 {
		req.Limit = 10
	}

	jobs, total, err := model.ListJobs(req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	items := make([]dao.JobSpec, len(jobs))
	for i, job := range jobs {
		spec, err := dao.FromJobModel(&job)
		if err != nil {
			s.writeError(c, http.StatusInternalServerError, err)
			return
		}
		items[i] = *spec
	}

	resp := dao.ListJobsResponse{
		Items: items,
		Total: total,
	}
	c.JSON(http.StatusOK, resp)
}

// handleStartJob 启动任务
// @Summary 启动任务
// @Description 根据job_id启动任务
// @Tags 任务
// @Accept json
// @Produce json
// @Param job_id path string true "任务job_id"
// @Success 200 "启动成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job/{job_id}/start [put]
func (s *Server) handleStartJob(c *gin.Context) {
	job := c.MustGet(jobKey).(*model.Job)
	job.Status = model.JobStatusRunning
	if err := model.UpdateJob(job); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// handleStopJob 停止任务
// @Summary 停止任务
// @Description 根据job_id停止任务
// @Tags 任务
// @Accept json
// @Produce json
// @Param job_id path string true "任务job_id"
// @Success 200 "停止成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "任务不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/job/{job_id}/stop [put]
func (s *Server) handleStopJob(c *gin.Context) {
	job := c.MustGet(jobKey).(*model.Job)
	job.Status = model.JobStatusStopped
	if err := model.UpdateJob(job); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
