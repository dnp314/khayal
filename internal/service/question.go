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
	CreateQuestion(ctx context.Context, question string) error
	GetQuestion(ctx context.Context, id int) (entities.Question, error)
	GetAllQuestions(ctx context.Context) ([]entities.Question, error)
}

type QuestionService struct {
	repo QuestionRepository
}

func NewQuestionService(repo QuestionRepository) *QuestionService {
	return &QuestionService{
		repo: repo,
	}
}

func (s *QuestionService) CreateQuestion(ctx context.Context, question string) error {
	if strings.TrimSpace(question) == "" {
		return ErrEmptyQuestion
	}
	if err := s.repo.CreateQuestion(ctx, question); err != nil {
		return fmt.Errorf("create question: %w", err)
	}
	return nil
}

func (s *QuestionService) GetQuestion(ctx context.Context, id int) (entities.Question, error) {
	question, err := s.repo.GetQuestion(ctx, id)
	if err != nil {
		return entities.Question{}, fmt.Errorf("get question: %w", err)
	}
	return question, nil
}

func (s *QuestionService) GetAllQuestions(ctx context.Context) ([]entities.Question, error) {
	questions, err := s.repo.GetAllQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all questions: %w", err)
	}
	return questions, nil
}
