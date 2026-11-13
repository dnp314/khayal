package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const SessionDuration = 7 * 24 * time.Hour

var ErrSessionExpired = errors.New("session expired")

type AuthRepository interface {
	PostRegister(ctx context.Context, username, password, email string) error
	GetUserByName(ctx context.Context, username string) (entities.User, error)
	GetUserByID(ctx context.Context, id int) (entities.User, error)
	CreateSession(ctx context.Context, session entities.Session) error
	GetSession(ctx context.Context, token string) (entities.Session, error)
	DeleteSession(ctx context.Context, token string) error
}

type AuthService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) RegisterPost(ctx context.Context, username, password, email string) (string, error) {
	if err := s.repo.PostRegister(ctx, username, password, email); err != nil {
		return "", fmt.Errorf("register user %q: %w", username, err)
	}

	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get registered user %q: %w", username, err)
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", err
	}

	slog.Info("user registered", "username", username, "user_id", user.ID)
	return token, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		return "", fmt.Errorf("login user %q: %w", username, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("verify password for user %q: %w", username, err)
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", err
	}

	slog.Info("user logged in", "username", username, "user_id", user.ID)
	return token, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (entities.User, error) {
	session, err := s.repo.GetSession(ctx, token)
	if err != nil {
		return entities.User{}, fmt.Errorf("validate session: %w", err)
	}

	if !time.Now().Before(session.ExpiresAt) {
		if err := s.repo.DeleteSession(ctx, token); err != nil {
			slog.Error("delete expired session", "user_id", session.UserID, "error", err)
		}
		return entities.User{}, ErrSessionExpired
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return entities.User{}, fmt.Errorf("get user for session: %w", err)
	}

	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if err := s.repo.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("log out: %w", err)
	}
	return nil
}

func (s *AuthService) createSession(ctx context.Context, userID int) (string, error) {
	session := entities.Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(SessionDuration),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("create session for user %d: %w", userID, err)
	}
	return session.ID, nil
}
