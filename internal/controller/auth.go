package controller

import (
	"context"
	"log/slog"

	"github.com/dnp314/khayal/internal/components"
	"github.com/dnp314/khayal/internal/dtos"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterPost(ctx context.Context, username, password, email string) error
	Login(ctx context.Context, username, password string) error
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

	err := ac.service.Login(ctx, username, password)
	if err != nil {
		slog.Error("login failed", "username", username, "error", err)
		if renderErr := components.LoginPage("Invalid username or password").
			Render(ctx.Request.Context(), ctx.Writer); renderErr != nil {
			slog.Error("render login page after failed login", "error", renderErr)
		}
		return
	}
}

func (ac *AuthController) PostRegister(ctx *gin.Context) {

	var req dtos.RegisterRequest

	if err := ctx.ShouldBind(&req); err != nil {
		slog.Warn("invalid registration request", "error", err)
		// re-render the form with an error message
		if renderErr := components.RegisterPage("Invalid input: check username, email and password").
			Render(ctx.Request.Context(), ctx.Writer); renderErr != nil {
			slog.Error("render register page after invalid input", "error", renderErr)
		}
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password", "username", req.Username, "error", err)
		if renderErr := components.RegisterPage("Unable to create account").
			Render(ctx.Request.Context(), ctx.Writer); renderErr != nil {
			slog.Error("render register page after password hash failure", "error", renderErr)
		}
		return
	}

	err = ac.service.RegisterPost(ctx, req.Username, string(passwordHash), req.Email)
	if err != nil {
		slog.Error("register user", "username", req.Username, "error", err)
		if renderErr := components.RegisterPage("Unable to create account").
			Render(ctx.Request.Context(), ctx.Writer); renderErr != nil {
			slog.Error("render register page after registration failure", "error", renderErr)
		}
		return
	}
}
