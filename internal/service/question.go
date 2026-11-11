package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dnp314/khayal/internal/entities"
)

var ErrEmptyQuestion = errors.New("question cannot be empty")

type QuestionRepository interface {
	CreateQuestion(ctx context.Context, userID int, question string) error
	GetQuestion(ctx context.Context, userID, id int) (entities.Question, error)
	GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error)
}

type QuestionService struct {
	repo QuestionRepository
}

func NewQuestionService(repo QuestionRepository) *QuestionService {
	return &QuestionService{
		repo: repo,
	}
}

func (s *QuestionService) CreateQuestion(ctx context.Context, userID int, question string) error {
	if strings.TrimSpace(question) == "" {
		return ErrEmptyQuestion
	}
	if err := s.repo.CreateQuestion(ctx, userID, question); err != nil {
		return fmt.Errorf("create question: %w", err)
	}
	return nil
}

func (s *QuestionService) GetQuestion(ctx context.Context, userID, id int) (entities.Question, error) {
	question, err := s.repo.GetQuestion(ctx, userID, id)
	if err != nil {
		return entities.Question{}, fmt.Errorf("get question: %w", err)
	}
	return question, nil
}

func (s *QuestionService) GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error) {
	questions, err := s.repo.GetAllQuestions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get all questions: %w", err)
	}
	return questions, nil
}
