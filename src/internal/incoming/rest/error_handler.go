package rest

import (
	"errors"
	"log/slog"
	"net/http"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func errorHandler(next HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err == nil {
			return
		}

		status, clientMessage := mapError(err)

		logAttrs := []any{
			"err", err,
			"status", status,
			"method", r.Method,
			"route", r.Pattern,
		}

		if status >= http.StatusInternalServerError {
			slog.ErrorContext(r.Context(), "request failed", logAttrs...)
		} else {
			slog.WarnContext(r.Context(), "request rejected", logAttrs...)
		}
		err = writeJSON(r.Context(), w, status, ErrorResponse{Message: clientMessage})
		if err != nil {
			slog.ErrorContext(r.Context(), "error writing response", "err", err)
		}
	}
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, model.ErrInvalidRequest):
		return http.StatusBadRequest, "bad request"

	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, "not found"

	case errors.Is(err, model.ErrLimitReached):
		return http.StatusConflict, "limit reached"

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
