package controller

import (
	"context"
	"log"

	"github.com/dnp314/khayal/internal/components"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterPost(ctx context.Context, username, password string) error
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
		log.Println(err)
	}
}

func ShowRegister(ctx *gin.Context) {
	err := components.RegisterPage("").Render(
		ctx,
		ctx.Writer,
	)
	if err != nil {
		log.Println(err)
	}
}
func (ac *AuthController) Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	err := ac.service.Login(ctx, username, password)

	if err != nil {
		log.Println(err)
	}
}

func (ac *AuthController) PostRegister(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(username, err)
	}

	err = ac.service.RegisterPost(ctx, username, string(passwordHash))

	if err != nil {
		log.Println(err)
	}
}
