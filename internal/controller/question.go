package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dnp314/khayal/internal/entities"

	"github.com/dnp314/khayal/internal/components"

	"github.com/gin-gonic/gin"
)

type QuestionService interface {
	GetQuestion(ctx context.Context, id int) (entities.Question, error)
	GetAllQuestions(ctx context.Context) ([]entities.Question, error)
	CreateQuestion(ctx context.Context, question string) error
}

type QuestionController struct {
	questionService QuestionService
	khayalService   KhayalService
}

func NewQuestionController(questionService QuestionService, khayalService KhayalService) *QuestionController {
	return &QuestionController{
		questionService: questionService,
		khayalService:   khayalService,
	}
}

func (c *QuestionController) Home(ctx *gin.Context) {
	questions, err := c.questionService.GetAllQuestions(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	err = components.HomePage(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
	if err != nil {
		ctx.Error(err)
	}
}

func (c *QuestionController) CreateQuestion(ctx *gin.Context) {

	question := ctx.PostForm("question")
	err := c.questionService.CreateQuestion(ctx, question)
	if err != nil {
		if err.Error() == "title cannot be empty" {
			fmt.Println("Question can't be empty")
		} else {
			fmt.Println("Error", err)
			return
		}
	}

	questions, _ := c.questionService.GetAllQuestions(ctx)

	_ = components.Questions(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)

}

func (c *QuestionController) GetQuestion(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}
	question, err := c.questionService.GetQuestion(ctx.Request.Context(), id)
	if err != nil {
		fmt.Println(err)
	}
	khayals, _ := c.khayalService.GetAllKhayals(ctx, id)

	fmt.Println(question)
	_ = components.QuestionPage(question, khayals).Render(
		ctx,
		ctx.Writer,
	)
}

func (c *QuestionController) GetAllQuestions(ctx *gin.Context) {

	questions, _ := c.questionService.GetAllQuestions(ctx)
	_ = components.Questions(questions).Render(
		ctx,
		ctx.Writer,
	)
}
