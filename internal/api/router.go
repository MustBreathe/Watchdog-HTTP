package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type Router struct {
	db *sql.DB
}

func NewRouter(db *sql.DB) *Router {
	return &Router{
		db: db,
	}
}

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Incoming] %-4s - %s", r.Method, r.RequestURI)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.RequestURI == "/health" {
		w.WriteHeader(http.StatusOK)
		resp := Response{
			Message: "Healthy",
			Status:  "ok",
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
