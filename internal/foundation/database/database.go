// Package database provides support for access the database.
package database

import (
	"context"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // The database driver in use.
	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/business/data/schema"
	"go.opentelemetry.io/otel/trace"
)

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
}

// Open knows how to open a database connection based on the configuration.
// It runs the necessary migrations prior to returning the DB client.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, errors.Wrap(err, "connect database")
	}

	if err := schema.Migrate(db); err != nil {
		return nil, errors.Wrap(err, "migrate database")
	}

	return db, nil
}

// StatusCheck returns an error if it can't successfully talk to the database.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "foundation.database.statuscheck")
	defer span.End()

	// Running this query forces a round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}
