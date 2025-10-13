package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"lumina/internal/dao"
	"lumina/internal/model"
)

const conversationKey = "conversation"

func SetConversationToContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		conversation, err := model.GetConversationByUuid(uuid)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.Set(conversationKey, conversation)
		c.Next()
	}
}

// handleCreateConversation 创建对话.
// @Summary 创建对话
// @Description 创建一个新的对话，标题为请求体中的title字段
// @Tags 对话
// @Accept json
// @Produce json
// @Param request body dao.CreateConversationRequest true "创建对话请求体"
// @Success 200 {object} dao.CreateConversationResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /v1/conversation [post]
func (s *Server) handleCreateConversation(c *gin.Context) {
	var req dao.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uuid := uuid.New().String()
	conversation := &model.Conversation{
		Uuid:  uuid,
		Title: req.Title,
	}
	if err := model.CreateConversation(conversation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dao.CreateConversationResponse{
		Uuid: conversation.Uuid,
	})
}

// handleListConversations 获取对话列表.
// @Summary 获取对话列表
// @Description 获取所有对话，支持分页
// @Tags 对话
// @Accept json
// @Produce json
// @Param start query int false "分页起始索引"
// @Param limit query int false "分页每页数量"
// @Success 200 {object} dao.ListConversationsResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /v1/conversation [get]
func (s *Server) handleListConversations(c *gin.Context) {
	var req dao.ListConversationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversations, total, err := model.ListConversations(req.Start, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*dao.ConversationSpec, 0, len(conversations))
	for _, conversation := range conversations {
		item, err := dao.FromConversationModel(conversation)
		if err != nil {
			s.writeError(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, dao.ListConversationsResponse{
		Items: items,
		Total: total,
	})
}

// handleGetConversation 获取对话.
// @Summary 获取对话
// @Description 获取指定UUID的对话
// @Tags 对话
// @Accept json
// @Produce json
// @Param uuid path string true "对话UUID"
// @Success 200 {object} dao.ConversationSpec
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /v1/conversation/{uuid} [get]
func (s *Server) handleGetConversation(c *gin.Context) {
	conversation := c.MustGet(conversationKey).(*model.Conversation)
	item, err := dao.FromConversationModel(conversation)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// handleDeleteConversation 删除对话.
// @Summary 删除对话
// @Description 删除指定UUID的对话
// @Tags 对话
// @Accept json
// @Produce json
// @Param uuid path string true "对话UUID"
// @Success 200 {object} map[string]string "message: conversation deleted"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /v1/conversation/{uuid} [delete]
func (s *Server) handleDeleteConversation(c *gin.Context) {
	conversation := c.MustGet(conversationKey).(*model.Conversation)
	uuid := conversation.Uuid
	if err := model.DeleteConversationByUuid(uuid); err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

// handleListLLMMessages 获取LLM消息列表
// @Summary 获取LLM消息列表
// @Description 根据conversationId分页获取LLM消息列表
// @Tags 对话
// @Accept json
// @Produce json
// @Param conversationId query int true "对话ID"
// @Param start query int false "起始位置" default(0)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} dao.ListLLMMessagesResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /v1/conversation/message [get]
func (s *Server) handleListLLMMessages(c *gin.Context) {
	var req dao.ListLLMMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	} else if req.Limit > 100 {
		req.Limit = 100
	}

	conversation := c.MustGet(conversationKey).(*model.Conversation)
	messages, total, err := model.GetLLMMessagesByConversationId(conversation.Id, req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	items := make([]*dao.LLMMessageSpec, len(messages))
	for i, message := range messages {
		items[i] = dao.FromLLMMessageModel(message)
	}

	resp := dao.ListLLMMessagesResponse{
		Items: items,
		Total: total,
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleChat(c *gin.Context) {
	var req dao.ChatMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	conversation := c.MustGet(conversationKey).(*model.Conversation)
	_, _, err := conversation.GetLLMMessages(0, 20)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}
}
