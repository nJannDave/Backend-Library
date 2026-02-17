package token

import (
	"stmnplibrary/security/jwt/claims"

	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateToken(userId int, ttl time.Duration, role string) (string, error) {
	data := &claims.JWTClaims{
		UserId: userId,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenStr, err := token.SignedString([]byte(os.Getenv("SecretKey")))
	if err != nil {
		return "", fmt.Errorf("internal server error: failed signed token: %w", err)
	}
	return tokenStr, nil
}

func GenerateToken(userId int, role string) (*claims.Token, error) {
	const ttlAccToken = 3 * time.Minute
	const ttlRefToken = 5 * 24 * time.Hour
	accToken, err := generateToken(userId, ttlAccToken, role)
	if err != nil {
		return nil, err
	}
	refToken, err := generateToken(userId, ttlRefToken, role)
	if err != nil {
		return nil, err
	}
	token := &claims.Token{
		AccessToken:  accToken,
		RefreshToken: refToken,
	}
	return token, nil
}

func ValidateToken(tokenStr string) (*claims.JWTClaims, error) {
	data := &claims.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, data,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("SecretKey")), nil
		})
	if err != nil {
		return &claims.JWTClaims{}, fmt.Errorf("internal server error: failed parse token: %w", err)
	}
	if !token.Valid {
		return &claims.JWTClaims{}, fmt.Errorf("invalid token")
	}
	return data, nil
}
