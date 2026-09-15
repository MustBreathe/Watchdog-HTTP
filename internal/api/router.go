package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"watchdog/main.go/internal/monitor"
)

type Router struct {
	db  *sql.DB
	mux *http.ServeMux
}

func NewRouter(db *sql.DB) *Router {
	mux := http.NewServeMux()
	registerHandlers(mux, db)
	return &Router{
		db:  db,
		mux: mux,
	}
}

func registerHandlers(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("GET /health", healthHandler)
	monitorHandler := NewMonitorHandler(monitor.CreateMonitorService(db))

	mux.HandleFunc("POST /api/v1/monitors", monitorHandler.Create)
	mux.HandleFunc("GET /api/v1/monitors", monitorHandler.List)
	mux.HandleFunc("GET /api/v1/monitors/{id}", monitorHandler.Get)
	mux.HandleFunc("PUT /api/v1/monitors/{id}", monitorHandler.Update)
	mux.HandleFunc("PATCH /api/v1/monitors/{id}/enable", monitorHandler.Enable)
	mux.HandleFunc("PATCH /api/v1/monitors/{id}/disable", monitorHandler.Disable)
	mux.HandleFunc("DELETE /api/v1/monitors/{id}", monitorHandler.Delete)
}

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Incoming] %-4s - %s", r.Method, r.RequestURI)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	handler, pattern := router.mux.Handler(r)
	if pattern == "" {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	handler.ServeHTTP(w, r)

}

func writeJSON(w http.ResponseWriter, status int, value any) {
	body, err := json.Marshal(value)

	if err != nil {
		writeError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, Response{
		Message: message,
		Status:  strconv.Itoa(status),
	})
}
