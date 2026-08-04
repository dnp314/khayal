package main

import (
	"khayal/internal/controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

func createRouter() *gin.Engine {
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	router.Group("/v1/view")

	// passing user as context
	router.GET("/get_questions", controller.GetQuestion)
	router.GET("/get_questions/:id", controller.GetQuestion)
	router.POST("/create_question", controller.CreateQuestion)

	return router
}
