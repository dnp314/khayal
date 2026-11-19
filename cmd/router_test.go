package main

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// Registering routes is where Gin detects conflicting wildcards, so this
// catches route conflicts that compilation alone cannot.
func TestCreateRouterRegistersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createRouter(Controllers{}, nil)
}
