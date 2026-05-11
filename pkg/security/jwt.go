package security

import (
	"time"

	"backend-go/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type Permission struct {
	Feature  string `json:"feature"`
	View     bool   `json:"view"`
	Create   bool   `json:"create"`
	Delete   bool   `json:"delete"`
	Activate bool   `json:"activate"`
}

type JWTClaims struct {
	UserID      string       `json:"id"`
	Email       string       `json:"email"`
	RoleID      string       `json:"roleId"`
	Permissions []Permission `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

var jwtParseWithClaims = jwt.ParseWithClaims

func GenerateToken(userID, email, roleID string, permissions []Permission) (string, error) {
	claims := &JWTClaims{
		UserID:      userID,
		Email:       email,
		RoleID:      roleID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwtParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
