package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/dnp314/khayal/internal/entities"
)

var (
	ErrEmptyKhayal       = errors.New("khayal cannot be empty")
	ErrInvalidReturnTime = errors.New("invalid return time")
)

type ReturnMode string

const (
	ReturnModeRandom ReturnMode = "random"
	ReturnModeGap    ReturnMode = "gap"
	ReturnModeDate   ReturnMode = "date"
)

const (
	minRandomReturnDays = 60
	maxRandomReturnDays = 365
)

// ReturnSchedule is the return-time choice submitted with a new khayal.
type ReturnSchedule struct {
	Mode ReturnMode
	Gap  string
	Date string
}

// Resolve converts the selected schedule into the next return time for the
// question. Random returns land somewhere between two months and a year out.
func (s ReturnSchedule) Resolve(now time.Time) (time.Time, error) {
	switch s.Mode {
	case "", ReturnModeRandom:
		days := minRandomReturnDays + rand.IntN(maxRandomReturnDays-minRandomReturnDays+1)
		return now.AddDate(0, 0, days), nil
	case ReturnModeGap:
		return resolveGap(s.Gap, now)
	case ReturnModeDate:
		return resolveDate(s.Date, now)
	default:
		return time.Time{}, ErrInvalidReturnTime
	}
}

func resolveGap(gap string, now time.Time) (time.Time, error) {
	switch gap {
	case "1w":
		return now.AddDate(0, 0, 7), nil
	case "1m":
		return now.AddDate(0, 1, 0), nil
	case "2m":
		return now.AddDate(0, 2, 0), nil
	case "6m":
		return now.AddDate(0, 6, 0), nil
	case "1y":
		return now.AddDate(1, 0, 0), nil
	default:
		return time.Time{}, ErrInvalidReturnTime
	}
}

func resolveDate(value string, now time.Time) (time.Time, error) {
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02"} {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err != nil {
			continue
		}
		if !parsed.After(now) {
			return time.Time{}, ErrInvalidReturnTime
		}
		return parsed.UTC(), nil
	}
	return time.Time{}, ErrInvalidReturnTime
}

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

func (s *KhayalService) CreateKhayal(ctx context.Context, questionID int, khayal string, schedule ReturnSchedule) error {
	if strings.TrimSpace(khayal) == "" {
		return ErrEmptyKhayal
	}

	returnAt, err := schedule.Resolve(time.Now().UTC())
	if err != nil {
		return err
	}

	if err := s.KhayalRepo.CreateKhayal(ctx, questionID, khayal, returnAt); err != nil {
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
