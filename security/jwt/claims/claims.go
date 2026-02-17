package claims

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserId int
	Role   string
	jwt.RegisteredClaims
}

type Token struct {
	AccessToken  string
	RefreshToken string
}
