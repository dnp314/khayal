package controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnp314/khayal/internal/components"
	"github.com/dnp314/khayal/internal/entities"

	"github.com/gin-gonic/gin"
)

type KhayalService interface {
	// 	GetKhayal(ctx context.Context, id int) (entities.Khayal, error)
	GetAllKhayals(ctx context.Context, questionID int) ([]entities.Khayal, error)
	CreateKhayal(ctx context.Context, questionID int, khayal string) error
}

type KhayalController struct {
	khayalService   KhayalService
	questionService QuestionService
}

func NewKhayalController(khayalService KhayalService, questionService QuestionService) *KhayalController {
	return &KhayalController{
		khayalService:   khayalService,
		questionService: questionService,
	}
}

// func (c *KhayalController) Home(ctx *gin.Context) {
// 	khayals, err := c.service.GetAllKhayals(ctx)
// 	if err != nil {
// 		ctx.Error(err)
// 		return
// 	}
// 	err = components.Home(khayals).Render(
// 		ctx.Request.Context(),
// 		ctx.Writer,
// 	)
// 	if err != nil {
// 		ctx.Error(err)
// 	}
// }

func (c *KhayalController) CreateKhayal(ctx *gin.Context) {
	khayal := ctx.PostForm("khayal")
	questionID, _ := strconv.Atoi(ctx.Param("questionID"))
	err := c.khayalService.CreateKhayal(ctx, questionID, khayal)
	if err != nil {
		if err.Error() == "title cannot be empty" {
			fmt.Println("Question can't be empty")
		} else {
			fmt.Println("Error", err)
			return
		}
	}

	khayals, _ := c.khayalService.GetAllKhayals(ctx, questionID)
	_ = components.Khayals(khayals).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)

	questions, _ := c.questionService.GetAllQuestions(ctx)
	_ = components.Questions(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
}

// func (c *KhayalController) GetKhayal(ctx *gin.Context) {
// 	id, err := strconv.Atoi(ctx.Param("id"))
// 	if err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"error": "invalid id",
// 		})
// 		return
// 	}
// 	khayal, err := c.service.GetKhayal(ctx.Request.Context(), id)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	fmt.Println(khayal)
// 	_ = components.Khayal(khayal).Render(
// 		ctx,
// 		ctx.Writer,
// 	)
// }

// func (c *KhayalController) GetAllKhayals(ctx *gin.Context) {

// 	khayals, _ := c.service.GetAllKhayals(ctx)
// 	_ = components.Khayals(khayals).Render(
// 		ctx,
// 		ctx.Writer,
// 	)
// }
