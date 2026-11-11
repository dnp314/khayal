package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dnp314/khayal/internal/components"
	"github.com/dnp314/khayal/internal/dtos"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/dnp314/khayal/internal/service"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterPost(ctx context.Context, username, password, email string) (string, error)
	Login(ctx context.Context, username, password string) (string, error)
	Logout(ctx context.Context, token string) error
}

type AuthController struct {
	service AuthService
}

func NewAuthController(service AuthService) *AuthController {
	return &AuthController{
		service: service,
	}

}

func ShowLogin(ctx *gin.Context) {
	err := components.LoginPage("").Render(ctx.Request.Context(), ctx.Writer)
	if err != nil {
		slog.Error("render login page", "error", err)
	}
}

func ShowRegister(ctx *gin.Context) {
	err := components.RegisterPage("").Render(
		ctx.Request.Context(),
		ctx.Writer,
	)
	if err != nil {
		slog.Error("render register page", "error", err)
	}
}

func (ac *AuthController) Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	token, err := ac.service.Login(ctx, username, password)
	if err != nil {
		slog.Error("login failed", "username", username, "error", err)
		renderLoginError(ctx, "Invalid username or password")
		return
	}

	middleware.SetSessionCookie(ctx, token, int(service.SessionDuration.Seconds()))
	redirectAfterAuth(ctx, "/")
}

func (ac *AuthController) PostRegister(ctx *gin.Context) {

	var req dtos.RegisterRequest

	if err := ctx.ShouldBind(&req); err != nil {
		slog.Warn("invalid registration request", "error", err)
		renderRegisterError(ctx, "Invalid input: check username, email and password")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password", "username", req.Username, "error", err)
		renderRegisterError(ctx, "Unable to create account")
		return
	}

	token, err := ac.service.RegisterPost(ctx, req.Username, string(passwordHash), req.Email)
	if err != nil {
		slog.Error("register user", "username", req.Username, "error", err)
		renderRegisterError(ctx, "Unable to create account")
		return
	}

	middleware.SetSessionCookie(ctx, token, int(service.SessionDuration.Seconds()))
	redirectAfterAuth(ctx, "/")
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

func renderLoginError(ctx *gin.Context, message string) {
	if ctx.GetHeader("HX-Request") == "true" {
		if err := components.ErrorBanner(message).Render(ctx.Request.Context(), ctx.Writer); err != nil {
			slog.Error("render login error", "error", err)
		}
		return
	}

	if err := components.LoginPage(message).Render(ctx.Request.Context(), ctx.Writer); err != nil {
		slog.Error("render login page after error", "error", err)
	}
}

func renderRegisterError(ctx *gin.Context, message string) {
	if ctx.GetHeader("HX-Request") == "true" {
		if err := components.ErrorBanner(message).Render(ctx.Request.Context(), ctx.Writer); err != nil {
			slog.Error("render register error", "error", err)
		}
		return
	}

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
