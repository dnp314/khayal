package controller

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dnp314/khayal/internal/components"
	"github.com/dnp314/khayal/internal/dtos"
	"github.com/dnp314/khayal/internal/entities"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/dnp314/khayal/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterEncrypted(ctx context.Context, user entities.User) (string, entities.User, error)
	AuthParams(ctx context.Context, email string) service.AuthParams
	Login(ctx context.Context, email, authKey string) (string, entities.User, error)
	LegacyLogin(ctx context.Context, email, password string) (string, entities.User, error)
	SetupLegacyCrypto(ctx context.Context, userID int, user entities.User) error
	ChangePassword(ctx context.Context, userID int, currentAuthKey string, user entities.User) (string, error)
	RecoveryParams(ctx context.Context, email string) service.RecoveryParams
	Recover(ctx context.Context, email, verifier string, user entities.User) (string, entities.User, error)
	GetUser(ctx context.Context, id int) (entities.User, error)
	Logout(ctx context.Context, token string) error
}

type LegacyMigrationService interface {
	GetLegacyRows(ctx context.Context, userID int) (service.LegacyRows, error)
	ImportLegacyRows(ctx context.Context, userID int, questions []service.EncryptedQuestionInput, khayals []service.EncryptedKhayalInput) error
}

type AuthController struct {
	service   AuthService
	migration LegacyMigrationService
}

func NewAuthController(service AuthService, migration LegacyMigrationService) *AuthController {
	return &AuthController{
		service:   service,
		migration: migration,
	}
}

func ShowLogin(ctx *gin.Context) {
	if err := components.LoginPage("").Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render login page", "error", err)
	}
}

func ShowRegister(ctx *gin.Context) {
	if err := components.RegisterPage("").Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render register page", "error", err)
	}
}

func ShowRecover(ctx *gin.Context) {
	if err := components.RecoverPage("").Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render recover page", "error", err)
	}
}

func (ac *AuthController) Account(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	user, err := ac.service.GetUser(ctx, userID)
	if err != nil {
		slog.Error("get account user", "user_id", userID, "error", err)
		ctx.String(http.StatusInternalServerError, "Unable to load account")
		return
	}
	if err := components.AccountPage(user).Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render account page", "user_id", userID, "error", err)
	}
}

func (ac *AuthController) UserKeys(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	user, err := ac.service.GetUser(ctx, userID)
	if err != nil {
		slog.Error("get user keys", "user_id", userID, "error", err)
		jsonError(ctx, http.StatusInternalServerError, "Unable to load keys")
		return
	}
	if !user.CryptoReady() {
		jsonError(ctx, http.StatusConflict, "Encryption is not set up")
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, gin.H{
		"wrapped_data_key":    b64Encode(user.WrappedDataKey),
		"wrapped_data_key_iv": b64Encode(user.WrappedDataKeyIV),
		"kdf_algorithm":       user.KDFAlgorithm,
		"kdf_iterations":      user.KDFIterations,
		"kdf_version":         user.KDFVersion,
		"salt":                b64Encode(user.KDFSalt),
	})
}

func (ac *AuthController) AuthParams(ctx *gin.Context) {
	email := strings.TrimSpace(ctx.Query("email"))
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, ac.service.AuthParams(ctx, email))
}

func (ac *AuthController) RecoveryParams(ctx *gin.Context) {
	email := strings.TrimSpace(ctx.Query("email"))
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, ac.service.RecoveryParams(ctx, email))
}

func (ac *AuthController) Login(ctx *gin.Context) {
	var req dtos.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid login request")
		return
	}

	token, user, err := ac.service.Login(ctx, strings.TrimSpace(req.Email), req.AuthKey)
	if err != nil {
		if errors.Is(err, service.ErrLegacyLoginNeeded) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":                 "legacy_login_required",
				"legacy_login_required": true,
			})
			return
		}
		slog.Warn("login failed", "email", strings.TrimSpace(req.Email))
		jsonError(ctx, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	ac.setSession(ctx, token)
	ctx.JSON(http.StatusOK, gin.H{
		"user_id":             user.ID,
		"wrapped_data_key":    b64Encode(user.WrappedDataKey),
		"wrapped_data_key_iv": b64Encode(user.WrappedDataKeyIV),
		"redirect":            "/",
	})
}

