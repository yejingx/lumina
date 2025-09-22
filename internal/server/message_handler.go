package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lumina/internal/dao"
	"lumina/internal/model"
)

const messageKey = "message"

func SetMessageToContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		messageIdStr := c.Param("message_id")
		if messageIdStr == "" {
			c.Next()
			return
		}

		messageId, err := strconv.Atoi(messageIdStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid message_id",
			})
			return
		}

		message, err := model.GetMessage(messageId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		} else if message == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "message not found",
			})
			return
		}
		c.Set(messageKey, message)
		c.Next()
	}
}

// handleCreateMessage 创建消息
// @Summary 创建消息
// @Description 创建消息
// @Tags 消息
// @Accept json
// @Produce json
// @Param req body dao.CreateMessageRequest true "创建消息请求"
// @Success 200 {object} dao.CreateMessageResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/message [post]
func (s *Server) handleCreateMessage(c *gin.Context) {
	var req dao.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	message := req.ToModel()

	if err := model.AddMessage(message); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	resp := dao.CreateMessageResponse{
		Id: message.Id,
	}
	c.JSON(http.StatusOK, resp)
}

// handleGetMessage 获取消息
// @Summary 获取消息
// @Description 根据message_id获取消息详情
// @Tags 消息
// @Accept json
// @Produce json
// @Param message_id path string true "消息message_id"
// @Success 200 {object} dao.MessageSpec "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "消息不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/message/{message_id} [get]
func (s *Server) handleGetMessage(c *gin.Context) {
	message := c.MustGet(messageKey).(*model.Message)

	spec := dao.FromMessageModel(message)

	c.JSON(http.StatusOK, spec)
}

// handleDeleteMessage 删除消息
// @Summary 删除消息
// @Description 根据message_id删除消息
// @Tags 消息
// @Accept json
// @Produce json
// @Param message_id path string true "消息message_id"
// @Success 200 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "消息不存在"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/message/{message_id} [delete]
func (s *Server) handleDeleteMessage(c *gin.Context) {
	message := c.MustGet(messageKey).(*model.Message)

	if err := model.DeleteMessage(message.Id); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// handleListMessages 获取消息列表
// @Summary 获取消息列表
// @Description 根据jobId分页获取消息列表
// @Tags 消息
// @Accept json
// @Produce json
// @Param jobId query int true "任务ID"
// @Param start query int false "起始位置" default(0)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} dao.ListMessagesResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /api/v1/message [get]
func (s *Server) handleListMessages(c *gin.Context) {
	req := &dao.ListMessagesRequest{}
	if err := c.ShouldBindQuery(req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	messages, err := model.GetMessagesByJobId(req.JobId, req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	total, err := model.CountMessagesByJobId(req.JobId)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	items := make([]dao.MessageSpec, len(messages))
	for i, message := range messages {
		items[i] = *dao.FromMessageModel(message)
	}

	resp := dao.ListMessagesResponse{
		Items: items,
		Total: total,
	}
	c.JSON(http.StatusOK, resp)
}
