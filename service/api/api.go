/*
Package api contains all HTTP handlers used by the WASAText backend.

This package is responsible for:
  • defining the HTTP router and registering all API endpoints
  • wiring requests to the database layer
  • providing the main entry point for the web API

To use this package, create a new Router with New() passing a valid Config.
The returned Router exposes a Handler() method that can be used as the
http.Server handler.

Example:

	// Create the API router
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: appdb,
	})
	if err != nil {
		logger.WithError(err).Error("cannot create API server instance")
		return err
	}
	router := apirouter.Handler()

	// Create the HTTP server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start listening for requests
	apiserver.ListenAndServe()

See cmd/webapi/main.go for the complete startup sequence.
*/
package api

import (
	"errors"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Config provides dependencies and configuration to New.
type Config struct {
	// Logger is the logger used by the API layer.
	Logger logrus.FieldLogger

	// Database is the application database implementation.
	Database database.AppDatabase
}

// Router is the public interface exposed by this package.
type Router interface {
	// Handler returns an HTTP handler that serves all WASAText APIs.
	Handler() http.Handler

	// Close releases any resource held by the router (if any).
	Close() error
}

// New creates and returns a new Router instance.
func New(cfg Config) (Router, error) {
	// Validate configuration.
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	// Create the underlying HTTP router where all endpoints are registered.
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	return &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}, nil
}

type _router struct {
	router *httprouter.Router

	// baseLogger is used outside of request contexts (e.g. background tasks).
	// For request-scoped logs, prefer the logger in the request context.
	baseLogger logrus.FieldLogger

	db database.AppDatabase
}
