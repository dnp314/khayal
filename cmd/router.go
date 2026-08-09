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

	router.GET("", controller.Home)
	// passing user as context
	//	router.POST("/create_question", controller.CreateQuestion)
	router.POST("/create_question", controller.CreateQuestion)

	return router
}
