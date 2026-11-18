package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
)

func TestNewRestApp_RejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()

		app, err := NewRestApp(nil, service.NewGreetingService())

		require.Error(t, err)
		assert.Nil(t, app)
	})

	t.Run("nil greeting service", func(t *testing.T) {
		t.Parallel()

		app, err := NewRestApp(&config.Config{BasePath: "code-challenge"}, nil)

		require.Error(t, err)
		assert.Nil(t, app)
	})
}

func TestNewRestApp_Succeeds(t *testing.T) {
	t.Parallel()

	app, err := NewRestApp(&config.Config{BasePath: "code-challenge"}, service.NewGreetingService())

	require.NoError(t, err)
	require.NotNil(t, app)
}

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "invalid request maps to 400",
			err:             model.ErrInvalidRequest,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "bad request",
		},
		{
			name:            "not found maps to 404",
			err:             model.ErrNotFound,
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "not found",
		},
		{
			name:            "limit reached maps to 409",
			err:             model.ErrLimitReached,
			expectedStatus:  http.StatusConflict,
			expectedMessage: "limit reached",
		},
		{
			name:            "unclassified error maps to 500",
			err:             errors.New("downstream system unavailable"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "internal server error",
		},
		{
			name:            "wrapped sentinel is still recognised",
			err:             fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", model.ErrNotFound)),
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, message := mapError(tc.err)

			assert.Equal(t, tc.expectedStatus, status)
			assert.Equal(t, tc.expectedMessage, message)
		})
	}
}

func TestErrorHandler_PassesThroughOnSuccess(t *testing.T) {
	t.Parallel()

	handler := errorHandler(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusTeapot)
		return nil
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	assert.Equal(t, http.StatusTeapot, res.Code)
	assert.Empty(t, res.Body.String())
}

func TestErrorHandler_WritesErrorResponse(t *testing.T) {
	t.Parallel()

	const internalDetail = "connection refused on 10.0.0.1:5432"

	handler := errorHandler(func(_ http.ResponseWriter, _ *http.Request) error {
		return fmt.Errorf("querying database: %s", internalDetail)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.JSONEq(t, `{"message":"internal server error"}`, res.Body.String())
	assert.NotContains(t, res.Body.String(), internalDetail)
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	t.Run("writes payload and content type", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		res := httptest.NewRecorder()

		require.NoError(t, writeJSON(res, req, http.StatusCreated, map[string]string{"a": "b"}))

		assert.Equal(t, http.StatusCreated, res.Code)
		assert.Equal(t, "application/json; charset=utf-8", res.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"a":"b"}`, res.Body.String())
	})

	t.Run("nil payload writes status only", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		res := httptest.NewRecorder()

		require.NoError(t, writeJSON(res, req, http.StatusNoContent, nil))

		assert.Equal(t, http.StatusNoContent, res.Code)
		assert.Empty(t, res.Body.String())
	})

	t.Run("unencodable payload does not panic and keeps status", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/x", nil)
		res := httptest.NewRecorder()
		require.NoError(t, writeJSON(res, req, http.StatusOK, make(chan int)))

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

type decodeTarget struct {
	Name string `json:"name"`
}

func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		expectError bool
		expected    string
	}{
		{name: "valid body", body: `{"name":"Welt"}`, expected: "Welt"},
		{
			name:     "missing field decodes to zero value",
			body:     `{}`,
			expected: "",
		},
		{name: "malformed json", body: `{"name":`, expectError: true},
		{name: "empty body", body: ``, expectError: true},
		{name: "unknown field", body: `{"name":"Welt","x":1}`, expectError: true},
		{name: "multiple documents", body: `{"name":"a"}{"name":"b"}`, expectError: true},
		{name: "wrong type", body: `{"name":123}`, expectError: true},
		{name: "json array instead of object", body: `[{"name":"a"}]`, expectError: true},
		{
			name:        "oversized body",
			body:        fmt.Sprintf(`{"name":%q}`, strings.Repeat("a", 2*maxRequestBodyBytes)),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(
				t.Context(), http.MethodPost, "/x", strings.NewReader(tc.body),
			)

			var target decodeTarget
			err := decodeJSON(req, &target)

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, target.Name)
		})
	}
}

func TestDecodeJSON_NilBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/x", nil)
	req.Body = nil

	var target decodeTarget
	require.Error(t, decodeJSON(req, &target))
}
