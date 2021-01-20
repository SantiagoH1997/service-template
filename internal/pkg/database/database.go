// Package database provides support for access the database.
package database

import (
	"context"
	"net/url"
	"time"

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

// NewDBClient runs the necessary migrations prior to returning
// returning a new DB client based on the configuration.
func NewDBClient(cfg Config) (*sqlx.DB, error) {
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

	db, err := openDB(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "open database connection")
	}

	if err := schema.Migrate(db); err != nil {
		return nil, errors.Wrap(err, "migrate database")
	}

	return db, nil
}

// StatusCheck returns an error if it can't successfully talk to the database.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "internal.pkg.database.statuscheck")
	defer span.End()

	// Running this query forces a round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

func openDB(URL string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", URL)
	if err != nil {
		return nil, errors.Wrap(err, "connect database")
	}

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var pingError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		return nil, errors.Wrap(err, "database never ready")
	}

	return db, nil
}
