package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"education/internal/database"
	"education/internal/handlers"
	"education/internal/middleware"
	"education/internal/routes"
)

type App struct {
	Address string
	Stores  []*database.Store
	Router  http.Handler
}

func New() (*App, error) {
	application := &App{
		Address: "127.0.0.1:8000",
		Stores:  database.Catalog(),
	}

	for _, store := range application.Stores {
		if err := database.Open(store); err != nil {
			application.Close()
			return nil, err
		}
	}

	pageHandler, err := handlers.NewPageHandler()
	if err != nil {
		application.Close()
		return nil, err
	}

	recordHandler := handlers.NewRecordHandler(application.Stores)
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, recordHandler, pageHandler)
	application.Router = middleware.Chain(mux)

	return application, nil
}

func (a *App) Close() {
	for _, store := range a.Stores {
		if store.DB != nil {
			store.DB.Close()
		}
	}
}

func (a *App) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              a.Address,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopped := make(chan struct{})
	defer close(stopped)
	go func() {
		select {
		case <-ctx.Done():
			shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownContext); err != nil {
				server.Close()
			}
		case <-stopped:
		}
	}()

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
