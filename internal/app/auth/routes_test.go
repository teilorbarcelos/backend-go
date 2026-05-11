package auth

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"backend-go/pkg/database"
)

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	publicRG := r.Group("/v1")
	protectedRG := r.Group("/v1")
	
	// We use the real DB from the test environment initialized in TestMain
	RegisterRoutes(publicRG, protectedRG, database.DB)
	
	routes := r.Routes()
	
	expectedRoutes := map[string]string{
		"POST": "/v1/auth/login",
		"GET":  "/v1/auth/me",
	}

	for _, route := range routes {
		for method, path := range expectedRoutes {
			if route.Method == method && route.Path == path {
				delete(expectedRoutes, method)
			}
		}
	}
	
	assert.Empty(t, expectedRoutes, "Some expected routes were not registered")
}
