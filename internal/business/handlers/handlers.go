// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-kit/kit/metrics"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/santiagoh1997/service-template/internal/business/auth"
	"github.com/santiagoh1997/service-template/internal/business/mid"
	"github.com/santiagoh1997/service-template/internal/business/service"
	"github.com/santiagoh1997/service-template/internal/pkg/web"
)

// NewHTTPHandler constructs an http.Handler with all the application routes defined.
func NewHTTPHandler(
	build string,
	shutdown chan os.Signal,
	log *log.Logger,
	errorCount metrics.Counter,
	redMetrics metrics.Histogram,
	requestCount metrics.Counter,
	requestLatency metrics.Histogram,
	a *auth.Auth,
	db *sqlx.DB,
) http.Handler {

	// The web.App holds all routes and all the common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(errorCount, redMetrics), mid.Panics(log))

	// Register debug check endpoints.
	ch := checkHandler{
		build: build,
		db:    db,
	}
	app.HandleDebug(http.MethodGet, "/readiness", ch.readiness)
	app.HandleDebug(http.MethodGet, "/liveness", ch.liveness)

	// Register metrics endpoint.
	ph := promhttp.Handler()
	prometheusHandler := func(_ context.Context, w http.ResponseWriter, r *http.Request) error {
		ph.ServeHTTP(w, r)
		return nil
	}
	app.HandleDebug(http.MethodGet, "/metrics", prometheusHandler)

	// Register main endpoints.
	us := service.New(log, db)
	us = service.NewInstrumentingUserService(requestCount, requestLatency, us)
	uh := userHandler{
		svc:  us,
		auth: a,
	}
	app.Handle(http.MethodGet, "/v1/users/:page/:rows", uh.getAll, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/token/:kid", uh.token)
	app.Handle(http.MethodGet, "/v1/users/:id", uh.getByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/users", uh.create)
	app.Handle(http.MethodPut, "/v1/users/:id", uh.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/users/:id", uh.delete, mid.Authenticate(a))

	return app
}
