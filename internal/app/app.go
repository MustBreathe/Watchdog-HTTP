package app

import (
	"cmd/watchdog/main.go/internal/api"
	"net/http"
	"os"
	"time"
)

type App struct {
	httpServer *http.Server
	router     *api.Router
}

func NewApplication() (*App, error) {
	router := api.NewRouter()

	httpServer := &http.Server{
		Addr:         os.Getenv("HTTP_ADDR"),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		httpServer: httpServer,
		router:     router,
	}, nil
}

func (app *App) Run() error {
	err := app.httpServer.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
	return nil
}
