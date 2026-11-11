package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session"
	UserIDContextKey  = "user_id"
)

type SessionService interface {
	ValidateSession(ctx context.Context, token string) (entities.User, error)
}

func Auth(auth SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		user, err := auth.ValidateSession(c.Request.Context(), token)
		if err != nil {
			slog.Warn("invalid session", "error", err)
			ClearSessionCookie(c)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Set(UserIDContextKey, user.ID)
		c.Next()
	}
}

func SetSessionCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, token, maxAge, "/", "", false, true)
}

func ClearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}
