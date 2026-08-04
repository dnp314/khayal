package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"khayal/internal/components"

	"github.com/gin-gonic/gin"
)

func CreateQuestion(c *gin.Context) {
	c.String(http.StatusOK, "created")
}
func GetQuestion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		fmt.Printf("What ?")
	}
	component := components.Hello(id)
	component.Render(context.Background(), os.Stdout)
}
