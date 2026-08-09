package controller

import (
	"fmt"
	"khayal/internal/components"

	"github.com/gin-gonic/gin"
)

var questions []string

func CreateQuestion(c *gin.Context) {

	question := c.PostForm("question")
	questions = append(questions, question)
	fmt.Println(questions)
	fmt.Println("question received", question)
	_ = components.Question(questions).Render(
		c.Request.Context(),
		c.Writer,
	)
}

func Home(c *gin.Context) {

	err := components.Home().Render(
		c.Request.Context(),
		c.Writer,
	)
	if err != nil {
		c.Error(err)
	}
}
