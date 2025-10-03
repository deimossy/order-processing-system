package auth

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/deimossy/order-processing-system/internal/user/domain"
	errs "github.com/deimossy/order-processing-system/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

func ParseAndValidateAccessToken(token string, pubKey *rsa.PublicKey) (*domain.AccessTokenClaims, error) {
	claims := &domain.AccessTokenClaims{}

	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errs.ErrAccessTokenInvalid
		}
		return pubKey, nil
	}, jwt.WithLeeway(2*time.Minute))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errs.ErrAccessTokenExpired
		}
		return nil, err
	}

	if !parsed.Valid {
		return nil, errs.ErrAccessTokenInvalid
	}

	if !jwtFieldsValid(claims) {
		return nil, errs.ErrAccessTokenInvalid
	}

	return claims, nil
}

func jwtFieldsValid(claims *domain.AccessTokenClaims) bool {
	if claims.Subject == "" || claims.ID == "" || claims.IssuedAt == nil || claims.Email == "" {
		return false
	} else {
		return true
	}
}
