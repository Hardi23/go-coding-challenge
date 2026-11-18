package rest

import "net/http"

type HealthResponse struct {
	Status string `json:"status"`
	Stage  string `json:"stage"`
}

func (app *RestApp) HandleGetHealth(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, r, http.StatusOK, HealthResponse{
		Status: "UP",
		Stage:  app.config.Stage,
	})
}
