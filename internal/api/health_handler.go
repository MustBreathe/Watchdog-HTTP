package api

import (
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {

	writeJSON(w, http.StatusOK, Response{
		Message: "Healthy",
		Status:  "ok",
	})

}
