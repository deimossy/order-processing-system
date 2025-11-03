package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/deimossy/order-processing-system/internal/user/domain"
	"github.com/google/uuid"
)

func GenerateRefreshToken(userID string, ttl time.Duration) (*domain.RefreshToken, string, error) {
	plain, err := generatePlainToken(64)
	if err != nil {
		return nil, "", err
	}

	hashed := HashToken(plain)

	token := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: hashed,
		ExpiresAt: time.Now().UTC().Add(ttl),
		CreatedAt: time.Now().UTC(),
	}

	return token, plain, nil
}

func generatePlainToken(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func HashToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(h[:])
}
