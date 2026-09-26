package service

import "context"

type AuthRepository interface {
	PostRegister(ctx context.Context, username string, password string) error
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
	return err
}
