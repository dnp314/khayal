package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionDuration      = 7 * 24 * time.Hour
	DefaultKDFAlgorithm  = "PBKDF2-SHA256"
	DefaultKDFIterations = 600000
	DefaultKDFVersion    = 1
)

var (
	ErrSessionExpired     = errors.New("session expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLegacyLoginNeeded  = errors.New("legacy login required")
	ErrCryptoNotSetup     = errors.New("encryption is not set up")
	ErrRecoveryFailed     = errors.New("recovery verification failed")
)

type AuthParams struct {
	Legacy        bool   `json:"legacy"`
	KDFAlgorithm  string `json:"kdf_algorithm"`
	KDFIterations int    `json:"kdf_iterations"`
	KDFVersion    int    `json:"kdf_version"`
	Salt          string `json:"salt"`
}

type RecoveryParams struct {
	WrappedDataKey string `json:"wrapped_data_key"`
	WrappedKeyIV   string `json:"wrapped_data_key_iv"`
}

type AuthRepository interface {
	CreateUser(ctx context.Context, user entities.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
	GetUserByID(ctx context.Context, id int) (entities.User, error)
	UpdateAuthKeys(ctx context.Context, userID int, authKeyHash, kdfAlgorithm string, kdfIterations int, salt, wrappedDataKey, wrappedDataKeyIV []byte) error
	SetupLegacyCrypto(ctx context.Context, userID int, authKeyHash, kdfAlgorithm string, kdfIterations int, salt, wrappedDataKey, wrappedDataKeyIV, recoveryWrappedDataKey, recoveryWrappedDataKeyIV, recoveryVerifier []byte) error
	DeleteSessionsForUser(ctx context.Context, userID int) error
	CreateSession(ctx context.Context, session entities.Session) error
	GetSession(ctx context.Context, token string) (entities.Session, error)
	DeleteSession(ctx context.Context, token string) error
}

type AuthService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) RegisterEncrypted(ctx context.Context, user entities.User) (string, entities.User, error) {
	now := time.Now()
	user.CryptoSetupAt = &now

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return "", entities.User{}, fmt.Errorf("register user %q: %w", user.Name, err)
	}
	user.ID = id
	user.PasswordHash = ""

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", entities.User{}, err
	}

	slog.Info("user registered", "user_id", user.ID)
	return token, user, nil
}

func (s *AuthService) AuthParams(ctx context.Context, email string) AuthParams {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return fakeAuthParams(email)
	}
	if !user.CryptoReady() {
		return AuthParams{Legacy: true}
	}
	return AuthParams{
		KDFAlgorithm:  user.KDFAlgorithm,
		KDFIterations: user.KDFIterations,
		KDFVersion:    user.KDFVersion,
		Salt:          encodeBase64(user.KDFSalt),
	}
}

func (s *AuthService) Login(ctx context.Context, email, authKey string) (string, entities.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", entities.User{}, ErrInvalidCredentials
	}
	if !user.CryptoReady() {
		return "", entities.User{}, ErrLegacyLoginNeeded
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.AuthKeyHash), []byte(authKey)); err != nil {
		return "", entities.User{}, ErrInvalidCredentials
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", entities.User{}, err
	}
	slog.Info("user logged in", "user_id", user.ID)
	return token, user, nil
}

