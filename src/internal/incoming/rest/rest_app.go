package rest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
)

type RestApp struct {
	config          *config.Config
	greetingService service.GreetingService
}

func NewRestApp(cfg *config.Config, greetingService service.GreetingService) (*RestApp, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	if greetingService == nil {
		return nil, fmt.Errorf("greetingService must not be nil")
	}

	return &RestApp{
		config:          cfg,
		greetingService: greetingService,
	}, nil
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return nil
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.ErrorContext(r.Context(), "encoding json response", "err", err, "status", status)
	}

	return nil
}

const maxRequestBodyBytes = 1 << 20

func decodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	body := http.MaxBytesReader(nil, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		// Die Fehlermeldung des Decoders kann Teile des Bodys enthalten und
		// damit potenziell PII. Sie wird daher nicht an den Client gegeben,
		// sondern nur intern verarbeitet.
		return fmt.Errorf("decoding request body: %w", err)
	}

	if decoder.More() {
		return fmt.Errorf("request body contains more than one json document")
	}

	return nil
}
