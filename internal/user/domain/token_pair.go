package domain

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
}
