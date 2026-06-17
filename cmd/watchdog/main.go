package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	appcli "cmd/watchdog/main.go/internal/app/cli"
)

const MIN_NUMBER_OF_IMPUT_PARAMETERS = 2

func main() {

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)

	cli := appcli.CLI()

	cli.Execute(os.Args[1:])

	// s := <-c
	// fmt.Println("Cleanup...", s)
	// os.Exit(0)
}

type ServerStrategy struct{}

func (ss ServerStrategy) Execute() {

	port := os.Getenv("HTTP_ADDR")

	s := http.Server{
		Addr:         port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      HelloHandler{},
	}
	err := s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}

type HelloHandler struct{}

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (hh HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
