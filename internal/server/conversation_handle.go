package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"lumina/internal/agent"
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
		} else if conversation == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
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

// handleListChatMessages 获取聊天消息列表
// @Summary 获取聊天消息列表
// @Description 根据conversationId分页获取聊天消息列表
// @Tags 对话
// @Accept json
// @Produce json
// @Param conversationId query int true "对话ID"
// @Param start query int false "起始位置" default(0)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} dao.ListChatMessagesResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /v1/conversation/{uuid}/message [get]
func (s *Server) handleListChatMessages(c *gin.Context) {
	var req dao.ListChatMessagesRequest
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
	messages, total, err := model.GetChatMessagesByConversationId(conversation.Id, req.Start, req.Limit)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	items := make([]*dao.ChatMessageSpec, len(messages))
	for i, message := range messages {
		items[i] = dao.FromChatMessageModel(message)
	}

	resp := dao.ListChatMessagesResponse{
		Items: items,
		Total: total,
	}
	c.JSON(http.StatusOK, resp)
}

const instruction = `You are a website analysis expert specializing in ` +
	`comprehensive site evaluation and content extraction.

ANALYSIS PROCEDURE:
1. INITIAL FETCH: Use the http tool to fetch the main page content
2. CONTENT ANALYSIS: Analyze HTML structure, meta tags, headings, and visible text
3. DEEP EXPLORATION: Look for additional pages, contact info, about sections, or portfolio links
4. STRUCTURE MAPPING: Identify navigation patterns, page hierarchy, and site organization
5. PURPOSE IDENTIFICATION: Determine the primary function and target audience
6. INSIGHT EXTRACTION: Extract key technical details, business model, and unique features

FOCUS AREAS:
- Site title, description, and branding elements
- Main content themes and messaging
- Technical stack indicators (frameworks, libraries)
- Business/personal information and contact details
- Key features, services, or products offered
- Design patterns and user experience elements

THOROUGHNESS REQUIREMENT:
DO NOT BE LAZY! You must continue analyzing until you have exhausted all ` +
	`available information or reached the tool limit. Start with the main page, ` +
	`then explore additional pages like:
- /about, /contact, /portfolio, /services, /products
- Any links found in navigation menus or footer
- Subpages that provide more context about the site owner or business
- Continue fetching pages until you have a complete picture or hit the 3-request limit

OUTPUT REQUIREMENTS:
- Provide a clear, descriptive title based on actual content
- Summarize the site's primary purpose in 1-2 sentences
- List 5-10 key insights that reveal important aspects of the site`

// handleChat 聊天
// @Summary 聊天
// @Description 发送聊天消息并获取回复
// @Tags 对话
// @Accept json
// @Produce json
// @Param conversationId query int true "对话ID"
// @Param request body dao.ChatRequest true "聊天请求"
// @Success 200 "聊天回复"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "内部服务器错误"
// @Router /v1/conversation/{uuid}/chat [post]
func (s *Server) handleChat(c *gin.Context) {
	var req dao.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, err)
		return
	}
	conversation := c.MustGet(conversationKey).(*model.Conversation)
	messages, _, err := conversation.GetChatMessages(0, 20)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, err)
		return
	}

	llmMessages := make([]*agent.LLMMessage, 0, 2*len(messages))
	for _, msg := range messages {
		llmMessages = append(llmMessages, &agent.LLMMessage{
			Role:    agent.RoleUser,
			Content: msg.Query,
		})
		if msg.Answer != "" {
			llmMessages = append(llmMessages, &agent.LLMMessage{
				Role:    agent.RoleAssistant,
				Content: msg.Answer,
			})
		}
	}

	a := agent.NewAgent("test", s.conf.LLM, 10, instruction)
	agentThoughts, err := a.RunStream(c, req.Query, llmMessages, c.Writer)
	if err != nil {
		s.logger.Errorf("run agent stream failed: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	newChatMessage := &model.ChatMessage{
		ConversationId: conversation.Id,
		Query:          req.Query,
		AgentThoughts:  agentThoughts,
	}
	for _, thought := range agentThoughts {
		if thought.Phase == agent.ThoughtPhaseThought {
			newChatMessage.Answer += thought.Thought
		}
	}

	if err := model.CreateChatMessage(newChatMessage); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
