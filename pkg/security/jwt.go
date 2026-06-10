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
	UserID         string       `json:"id"`
	Email          string       `json:"email"`
	RoleID         string       `json:"roleId"`
	SessionVersion int          `json:"sessionVersion"`
	Permissions    []Permission `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

var jwtParseWithClaims = jwt.ParseWithClaims

func GenerateToken(userID, email, roleID string, permissions []Permission, sessionVersion int) (string, error) {
	return generateToken(userID, email, roleID, permissions, sessionVersion, 24*time.Hour)
}

func GenerateRefreshToken(userID, email, roleID string) (string, error) {
	return generateToken(userID, email, roleID, nil, 0, 7*24*time.Hour)
}

func generateToken(userID, email, roleID string, permissions []Permission, sessionVersion int, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:         userID,
		Email:          email,
		RoleID:         roleID,
		SessionVersion: sessionVersion,
		Permissions:    permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
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
