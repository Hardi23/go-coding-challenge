package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
)

const routerBasePath = "code-challenge"

func newRouter(t *testing.T, cfg *config.Config) http.Handler {
	t.Helper()

	app, err := NewRestApp(cfg, service.NewGreetingService())
	require.NoError(t, err)

	return app.NewRouter()
}

func routerConfig(stage string, enableSwaggerUi bool) *config.Config {
	return &config.Config{
		Stage:           stage,
		BasePath:        routerBasePath,
		LogLevel:        "error",
		EnableSwaggerUi: enableSwaggerUi,
		MaxEntries:      10,
	}
}

func TestNewRouter_SwaggerUiToggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		stage           string
		enableSwaggerUi bool
		expectAvailable bool
	}{
		{name: "dev with flag enabled", stage: "dev", enableSwaggerUi: true, expectAvailable: true},
		{name: "ts with flag enabled", stage: "ts", enableSwaggerUi: true, expectAvailable: true},
		{name: "dev with flag disabled", stage: "dev", enableSwaggerUi: false, expectAvailable: false},
		{name: "prod ignores enabled flag", stage: "prod", enableSwaggerUi: true, expectAvailable: false},
		{name: "prod with flag disabled", stage: "prod", enableSwaggerUi: false, expectAvailable: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			router := newRouter(t, routerConfig(tc.stage, tc.enableSwaggerUi))

			expectedStatus := http.StatusNotFound
			if tc.expectAvailable {
				expectedStatus = http.StatusOK
			}

			for _, path := range []string{"/swagger-ui", "/openapi.yaml"} {
				req := httptest.NewRequestWithContext(
					t.Context(), http.MethodGet, "/"+routerBasePath+path, nil,
				)
				res := httptest.NewRecorder()
				router.ServeHTTP(res, req)

				assert.Equal(t, expectedStatus, res.Code, "path %s", path)
			}
		})
	}
}

func TestNewRouter_RoutesRespectBasePath(t *testing.T) {
	t.Parallel()

	cfg := routerConfig("dev", false)
	cfg.BasePath = "schadenfall-service"
	router := newRouter(t, cfg)

	req := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "/schadenfall-service/greeting?name=Welt", nil,
	)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	staleReq := httptest.NewRequestWithContext(
		t.Context(), http.MethodGet, "/code-challenge/greeting?name=Welt", nil,
	)
	staleRes := httptest.NewRecorder()
	router.ServeHTTP(staleRes, staleReq)
	assert.Equal(t, http.StatusNotFound, staleRes.Code)
}

func TestNewRouter_GreetingMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method         string
		expectedStatus int
	}{
		{method: http.MethodGet, expectedStatus: http.StatusOK},
		{method: http.MethodPut, expectedStatus: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, expectedStatus: http.StatusMethodNotAllowed},
		{method: http.MethodPatch, expectedStatus: http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			t.Parallel()

			router := newRouter(t, routerConfig("dev", false))

			req := httptest.NewRequestWithContext(
				t.Context(), tc.method, "/"+routerBasePath+"/greeting?name=Welt", nil,
			)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
		})
	}
}

func TestNewRouter_HealthRejectsNonGet(t *testing.T) {
	t.Parallel()

	router := newRouter(t, routerConfig("dev", false))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/"+routerBasePath+"/health", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
}
