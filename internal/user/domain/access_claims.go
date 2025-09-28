package domain

import "github.com/golang-jwt/jwt/v5"

// обертка над дефолтными клеймсами
type AccessTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
