package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lumina/internal/dao"
	"lumina/internal/model"
)

// handleCreateWorkflow 创建工作流
// @Summary 创建工作流
// @Description 创建工作流
// @Tags 工作流
// @Accept json
// @Produce json
// @Param req body dao.CreateWorkflowRequest true "创建工作流请求"
// @Success 200 {object} dao.CreateWorkflowResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/workflow [post]
func (s *Server) handleCreateWorkflow(c *gin.Context) {
	var req dao.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	workflow := req.ToModel()
	if err := model.CreateWorkflow(workflow); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.CreateWorkflowResponse{
		Id:   workflow.Id,
		Uuid: workflow.Uuid,
	}

	c.JSON(http.StatusOK, resp)
}

// handleGetWorkflow 获取工作流
// @Summary 获取工作流
// @Description 获取工作流
// @Tags 工作流
// @Accept json
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 {object} dao.WorkflowSpec "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "工作流不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/workflow/{workflow_id} [get]
func (s *Server) handleGetWorkflow(c *gin.Context) {
	workflowId, err := strconv.Atoi(c.Param("workflow_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	workflow, err := model.GetWorkflowById(workflowId)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if workflow == nil {
		s.writeError(c, http.StatusNotFound, errors.New("workflow not found"))
		return
	}

	spec := dao.FromWorkflowModel(workflow)
	c.JSON(http.StatusOK, spec)
}

// handleUpdateWorkflow 更新工作流
// @Summary 更新工作流
// @Description 更新工作流
// @Tags 工作流
// @Accept json
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Param req body dao.UpdateWorkflowRequest true "更新工作流请求"
// @Success 200 "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "工作流不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/workflow/{workflow_id} [put]
func (s *Server) handleUpdateWorkflow(c *gin.Context) {
	workflowId, err := strconv.Atoi(c.Param("workflow_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	var req dao.UpdateWorkflowRequest
	if err2 := c.ShouldBindJSON(&req); err2 != nil {
		s.writeError(c, http.StatusBadRequest, err2)
		return
	}

	workflow, err := model.GetWorkflowById(workflowId)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	} else if workflow == nil {
		s.writeError(c, http.StatusNotFound, errors.New("workflow not found"))
		return
	}
	req.UpdateModel(workflow)

	if err := model.UpdateWorkflow(workflow); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// handleDeleteWorkflow 删除工作流
// @Summary 删除工作流
// @Description 删除工作流
// @Tags 工作流
// @Accept json
// @Produce json
// @Param workflow_id path int true "工作流ID"
// @Success 200 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/workflow/{workflow_id} [delete]
func (s *Server) handleDeleteWorkflow(c *gin.Context) {
	workflowId, err := strconv.Atoi(c.Param("workflow_id"))
	if err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	if err := model.DeleteWorkflow(workflowId); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// handleListWorkflows 列出工作流
// @Summary 列出工作流
// @Description 列出工作流
// @Tags 工作流
// @Accept json
// @Produce json
// @Param start query int true "分页起始位置"
// @Param limit query int true "分页每页数量"
// @Success 200 {object} dao.ListWorkflowResponse "列出成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/workflow [get]
func (s *Server) handleListWorkflows(c *gin.Context) {
	var req dao.ListWorkflowRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	// Set default values if not provided
	if req.Limit == 0 {
		req.Limit = 10
	}

	workflows, total, err := model.ListWorkflows(req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.ListWorkflowResponse{
		Workflows: make([]dao.WorkflowSpec, 0, len(workflows)),
		Total:     total,
	}
	for _, wf := range workflows {
		spec := dao.FromWorkflowModel(&wf)
		resp.Workflows = append(resp.Workflows, *spec)
	}

	c.JSON(http.StatusOK, resp)
}
