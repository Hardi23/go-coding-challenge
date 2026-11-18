package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/incoming/rest"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/test/util"
)

var greetingPath = "/" + util.TestBasePath + "/greeting"

func TestGetGreeting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		query           string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "returns greeting for valid name",
			query:           "?name=Welt",
			expectedStatus:  http.StatusOK,
			expectedMessage: "Hallo, Welt!",
		},
		{
			name:            "url encoded name is decoded",
			query:           "?name=Max%20Mustermann",
			expectedStatus:  http.StatusOK,
			expectedMessage: "Hallo, Max Mustermann!",
		},
		{
			name:           "missing query parameter yields 400",
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty query parameter yields 400",
			query:          "?name=",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "name exceeding 100 characters yields 400",
			query:          "?name=" + strings.Repeat("a", 101),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := util.NewTestHandler(t, nil)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, greetingPath+tc.query, nil)

			res := util.ExecuteRequest(t, handler, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
			assert.Contains(t, res.Header().Get("Content-Type"), "application/json")

			if tc.expectedStatus != http.StatusOK {
				var errorBody rest.ErrorResponse
				require.NoError(t, json.Unmarshal(res.Body.Bytes(), &errorBody))
				assert.NotEmpty(t, errorBody.Message)
				return
			}

			var body model.GreetingResponse
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
			assert.Equal(t, tc.expectedMessage, body.Message)
		})
	}
}

func TestPostGreeting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		body            string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "returns greeting for valid body",
			body:            `{"name":"Welt"}`,
			expectedStatus:  http.StatusOK,
			expectedMessage: "Hallo, Welt!",
		},
		{
			name:           "empty name yields 400",
			body:           `{"name":""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed json yields 400",
			body:           `{"name":`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body yields 400",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unknown field yields 400",
			body:           `{"name":"Welt","unexpected":"value"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "name exceeding 100 characters yields 400",
			body:           fmt.Sprintf(`{"name":%q}`, strings.Repeat("a", 101)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "multiple json documents yield 400",
			body:           `{"name":"Welt"}{"name":"Zwei"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "oversized body yields 400",
			body:           fmt.Sprintf(`{"name":%q}`, strings.Repeat("a", 2<<20)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := util.NewTestHandler(t, nil)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, greetingPath, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			res := util.ExecuteRequest(t, handler, req)

			assert.Equal(t, tc.expectedStatus, res.Code)

			if tc.expectedStatus != http.StatusOK {
				return
			}

			var body model.GreetingResponse
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
			assert.Equal(t, tc.expectedMessage, body.Message)
		})
	}
}

func TestGreetingServiceFailureYieldsInternalServerError(t *testing.T) {
	t.Parallel()

	serviceMock := util.NewGreetingServiceMock()
	serviceMock.
		On("Greet", mock.Anything, "Welt").
		Return(nil, fmt.Errorf("downstream system unavailable"))

	handler := util.NewTestHandler(t, serviceMock)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, greetingPath+"?name=Welt", nil)

	res := util.ExecuteRequest(t, handler, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	var errorBody rest.ErrorResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &errorBody))
	assert.Equal(t, "internal server error", errorBody.Message)
	assert.NotContains(t, res.Body.String(), "downstream system unavailable")

	serviceMock.AssertExpectations(t)
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := util.NewTestHandler(t, nil)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, greetingPath, nil)

	res := util.ExecuteRequest(t, handler, req)

	assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
}

func TestUnknownPath(t *testing.T) {
	t.Parallel()

	handler := util.NewTestHandler(t, nil)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/"+util.TestBasePath+"/does-not-exist", nil)

	res := util.ExecuteRequest(t, handler, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetHealth(t *testing.T) {
	t.Parallel()

	handler := util.NewTestHandler(t, nil)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/"+util.TestBasePath+"/health", nil)

	res := util.ExecuteRequest(t, handler, req)

	require.Equal(t, http.StatusOK, res.Code)

	var body rest.HealthResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	assert.Equal(t, "UP", body.Status)
	assert.Equal(t, "dev", body.Stage)
}
