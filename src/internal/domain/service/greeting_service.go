package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
)

type GreetingService interface {
	Greet(ctx context.Context, name string) (*model.GreetingResponse, error)
}

type greetingService struct {
	greetingTemplate string
}

func NewGreetingService() GreetingService {
	return &greetingService{
		greetingTemplate: "Hallo, %s!",
	}
}

func (s *greetingService) Greet(ctx context.Context, name string) (*model.GreetingResponse, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("%w: name must not be empty", model.ErrInvalidRequest)
	}

	slog.DebugContext(ctx, "building greeting", "nameLength", len(trimmed))

	return &model.GreetingResponse{
		Message: fmt.Sprintf(s.greetingTemplate, trimmed),
	}, nil
}