func (ac *AuthController) LegacyLogin(ctx *gin.Context) {
	var req dtos.LegacyLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid login request")
		return
	}

	token, user, err := ac.service.LegacyLogin(ctx, strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		slog.Warn("legacy login failed", "email", strings.TrimSpace(req.Email))
		jsonError(ctx, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	ac.setSession(ctx, token)
	ctx.JSON(http.StatusOK, gin.H{
		"user_id":  user.ID,
		"legacy":   true,
		"redirect": "/",
	})
}

func (ac *AuthController) Register(ctx *gin.Context) {
	var req dtos.EncryptedRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid registration request")
		return
	}
	if len(req.AuthKey) < 32 {
		jsonError(ctx, http.StatusBadRequest, "Invalid authentication key")
		return
	}

	salt, ok := decodeField(req.Salt)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid salt")
		return
	}
	wrappedKey, ok := decodeField(req.WrappedDataKey)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid wrapped key")
		return
	}
	wrappedIV, ok := decodeField(req.WrappedKeyIV)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid wrapped key IV")
		return
	}
	recoveryKey, ok := decodeField(req.RecoveryWrappedDataKey)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid recovery key")
		return
	}
	recoveryIV, ok := decodeField(req.RecoveryWrappedKeyIV)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid recovery key IV")
		return
	}
	verifier, ok := decodeField(req.RecoveryVerifier)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid recovery verifier")
		return
	}

	authKeyHash, err := bcrypt.GenerateFromPassword([]byte(req.AuthKey), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash auth key", "error", err)
		jsonError(ctx, http.StatusInternalServerError, "Unable to create account")
		return
	}

	user := entities.User{
		Name:                   strings.TrimSpace(req.Username),
		Email:                  strings.TrimSpace(req.Email),
		AuthKeyHash:            string(authKeyHash),
		KDFAlgorithm:           req.KDFAlgorithm,
		KDFIterations:          req.KDFIterations,
		KDFSalt:                salt,
		KDFVersion:             req.KDFVersion,
		WrappedDataKey:         wrappedKey,
		WrappedDataKeyIV:       wrappedIV,
		RecoveryWrappedDataKey: recoveryKey,
		RecoveryDataKeyIV:      recoveryIV,
		RecoveryVerifier:       verifier,
	}

	token, created, err := ac.service.RegisterEncrypted(ctx, user)
	if err != nil {
		slog.Warn("register user failed", "username", req.Username, "error", err)
		jsonError(ctx, http.StatusConflict, "Unable to create account")
		return
	}

	ac.setSession(ctx, token)
	ctx.JSON(http.StatusCreated, gin.H{
		"user_id":  created.ID,
		"redirect": "/",
	})
}

func (ac *AuthController) SetupLegacy(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	var req dtos.LegacySetupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid setup request")
		return
	}

	user, ok := buildCryptoUser(req.AuthKey, req.KDFAlgorithm, req.KDFIterations, req.KDFVersion,
		req.Salt, req.WrappedDataKey, req.WrappedKeyIV, req.RecoveryWrappedDataKey,
		req.RecoveryWrappedKeyIV, req.RecoveryVerifier)
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid encryption key material")
		return
	}

	if err := ac.service.SetupLegacyCrypto(ctx, userID, user); err != nil {
		slog.Error("setup legacy crypto", "user_id", userID, "error", err)
		jsonError(ctx, http.StatusInternalServerError, "Unable to set up encryption")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (ac *AuthController) ChangePassword(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	var req dtos.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid password change request")
		return
	}

	user, ok := buildCryptoUser(req.AuthKey, req.KDFAlgorithm, req.KDFIterations, req.KDFVersion,
		req.Salt, req.WrappedDataKey, req.WrappedKeyIV, "", "", "")
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid encryption key material")
		return
	}

	token, err := ac.service.ChangePassword(ctx, userID, req.CurrentAuthKey, user)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			jsonError(ctx, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		slog.Error("change password", "user_id", userID, "error", err)
		jsonError(ctx, http.StatusInternalServerError, "Unable to change password")
		return
	}

	ac.setSession(ctx, token)
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (ac *AuthController) Recover(ctx *gin.Context) {
	var req dtos.RecoveryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid recovery request")
		return
	}

	user, ok := buildCryptoUser(req.AuthKey, req.KDFAlgorithm, req.KDFIterations, req.KDFVersion,
		req.Salt, req.WrappedDataKey, req.WrappedKeyIV, "", "", "")
	if !ok {
		jsonError(ctx, http.StatusBadRequest, "Invalid encryption key material")
		return
	}

	token, recovered, err := ac.service.Recover(ctx, strings.TrimSpace(req.Email), req.RecoveryVerifier, user)
	if err != nil {
		slog.Warn("recover account failed", "email", strings.TrimSpace(req.Email))
		jsonError(ctx, http.StatusUnauthorized, "Invalid recovery code")
		return
	}

	ac.setSession(ctx, token)
	ctx.JSON(http.StatusOK, gin.H{
		"user_id":  recovered.ID,
		"redirect": "/",
	})
}

func (ac *AuthController) LegacyRows(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}
	rows, err := ac.migration.GetLegacyRows(ctx, userID)
	if err != nil {
		slog.Error("get legacy rows", "user_id", userID, "error", err)
		jsonError(ctx, http.StatusInternalServerError, "Unable to load legacy data")
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, rows)
}

