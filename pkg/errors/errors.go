package errors

import "errors"

var (
	ErrNotFound                   = errors.New("not found")
	ErrAlreadyExists              = errors.New("already exists")
	ErrValidation                 = errors.New("validation error")
	ErrMaxRetriesAttemptsExceeded = errors.New("max retries attempts exceeded")
)
