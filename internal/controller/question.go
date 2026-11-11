package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/dnp314/khayal/internal/service"

	"github.com/dnp314/khayal/internal/components"

	"github.com/gin-gonic/gin"
)

type QuestionService interface {
	GetQuestion(ctx context.Context, userID, id int) (entities.Question, error)
	GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error)
	CreateQuestion(ctx context.Context, userID int, question string) error
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
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	questions, err := c.questionService.GetAllQuestions(ctx, userID)
	if err != nil {
		slog.Error("get questions for home page", "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load questions")
		return
	}
	err = components.HomePage(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
	if err != nil {
		slog.Error("render home page", "user_id", userID, "error", err)
	}
}

func (c *QuestionController) CreateQuestion(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	question := ctx.PostForm("question")
	err := c.questionService.CreateQuestion(ctx, userID, question)
	if err != nil {
		if errors.Is(err, service.ErrEmptyQuestion) {
			slog.Warn("reject empty question", "user_id", userID)
			ctx.String(http.StatusBadRequest, "Question can't be empty")
			return
		}

		slog.Error("create question", "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to create question")
		return
	}

	questions, err := c.questionService.GetAllQuestions(ctx, userID)
	if err != nil {
		slog.Error("get questions after creating question", "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load questions")
		return
	}

	err = components.Questions(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
	if err != nil {
		slog.Error("render questions after creation", "user_id", userID, "error", err)
	}

}

func (c *QuestionController) GetQuestion(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		slog.Warn("invalid question id", "id", ctx.Param("id"), "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}
	question, err := c.questionService.GetQuestion(ctx.Request.Context(), userID, id)
	if err != nil {
		slog.Error("get question", "question_id", id, "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load question")
		return
	}
	khayals, err := c.khayalService.GetAllKhayals(ctx, id)
	if err != nil {
		slog.Error("get khayals for question", "question_id", id, "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load khayals")
		return
	}

	err = components.QuestionPage(question, khayals).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
	if err != nil {
		slog.Error("render question page", "question_id", id, "user_id", userID, "error", err)
	}
}

func (c *QuestionController) GetAllQuestions(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	questions, err := c.questionService.GetAllQuestions(ctx, userID)
	if err != nil {
		slog.Error("get all questions", "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load questions")
		return
	}
	err = components.Questions(questions).Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
	if err != nil {
		slog.Error("render questions", "user_id", userID, "error", err)
	}
}

func requireUserID(ctx *gin.Context) (int, bool) {
	value, exists := ctx.Get(middleware.UserIDContextKey)
	if !exists {
		slog.Error("user id missing from request context")
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return 0, false
	}

	userID, ok := value.(int)
	if !ok {
		slog.Error("invalid user id in request context", "value", value)
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return 0, false
	}

	return userID, true
}
