package service

import (
	"context"
	"log"
	"log/slog"

	"github.com/dnp314/khayal/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	PostRegister(ctx context.Context, username string, password string) error
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

func (s *AuthService) RegisterPost(ctx context.Context, username, password string) error {
	err := s.repo.PostRegister(ctx, username, password)
	if err != nil {
		log.Println(err)
	}

	slog.Info("user created", "username", username)
	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) error {
	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		slog.Error("get user", "user", user, "err", err)
	}
	if user.ID == 0 {
		log.Println("register first")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Println(err)
	}
	log.Println("logged in")

	return nil

}
