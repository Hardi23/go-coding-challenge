package model

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxNameLength = 100

type GreetingRequest struct {
	Name string `json:"name"`
}

func (r *GreetingRequest) Validate() error {
	name := strings.TrimSpace(r.Name)

	if name == "" {
		return fmt.Errorf("%w: field 'name' must not be empty", ErrInvalidRequest)
	}

	if utf8.RuneCountInString(name) > maxNameLength {
		return fmt.Errorf("%w: field 'name' must not exceed %d characters", ErrInvalidRequest, maxNameLength)
	}

	return nil
}

type GreetingResponse struct {
	Message string `json:"message"`
}
