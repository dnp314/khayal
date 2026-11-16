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
	"github.com/google/uuid"
)

type KhayalService interface {
	GetAllKhayals(ctx context.Context, userID, questionID int) ([]entities.Khayal, error)
	CreateKhayal(ctx context.Context, userID, questionID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error
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
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	questionID, err := strconv.Atoi(ctx.Param("questionID"))
	if err != nil {
		slog.Warn("invalid question id", "id", ctx.Param("questionID"), "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	clientID, err := uuid.Parse(ctx.PostForm("client_id"))
	if err != nil {
		slog.Warn("invalid khayal client id", "question_id", questionID, "user_id", userID)
		ctx.String(http.StatusBadRequest, "Invalid khayal")
		return
	}
	version, err := strconv.Atoi(ctx.PostForm("encryption_version"))
	if err != nil {
		slog.Warn("invalid khayal encryption version", "question_id", questionID, "user_id", userID)
		ctx.String(http.StatusBadRequest, "Invalid khayal")
		return
	}
	ciphertext, ok := decodeField(ctx.PostForm("khayal_ciphertext"))
	if !ok {
		slog.Warn("invalid khayal ciphertext", "question_id", questionID, "user_id", userID)
		ctx.String(http.StatusBadRequest, "Invalid khayal")
		return
	}

	err = c.khayalService.CreateKhayal(ctx, userID, questionID, clientID, ciphertext, version)
	if err != nil {
		if errors.Is(err, service.ErrEmptyKhayal) || errors.Is(err, service.ErrInvalidCiphertext) {
			slog.Warn("reject invalid khayal", "question_id", questionID, "user_id", userID)
			ctx.String(http.StatusBadRequest, "Khayal can't be empty")
			return
		}
		slog.Error("create khayal", "question_id", questionID, "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to create khayal")
		return
	}

	khayals, err := c.khayalService.GetAllKhayals(ctx, userID, questionID)
	if err != nil {
		slog.Error("get khayals after creation", "question_id", questionID, "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load khayals")
		return
	}
	if err := components.Khayals(khayals).Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render khayals after creation", "question_id", questionID, "error", err)
		return
	}

	questions, err := c.questionService.GetAllQuestions(ctx, userID)
	if err != nil {
		slog.Error("get questions after creating khayal", "question_id", questionID, "user_id", userID, "error", err)
		return
	}
	if err := components.Questions(questions).Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render questions after creating khayal", "question_id", questionID, "error", err)
	}
}
