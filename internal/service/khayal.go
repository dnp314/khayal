package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dnp314/khayal/internal/entities"
)

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
	if khayal == "" {
		return errors.New("title cannot be empty")
	}
	err := s.KhayalRepo.CreateKhayal(ctx, questionID, khayal, time.Now().Add(1*time.Minute))
	if err != nil {
		return err
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
	fmt.Println(khayals)
	if err != nil {
		return nil, err
	}

	return khayals, nil

}
