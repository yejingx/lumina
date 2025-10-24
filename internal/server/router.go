package server

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) SetUpRouter() *gin.Engine {
	router := gin.New()
	router.Use(RequestId())
	router.Use(Logger())
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})
	// Serve static files from dashboard/build
	router.Static("/static", "./dashboard/build/static")
	router.NoRoute(func(c *gin.Context) {
		// API requests should 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		// All other routes go to index.html for client-side routing
		c.File("./dashboard/build/index.html")
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	apiV1 := router.Group("/api/v1")
	s.SetUpApiV1Router(apiV1)

	return router
}

func (s *Server) SetUpApiV1Router(apiV1 *gin.RouterGroup) {
	apiV1.POST("/login", s.handleLogin)
	apiV1.POST("/logout", s.handleLogout)

	device := apiV1.Group("/device")
	device.POST("/register", s.handleRegister)
	device.GET("", s.handleListDevices)
	device.GET("/:device_id", s.handleGetDevice)
	device.DELETE("/:device_id", s.handleDeleteDevice)

	deviceAuthed := device.Group("").Use(DeviceAuth())
	deviceAuthed.POST("/unregister", s.handleUnregister)
	deviceAuthed.GET("/jobs", s.handleGetDeviceJobs)

	accessToken := apiV1.Group("/access-token")
	accessToken.GET("", s.handleListAccessToken)
	accessToken.POST("", s.handleCreateAccessToken)
	accessToken.DELETE("/:token_id", s.handleDeleteAccessToken)
	accessToken.GET("/:token_id", s.handleGetAccessToken)

	workflow := apiV1.Group("/workflow")
	workflow.GET("", s.handleListWorkflows)
	workflow.POST("", s.handleCreateWorkflow)
	workflow.GET("/:workflow_id", s.handleGetWorkflow)
	workflow.PUT("/:workflow_id", s.handleUpdateWorkflow)
	workflow.DELETE("/:workflow_id", s.handleDeleteWorkflow)

	job := apiV1.Group("/job")
	job.Use(SetJobToContext())
	job.GET("", s.handleListJobs)
	job.POST("", s.handleCreateJob)
	job.GET("/:job_id", s.handleGetJob)
	job.PUT("/:job_id", s.handleUpdateJob)
	job.DELETE("/:job_id", s.handleDeleteJob)
	job.PUT("/:job_id/start", s.handleStartJob)
	job.PUT("/:job_id/stop", s.handleStopJob)
	job.GET("/:job_id/stats", s.handleJobStats)

	apiV1.GET("/message", s.handleListMessages)
	apiV1.POST("/message", s.handleCreateMessage)
	message := apiV1.Group("/message/:message_id")
	message.Use(SetMessageToContext())
	message.GET("", s.handleGetMessage)
	message.DELETE("", s.handleDeleteMessage)

	apiV1.GET("/conversation", s.handleListConversations)
	apiV1.POST("/conversation", s.handleCreateConversation)
	conversation := apiV1.Group("/conversation/:uuid")
	conversation.Use(SetConversationToContext())
	conversation.GET("", s.handleGetConversation)
	conversation.DELETE("", s.handleDeleteConversation)
	conversation.GET("/message", s.handleListChatMessages)
	conversation.POST("/chat", s.handleChat)
	conversation.POST("/title", s.handleGenChatTitle)

	v1Authed := apiV1.Group("")
	// v1Authed.Use(NeedAuth(false))

	v1UserSettings := v1Authed.Group("/settings")
	v1UserSettings.GET("/profile", s.handleGetUserProfile)

	{
		v1Admin := v1Authed.Group("/admin")

		v1Admin.GET("/users", s.handleAdminListUsers)
		v1Admin.POST("/users", s.handleAdminCreateUsers)
		v1Admin.DELETE("/user/:user_id", s.handleAdminDeleteUser)
	}
}
