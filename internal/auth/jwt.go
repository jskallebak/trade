package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int32, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is empty")
	}

	ea := time.Now().Add(24 * time.Hour)

	claims := Claims{
		UserID: userID,
		Email:  email,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(ea),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "trading-dashboard",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("Error signing the token")
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string) (userID int32, email string, err error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		return -1, "", fmt.Errorf("JWT_SECRET is empty")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return -1, "", fmt.Errorf("error parsing token: %v", err)
	}

	if !token.Valid {
		return -1, "", fmt.Errorf("invalid token")
	}

	return claims.UserID, claims.Email, nil
}
