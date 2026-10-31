package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dnp314/khayal/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	PostRegister(ctx context.Context, username, password, email string) error
	GetUserByName(ctx context.Context, username string) (entities.User, error)
}

type AuthService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) RegisterPost(ctx context.Context, username, password, email string) error {
	if err := s.repo.PostRegister(ctx, username, password, email); err != nil {
		return fmt.Errorf("register user %q: %w", username, err)
	}

	slog.Info("user registered", "username", username)
	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) error {
	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		return fmt.Errorf("login user %q: %w", username, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return fmt.Errorf("verify password for user %q: %w", username, err)
	}

	slog.Info("user logged in", "username", username)
	return nil
}
