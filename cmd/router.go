package main

import (
	"net/http"

	"github.com/dnp314/khayal/internal/controller"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/gin-gonic/gin"
)

func createRouter(c Controllers, sessionService middleware.SessionService) *gin.Engine {

	router := gin.New()

	// router.Use(middleware.RequestID())

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	router.Static("/static", "./static")

	router.Use(middleware.LoadSession(sessionService))

	router.GET("/login", controller.ShowLogin)
	router.POST("/login", c.Auth.Login)
	router.GET("/register", controller.ShowRegister)
	router.POST("/register", c.Auth.PostRegister)
	router.GET("/", c.Questions.Home)

	authenticated := router.Group("/")
	authenticated.Use(middleware.RequireAuth())
	{
		authenticated.GET("/questions/:questionID", c.Questions.GetQuestion)
		authenticated.GET("/questions/:questionID/:questionSlug", c.Questions.GetQuestion)
		authenticated.GET("/questions", c.Questions.GetAllQuestions)
		authenticated.POST("/questions", c.Questions.CreateQuestion)
		authenticated.POST("/questions/:questionID/khayals", c.Khayals.CreateKhayal)
		authenticated.GET("/questions/:questionID/khayals/:khayalID", c.Khayals.GetKhayal)
		authenticated.GET("/questions/:questionID/:questionSlug/khayals/:khayalID/:khayalSlug", c.Khayals.GetKhayal)
		authenticated.POST("/logout", c.Auth.Logout)
		// patch or put for updating question ?
	}

	return router
}