func (ac *AuthController) ImportLegacyRows(ctx *gin.Context) {
	userID, ok := requireUserID(ctx)
	if !ok {
		return
	}

	var req dtos.LegacyImportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		jsonError(ctx, http.StatusBadRequest, "Invalid migration request")
		return
	}

	questions := make([]service.EncryptedQuestionInput, 0, len(req.Questions))
	for _, q := range req.Questions {
		clientID, err := uuid.Parse(q.ClientID)
		if err != nil {
			jsonError(ctx, http.StatusBadRequest, "Invalid question client id")
			return
		}
		ciphertext, ok := decodeField(q.Ciphertext)
		if !ok {
			jsonError(ctx, http.StatusBadRequest, "Invalid question ciphertext")
			return
		}
		questions = append(questions, service.EncryptedQuestionInput{
			ID: q.ID, ClientID: clientID, Ciphertext: ciphertext, EncryptionVersion: q.EncryptionVersion,
		})
	}

	khayals := make([]service.EncryptedKhayalInput, 0, len(req.Khayals))
	for _, k := range req.Khayals {
		clientID, err := uuid.Parse(k.ClientID)
		if err != nil {
			jsonError(ctx, http.StatusBadRequest, "Invalid khayal client id")
			return
		}
		ciphertext, ok := decodeField(k.Ciphertext)
		if !ok {
			jsonError(ctx, http.StatusBadRequest, "Invalid khayal ciphertext")
			return
		}
		khayals = append(khayals, service.EncryptedKhayalInput{
			ID: k.ID, QuestionID: k.QuestionID, ClientID: clientID,
			Ciphertext: ciphertext, EncryptionVersion: k.EncryptionVersion,
		})
	}

	if err := ac.migration.ImportLegacyRows(ctx, userID, questions, khayals); err != nil {
		slog.Error("import legacy rows", "user_id", userID, "error", err)
		jsonError(ctx, http.StatusBadRequest, "Unable to migrate data")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (ac *AuthController) Logout(ctx *gin.Context) {
	if token, err := ctx.Cookie(middleware.SessionCookieName); err == nil {
		if err := ac.service.Logout(ctx, token); err != nil {
			slog.Error("logout", "error", err)
		}
	}

	middleware.ClearSessionCookie(ctx)
	redirectAfterAuth(ctx, "/login")
}

func (ac *AuthController) setSession(ctx *gin.Context, token string) {
	middleware.SetSessionCookie(ctx, token, int(service.SessionDuration.Seconds()))
}

func buildCryptoUser(
	authKey, kdfAlgorithm string,
	kdfIterations, kdfVersion int,
	salt, wrappedDataKey, wrappedDataKeyIV,
	recoveryWrappedDataKey, recoveryWrappedDataKeyIV, recoveryVerifier string,
) (entities.User, bool) {
	if len(authKey) < 32 {
		return entities.User{}, false
	}
	authKeyHash, err := bcrypt.GenerateFromPassword([]byte(authKey), bcrypt.DefaultCost)
	if err != nil {
		return entities.User{}, false
	}
	saltBytes, ok := decodeField(salt)
	if !ok {
		return entities.User{}, false
	}
	wrapped, ok := decodeField(wrappedDataKey)
	if !ok {
		return entities.User{}, false
	}
	wrappedIV, ok := decodeField(wrappedDataKeyIV)
	if !ok {
		return entities.User{}, false
	}

	user := entities.User{
		AuthKeyHash:      string(authKeyHash),
		KDFAlgorithm:     kdfAlgorithm,
		KDFIterations:    kdfIterations,
		KDFSalt:          saltBytes,
		KDFVersion:       kdfVersion,
		WrappedDataKey:   wrapped,
		WrappedDataKeyIV: wrappedIV,
	}
	if recoveryWrappedDataKey != "" {
		recoveryKey, ok := decodeField(recoveryWrappedDataKey)
		if !ok {
			return entities.User{}, false
		}
		recoveryIV, ok := decodeField(recoveryWrappedDataKeyIV)
		if !ok {
			return entities.User{}, false
		}
		verifier, ok := decodeField(recoveryVerifier)
		if !ok {
			return entities.User{}, false
		}
		user.RecoveryWrappedDataKey = recoveryKey
		user.RecoveryDataKeyIV = recoveryIV
		user.RecoveryVerifier = verifier
	}
	return user, true
}

func decodeField(value string) ([]byte, bool) {
	if value == "" {
		return nil, false
	}
	if b, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return b, true
	}
	b, err := base64.StdEncoding.DecodeString(value)
	return b, err == nil
}

func b64Encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func jsonError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"error": message})
}

func renderLoginError(ctx *gin.Context, message string) {
	if err := components.LoginPage(message).Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render login page after error", "error", err)
	}
}

func renderRegisterError(ctx *gin.Context, message string) {
	if err := components.RegisterPage(message).Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render register page after error", "error", err)
	}
}

func redirectAfterAuth(ctx *gin.Context, path string) {
	if ctx.GetHeader("HX-Request") == "true" {
		ctx.Header("HX-Redirect", path)
		ctx.Status(http.StatusNoContent)
		return
	}
	ctx.Redirect(http.StatusSeeOther, path)
}
