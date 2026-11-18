package rest

import (
	"net/http"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/incoming/rest/_openapi"
)

func (app *RestApp) NewRouter() http.Handler {
	mux := http.NewServeMux()
	base := "/" + app.config.BasePath

	mux.Handle("GET "+base+"/greeting", errorHandler(app.HandleGetGreeting))
	mux.Handle("POST "+base+"/greeting", errorHandler(app.HandlePostGreeting))

	healthHandler := errorHandler(app.HandleGetHealth)
	mux.Handle("GET "+base+"/health", healthHandler)

	if app.config.EnableSwaggerUi && !app.config.IsProd() {
		_openapi.RegisterSwaggerUi(mux, base)
	}

	return mux
}
