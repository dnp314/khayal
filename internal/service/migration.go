package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
)

var ErrInvalidLegacyImport = errors.New("invalid legacy import")

type LegacyQuestionRepository interface {
	GetLegacyQuestions(ctx context.Context, userID int) ([]entities.LegacyQuestion, error)
	ImportLegacyQuestion(ctx context.Context, userID, id int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error
}

type LegacyKhayalRepository interface {
	GetLegacyKhayals(ctx context.Context, userID int) ([]entities.LegacyKhayal, error)
	ImportLegacyKhayal(ctx context.Context, userID, id, questionID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error
}

type MigrationService struct {
	questionRepo LegacyQuestionRepository
	khayalRepo   LegacyKhayalRepository
}

func NewMigrationService(questionRepo LegacyQuestionRepository, khayalRepo LegacyKhayalRepository) *MigrationService {
	return &MigrationService{
		questionRepo: questionRepo,
		khayalRepo:   khayalRepo,
	}
}

type LegacyRows struct {
	Questions []entities.LegacyQuestion `json:"questions"`
	Khayals   []entities.LegacyKhayal   `json:"khayals"`
}

type EncryptedQuestionInput struct {
	ID                int
	ClientID          uuid.UUID
	Ciphertext        []byte
	EncryptionVersion int
}

type EncryptedKhayalInput struct {
	ID                int
	QuestionID        int
	ClientID          uuid.UUID
	Ciphertext        []byte
	EncryptionVersion int
}

func (s *MigrationService) GetLegacyRows(ctx context.Context, userID int) (LegacyRows, error) {
	questions, err := s.questionRepo.GetLegacyQuestions(ctx, userID)
	if err != nil {
		return LegacyRows{}, fmt.Errorf("get legacy questions: %w", err)
	}
	khayals, err := s.khayalRepo.GetLegacyKhayals(ctx, userID)
	if err != nil {
		return LegacyRows{}, fmt.Errorf("get legacy khayals: %w", err)
	}
	return LegacyRows{Questions: questions, Khayals: khayals}, nil
}

func (s *MigrationService) ImportLegacyRows(
	ctx context.Context,
	userID int,
	questions []EncryptedQuestionInput,
	khayals []EncryptedKhayalInput,
) error {
	for _, q := range questions {
		if len(q.Ciphertext) < 12+16 || q.EncryptionVersion < 1 {
			return ErrInvalidLegacyImport
		}
		if err := s.questionRepo.ImportLegacyQuestion(ctx, userID, q.ID, q.ClientID, q.Ciphertext, q.EncryptionVersion); err != nil {
			return fmt.Errorf("import question %d: %w", q.ID, err)
		}
	}
	for _, k := range khayals {
		if len(k.Ciphertext) < 12+16 || k.EncryptionVersion < 1 {
			return ErrInvalidLegacyImport
		}
		if err := s.khayalRepo.ImportLegacyKhayal(ctx, userID, k.ID, k.QuestionID, k.ClientID, k.Ciphertext, k.EncryptionVersion); err != nil {
			return fmt.Errorf("import khayal %d: %w", k.ID, err)
		}
	}
	return nil
}
