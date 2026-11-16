package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
)

var (
	ErrEmptyQuestion     = errors.New("question cannot be empty")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

type QuestionRepository interface {
	CreateQuestion(ctx context.Context, userID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error
	GetQuestion(ctx context.Context, userID, id int) (entities.Question, error)
	GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error)
}

type QuestionService struct {
	repo QuestionRepository
}

func NewQuestionService(repo QuestionRepository) *QuestionService {
	return &QuestionService{repo: repo}
}

func (s *QuestionService) CreateQuestion(ctx context.Context, userID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error {
	if len(ciphertext) == 0 {
		return ErrEmptyQuestion
	}
	if len(ciphertext) < 12+16 || encryptionVersion < 1 {
		return ErrInvalidCiphertext
	}
	if err := s.repo.CreateQuestion(ctx, userID, clientID, ciphertext, encryptionVersion); err != nil {
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
