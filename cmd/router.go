package main

import (
	"net/http"

	"github.com/dnp314/khayal/internal/controller"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/gin-gonic/gin"
)

func createRouter(c Controllers, sessionService middleware.SessionService) *gin.Engine {

	router := gin.New()
	router.Use(middleware.SecurityHeaders())

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	router.Static("/static", "./static")

	router.GET("/login", controller.ShowLogin)
	router.POST("/login", c.Auth.Login)
	router.POST("/login/legacy", c.Auth.LegacyLogin)
	router.GET("/register", controller.ShowRegister)
	router.POST("/register", c.Auth.Register)
	router.GET("/recover", controller.ShowRecover)
	router.POST("/auth/recover", c.Auth.Recover)
	router.GET("/auth/params", c.Auth.AuthParams)
	router.GET("/auth/recovery/params", c.Auth.RecoveryParams)
	router.GET("/", middleware.OptionalAuth(sessionService), c.Questions.Home)

	authenticated := router.Group("/")
	authenticated.Use(middleware.Auth(sessionService))
	{
		authenticated.GET("/account", c.Auth.Account)
		authenticated.GET("/auth/keys", c.Auth.UserKeys)
		authenticated.POST("/auth/password", c.Auth.ChangePassword)
		authenticated.POST("/auth/legacy/setup", c.Auth.SetupLegacy)
		authenticated.GET("/auth/legacy/rows", c.Auth.LegacyRows)
		authenticated.POST("/auth/legacy/import", c.Auth.ImportLegacyRows)
		authenticated.GET("/questions/:id", c.Questions.GetQuestion)
		authenticated.GET("/questions", c.Questions.GetAllQuestions)
		authenticated.POST("/questions", c.Questions.CreateQuestion)
		authenticated.POST("/questions/:questionID/khayals", c.Khayals.CreateKhayal)
		authenticated.POST("/logout", c.Auth.Logout)
	}

	return router
}
