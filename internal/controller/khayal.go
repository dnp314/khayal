package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dnp314/khayal/internal/components"
	"github.com/dnp314/khayal/internal/entities"
	"github.com/dnp314/khayal/internal/service"

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

func (c *KhayalController) CreateKhayal(ctx *gin.Context) {
	khayal := ctx.PostForm("khayal")
	questionID, err := strconv.Atoi(ctx.Param("questionID"))
	if err != nil {
		slog.Warn("invalid question id", "id", ctx.Param("questionID"), "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}

	err = c.khayalService.CreateKhayal(ctx, questionID, khayal)
	if err != nil {
		if errors.Is(err, service.ErrEmptyKhayal) {
			slog.Warn("reject empty khayal", "question_id", questionID)
			ctx.String(http.StatusBadRequest, "Khayal can't be empty")
			return
		}

		slog.Error("create khayal", "question_id", questionID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to create khayal")
		return
	}

	khayals, err := c.khayalService.GetAllKhayals(ctx, questionID)
	if err != nil {
		slog.Error("get khayals after creation", "question_id", questionID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load khayals")
		return
	}
	if err := components.Khayals(khayals).Render(
		ctx.Request.Context(),
		ctx.Writer,
	); err != nil {
		slog.Error("render khayals after creation", "question_id", questionID, "error", err)
		return
	}

	questions, err := c.questionService.GetAllQuestions(ctx)
	if err != nil {
		slog.Error("get questions after creating khayal", "question_id", questionID, "error", err)
		return
	}
	if err := components.Questions(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	); err != nil {
		slog.Error("render questions after creating khayal", "question_id", questionID, "error", err)
	}
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
