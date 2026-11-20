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

// LoadSession attaches the current user to the request when a valid session
// cookie is present. It never blocks the request, so public pages can render
// either the signed-in or signed-out view.
func LoadSession(sessions SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.Next()
			return
		}

		user, err := sessions.ValidateSession(c.Request.Context(), token)
		if err != nil {
			slog.Warn("invalid session", "error", err)
			ClearSessionCookie(c)
			c.Next()
			return
		}

		c.Set(UserIDContextKey, user.ID)
		c.Next()
	}
}

// RequireAuth blocks requests that do not have a user loaded by LoadSession.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get(UserIDContextKey); !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

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
