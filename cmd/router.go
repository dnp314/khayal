package main

import (
	"net/http"

	"github.com/dnp314/khayal/internal/controller"
	"github.com/gin-gonic/gin"
)

func createRouter(c Controllers, auth gin.HandlerFunc) *gin.Engine {

	router := gin.New()

	// router.Use(middleware.RequestID())

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	router.Static("/static", "./static")

	router.GET("/login", controller.ShowLogin)
	router.POST("/login", c.Auth.Login)
	router.GET("/register", controller.ShowRegister)
	router.POST("/register", c.Auth.PostRegister)

	authenticated := router.Group("/")
	authenticated.Use(auth)
	{
		authenticated.GET("/", c.Questions.Home)
		authenticated.GET("/questions/:id", c.Questions.GetQuestion)
		authenticated.GET("/questions", c.Questions.GetAllQuestions)
		authenticated.POST("/questions", c.Questions.CreateQuestion)
		authenticated.POST("/questions/:questionID/khayals", c.Khayals.CreateKhayal)
		authenticated.POST("/logout", c.Auth.Logout)
		// authenticated.GET("/questions/:questionID/khayals/:khayalID", c.Khayals.GetKhayal)
	}

	return router
}
