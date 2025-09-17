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

	apiV1.POST("/register", s.handleRegister)
	agentAuthed := apiV1.Group("")
	agentAuthed.Use(AgentAuth())
	agentAuthed.POST("/unregister", s.handleUnregister)

	v1Authed := apiV1.Group("")
	v1Authed.Use(NeedAuth(false))

	v1UserSettings := v1Authed.Group("/settings")
	v1UserSettings.GET("/profile", s.handleGetUserProfile)

	{
		v1Admin := v1Authed.Group("/admin")

		v1Admin.GET("/users", s.handleAdminListUsers)
		v1Admin.POST("/users", s.handleAdminCreateUsers)
		v1Admin.DELETE("/user/:user_id", s.handleAdminDeleteUser)
	}
}
