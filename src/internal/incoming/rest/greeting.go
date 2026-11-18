package rest

import (
	"fmt"
	"log/slog"
	"net/http"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
)

func (app *RestApp) HandleGetGreeting(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	slog.DebugContext(ctx, "handling get greeting")
	request := model.GreetingRequest{Name: r.URL.Query().Get("name")}
	if err := request.Validate(); err != nil {
		return fmt.Errorf("validating greeting query: %w", err)
	}

	response, err := app.greetingService.Greet(ctx, request.Name)
	if err != nil {
		return fmt.Errorf("get greeting: %w", err)
	}

	return writeJSON(r.Context(), w, http.StatusOK, response)
}

func (app *RestApp) HandlePostGreeting(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	slog.DebugContext(ctx, "handling post greeting")

	var request model.GreetingRequest
	if err := decodeJSON(r, &request); err != nil {
		return fmt.Errorf("%w: %w", model.ErrInvalidRequest, err)
	}

	if err := request.Validate(); err != nil {
		return fmt.Errorf("validating greeting body: %w", err)
	}

	response, err := app.greetingService.Greet(ctx, request.Name)
	if err != nil {
		return fmt.Errorf("post greeting: %w", err)
	}

	return writeJSON(r.Context(), w, http.StatusOK, response)
}
