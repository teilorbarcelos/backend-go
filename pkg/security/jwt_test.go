package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"backend-go/pkg/config"
)

func TestJWT(t *testing.T) {
	// Setup
	config.AppConfig.JWTSecret = "test-secret"
	userID := "user-123"
	email := "test@test.com"
	roleID := "admin"
	permissions := []Permission{
		{Feature: "user", View: true, Create: true},
	}

	t.Run("Generate and Validate Token Success", func(t *testing.T) {
		token, err := GenerateToken(userID, email, roleID, permissions)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roleID, claims.RoleID)
		assert.Equal(t, permissions, claims.Permissions)
	})

	t.Run("Validate Token - Invalid Token String", func(t *testing.T) {
		claims, err := ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Validate Token - Wrong Secret", func(t *testing.T) {
		token, _ := GenerateToken(userID, email, roleID, permissions)
		
		// Change secret temporarily
		originalSecret := config.AppConfig.JWTSecret
		config.AppConfig.JWTSecret = "wrong-secret"
		defer func() { config.AppConfig.JWTSecret = originalSecret }()

		claims, err := ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Validate Token - Expired", func(t *testing.T) {
		// Manually create an expired token
		claims := &JWTClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(config.AppConfig.JWTSecret))

		resultClaims, err := ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, resultClaims)
	})

	t.Run("Validate Token - Invalid but no error", func(t *testing.T) {
		origParse := jwtParseWithClaims
		defer func() { jwtParseWithClaims = origParse }()

		jwtParseWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
			return &jwt.Token{Valid: false, Claims: &JWTClaims{}}, nil
		}

		claims, err := ValidateToken("any-token")
		assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
		assert.Nil(t, claims)
	})
}

func TestPassword(t *testing.T) {
	password := "secret123"
	
	t.Run("Hash and Check Success", func(t *testing.T) {
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		
		assert.True(t, CheckPasswordHash(password, hash))
	})
	
	t.Run("Check Failure", func(t *testing.T) {
		hash, _ := HashPassword(password)
		assert.False(t, CheckPasswordHash("wrong-password", hash))
	})
}
