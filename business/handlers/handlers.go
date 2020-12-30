// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/santiagoh1997/service-template/business/auth"
	"github.com/santiagoh1997/service-template/business/mid"
	"github.com/santiagoh1997/service-template/business/service"
	"github.com/santiagoh1997/service-template/foundation/web"
)

// NewHTTPHandler constructs an http.Handler with all the application routes defined.
func NewHTTPHandler(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *sqlx.DB) http.Handler {

	// The web.App holds all routes and all the common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register debug check endpoints.
	ch := checkHandler{
		build: build,
		db:    db,
	}
	app.HandleDebug(http.MethodGet, "/readiness", ch.readiness)
	app.HandleDebug(http.MethodGet, "/liveness", ch.liveness)

	// Register main endpoints.
	uh := userHandler{
		svc:  service.New(log, db),
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
