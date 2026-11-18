package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/config"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/incoming/rest"
	corsmiddleware "gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/pkg/cors"
)

const TestBasePath = "code-challenge"

func NewTestConfig() *config.Config {
	return &config.Config{
		Stage:                    "dev",
		BasePath:                 TestBasePath,
		Port:                     8080,
		LogLevel:                 "error",
		EnableSwaggerUi:          false,
		MaxEntries:               10,
		ReadHeaderTimeoutSeconds: 10,
		ShutdownTimeoutSeconds:   15,
		CorsConfig:               corsmiddleware.Config{},
	}
}

func NewTestHandler(t *testing.T, greetingService service.GreetingService) http.Handler {
	t.Helper()

	if greetingService == nil {
		greetingService = service.NewGreetingService()
	}

	app, err := rest.NewRestApp(NewTestConfig(), greetingService)
	if err != nil {
		t.Fatalf("creating rest app: %v", err)
	}

	return app.NewRouter()
}

func ExecuteRequest(t *testing.T, handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}
