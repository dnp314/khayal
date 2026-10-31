package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dnp314/khayal/internal/entities"
)

var ErrEmptyKhayal = errors.New("khayal cannot be empty")

type KhayalRepository interface {
	CreateKhayal(ctx context.Context, questionID int, khayal string, returnAt time.Time) error
	// GetKhayal(ctx context.Context, id int) (entities.Khayal, error)
	GetAllKhayals(ctx context.Context, questionID int) ([]entities.Khayal, error)
}

type KhayalService struct {
	KhayalRepo   KhayalRepository
	QuestionRepo QuestionRepository
}

func NewKhayalService(KhayalRepo KhayalRepository, QuestionRepo QuestionRepository) *KhayalService {
	return &KhayalService{
		KhayalRepo:   KhayalRepo,
		QuestionRepo: QuestionRepo,
	}
}

func (s *KhayalService) CreateKhayal(ctx context.Context, questionID int, khayal string) error {
	if strings.TrimSpace(khayal) == "" {
		return ErrEmptyKhayal
	}
	if err := s.KhayalRepo.CreateKhayal(ctx, questionID, khayal, time.Now().Add(1*time.Minute)); err != nil {
		return fmt.Errorf("create khayal: %w", err)
	}
	return nil
}

// func (s *KhayalService) GetKhayal(ctx context.Context, id int) (entities.Khayal, error) {
// 	question, err := s.repo.GetKhayal(ctx, id)
// 	if err != nil {
// 		return entities.Khayal{}, err
// 	}
// 	return question, nil
// }

func (s *KhayalService) GetAllKhayals(ctx context.Context, questionID int) ([]entities.Khayal, error) {

	khayals, err := s.KhayalRepo.GetAllKhayals(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("get all khayals for question %d: %w", questionID, err)
	}

	return khayals, nil

}
