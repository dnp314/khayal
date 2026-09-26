package service

import (
	"context"
	"errors"

	"github.com/dnp314/khayal/internal/entities"
)

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
	if question == "" {
		return errors.New("title cannot be empty")
	}
	err := s.repo.CreateQuestion(ctx, question)
	if err != nil {
		return err
	}
	return nil
}

func (s *QuestionService) GetQuestion(ctx context.Context, id int) (entities.Question, error) {
	question, err := s.repo.GetQuestion(ctx, id)
	if err != nil {
		return entities.Question{}, err
	}
	return question, nil
}

func (s *QuestionService) GetAllQuestions(ctx context.Context) ([]entities.Question, error) {
	questions, err := s.repo.GetAllQuestions(ctx)
	if err != nil {
		return nil, err
	}
	return questions, nil
}
