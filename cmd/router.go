package main

import (
	"net/http"

	"github.com/dnp314/khayal/internal/controller"
	"github.com/gin-gonic/gin"
)

func createRouter(c Controllers) *gin.Engine {

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
	// passing user as context
	//	router.POST("/create_question", controller.CreateQuestion)
	router.GET("/", c.Questions.Home)
	router.GET("/questions/:id", c.Questions.GetQuestion)
	router.GET("/questions", c.Questions.GetAllQuestions)
	router.POST("/questions", c.Questions.CreateQuestion)
	router.POST("/questions/:questionID/khayals", c.Khayals.CreateKhayal)

	return router
}