func (s *AuthService) LegacyLogin(ctx context.Context, email, password string) (string, entities.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user.PasswordHash == "" {
		return "", entities.User{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", entities.User{}, ErrInvalidCredentials
	}

	token, err := s.createSession(ctx, user.ID)
	if err != nil {
		return "", entities.User{}, err
	}
	slog.Info("legacy user logged in", "user_id", user.ID)
	return token, user, nil
}

func (s *AuthService) SetupLegacyCrypto(ctx context.Context, userID int, user entities.User) error {
	current, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user for legacy setup: %w", err)
	}
	if current.CryptoReady() {
		return ErrCryptoNotSetup
	}
	if err := s.repo.SetupLegacyCrypto(
		ctx, userID, user.AuthKeyHash, user.KDFAlgorithm, user.KDFIterations,
		user.KDFSalt, user.WrappedDataKey, user.WrappedDataKeyIV,
		user.RecoveryWrappedDataKey, user.RecoveryDataKeyIV, user.RecoveryVerifier,
	); err != nil {
		return fmt.Errorf("setup legacy crypto for user %d: %w", userID, err)
	}
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int, currentAuthKey string, user entities.User) (string, error) {
	current, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user for password change: %w", err)
	}
	if !current.CryptoReady() {
		return "", ErrCryptoNotSetup
	}
	if err := bcrypt.CompareHashAndPassword([]byte(current.AuthKeyHash), []byte(currentAuthKey)); err != nil {
		return "", ErrInvalidCredentials
	}

	if err := s.repo.UpdateAuthKeys(
		ctx, userID, user.AuthKeyHash, user.KDFAlgorithm, user.KDFIterations,
		user.KDFSalt, user.WrappedDataKey, user.WrappedDataKeyIV,
	); err != nil {
		return "", err
	}
	if err := s.repo.DeleteSessionsForUser(ctx, userID); err != nil {
		return "", err
	}
	return s.createSession(ctx, userID)
}

func (s *AuthService) RecoveryParams(ctx context.Context, email string) RecoveryParams {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || !user.CryptoReady() {
		return fakeRecoveryParams(email)
	}
	return RecoveryParams{
		WrappedDataKey: encodeBase64(user.RecoveryWrappedDataKey),
		WrappedKeyIV:   encodeBase64(user.RecoveryDataKeyIV),
	}
}

func (s *AuthService) Recover(ctx context.Context, email, verifier string, user entities.User) (string, entities.User, error) {
	current, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || !current.CryptoReady() {
		return "", entities.User{}, ErrRecoveryFailed
	}

	provided, err := decodeBase64(verifier)
	if err != nil || len(provided) != len(current.RecoveryVerifier) ||
		subtle.ConstantTimeCompare(provided, current.RecoveryVerifier) != 1 {
		return "", entities.User{}, ErrRecoveryFailed
	}

	if err := s.repo.UpdateAuthKeys(
		ctx, current.ID, user.AuthKeyHash, user.KDFAlgorithm, user.KDFIterations,
		user.KDFSalt, user.WrappedDataKey, user.WrappedDataKeyIV,
	); err != nil {
		return "", entities.User{}, err
	}
	if err := s.repo.DeleteSessionsForUser(ctx, current.ID); err != nil {
		return "", entities.User{}, err
	}

	token, err := s.createSession(ctx, current.ID)
	if err != nil {
		return "", entities.User{}, err
	}
	current.AuthKeyHash = user.AuthKeyHash
	current.KDFAlgorithm = user.KDFAlgorithm
	current.KDFIterations = user.KDFIterations
	current.KDFSalt = user.KDFSalt
	current.WrappedDataKey = user.WrappedDataKey
	current.WrappedDataKeyIV = user.WrappedDataKeyIV

	slog.Info("user recovered account", "user_id", current.ID)
	return token, current, nil
}

func (s *AuthService) GetUser(ctx context.Context, id int) (entities.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return entities.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
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

func fakeAuthParams(email string) AuthParams {
	sum := sha256.Sum256([]byte("khayal-fake-salt:" + strings.ToLower(strings.TrimSpace(email))))
	return AuthParams{
		KDFAlgorithm:  DefaultKDFAlgorithm,
		KDFIterations: DefaultKDFIterations,
		KDFVersion:    DefaultKDFVersion,
		Salt:          encodeBase64(sum[:16]),
	}
}

func fakeRecoveryParams(email string) RecoveryParams {
	normalized := strings.ToLower(strings.TrimSpace(email))
	first := sha256.Sum256([]byte("khayal-fake-recovery:" + normalized))
	second := sha256.Sum256([]byte("khayal-fake-recovery-extra:" + normalized))
	ivHash := sha256.Sum256([]byte("khayal-fake-recovery-iv:" + normalized))
	wrapped := make([]byte, 0, 48)
	wrapped = append(wrapped, first[:]...)
	wrapped = append(wrapped, second[:16]...)
	return RecoveryParams{
		WrappedDataKey: encodeBase64(wrapped),
		WrappedKeyIV:   encodeBase64(ivHash[:12]),
	}
}

func encodeBase64(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64Encode(b)
}

func decodeBase64(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	return base64Decode(s)
}
