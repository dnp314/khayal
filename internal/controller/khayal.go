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
	CreateKhayal(ctx context.Context, questionID int, khayal string, schedule service.ReturnSchedule) error
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

	schedule := service.ReturnSchedule{
		Mode: service.ReturnMode(ctx.PostForm("return_mode")),
		Gap:  ctx.PostForm("return_gap"),
		Date: ctx.PostForm("return_date"),
	}

	err = c.khayalService.CreateKhayal(ctx, questionID, khayal, schedule)
	if err != nil {
		if errors.Is(err, service.ErrEmptyKhayal) {
			slog.Warn("reject empty khayal", "question_id", questionID)
			ctx.String(http.StatusBadRequest, "Khayal can't be empty")
			return
		}
		if errors.Is(err, service.ErrInvalidReturnTime) {
			slog.Warn("reject invalid return time", "question_id", questionID, "mode", schedule.Mode, "gap", schedule.Gap, "date", schedule.Date)
			ctx.String(http.StatusBadRequest, "Choose a valid return time in the future")
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

	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	questions, err := c.questionService.GetAllQuestions(ctx, userID)
	if err != nil {
		slog.Error("get questions after creating khayal", "question_id", questionID, "user_id", userID, "error", err)
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
