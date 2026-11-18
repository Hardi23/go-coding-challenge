package _openapi

import (
	"embed"
	"log/slog"
	"net/http"
)

//go:embed openapi.yaml
var specFS embed.FS

const specFileName = "openapi.yaml"

func RegisterSwaggerUi(mux *http.ServeMux, basePath string) {
	specPath := basePath + "/" + specFileName
	uiPath := basePath + "/swagger-ui"

	mux.HandleFunc("GET "+specPath, func(w http.ResponseWriter, r *http.Request) {
		data, err := specFS.ReadFile(specFileName)
		if err != nil {
			slog.ErrorContext(r.Context(), "reading embedded openapi spec", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		if _, err := w.Write(data); err != nil {
			slog.ErrorContext(r.Context(), "writing openapi spec", "err", err)
		}
	})

	mux.HandleFunc("GET "+uiPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(swaggerUiHTML(specPath))); err != nil {
			slog.ErrorContext(r.Context(), "writing swagger ui", "err", err)
		}
	})

	slog.Info("swagger ui registered", "ui", uiPath, "spec", specPath)
}

func swaggerUiHTML(specPath string) string {
	return `<!DOCTYPE html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>golang-code-challenge – API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "` + specPath + `",
        dom_id: "#swagger-ui",
        deepLinking: true,
        tryItOutEnabled: true
      });
    };
  </script>
</body>
</html>`
}
