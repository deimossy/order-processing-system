package errors

import "errors"

var (
	ErrNotFound                   = errors.New("not found")
	ErrAlreadyExists              = errors.New("already exists")
	ErrValidation                 = errors.New("validation error")
	ErrTokenExpired               = errors.New("token has been expires")
	ErrTokenInvalid               = errors.New("token is invalid")
	ErrMaxRetriesAttemptsExceeded = errors.New("max retries attempts exceeded")
)
