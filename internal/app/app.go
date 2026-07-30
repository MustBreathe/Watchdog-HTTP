package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"watchdog/main.go/internal/api"
	"watchdog/main.go/internal/storage"
)

type App struct {
	httpServer *http.Server
	router     *api.Router
	logger     *io.Writer
	db         *sql.DB
}

func NewApplication(out io.Writer) (*App, error) {

	if out == nil {
		out = io.Discard
		return nil, errors.New("logger context not exists, please provide a valid logger")
	}

	fmt.Fprintln(out, "Initializing server...")

	db, err := storage.Open(out)

	if err != nil {
		return nil, fmt.Errorf("Can't open database %s, %w", os.Getenv("WATCHDOG_DATABASE_DSN"), err)
	}

	router := api.NewRouter(db)

	httpServer := &http.Server{
		Addr:         os.Getenv("WATCHDOG_HTTP_ADDR"),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	fmt.Fprintln(out, "Server will listen on", os.Getenv("WATCHDOG_HTTP_ADDR"))

	return &App{
		httpServer: httpServer,
		router:     router,
		logger:     &out,
		db:         db,
	}, nil
}

func (app *App) Run(ctx context.Context) error {

	err := app.httpServer.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
	return nil
}
