package model

import "errors"

var (
	ErrInvalidRequest = errors.New("the provided request is invalid")
	ErrNotFound       = errors.New("the requested resource was not found")
	ErrLimitReached   = errors.New("the configured limit was reached")
)
