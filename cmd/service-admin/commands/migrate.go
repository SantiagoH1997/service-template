package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/data/schema"
	"github.com/santiagoh1997/service-template/internal/pkg/database"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

// Migrate creates the schema in the database.
func Migrate(cfg database.Config) error {
	db, err := database.NewDBClient(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	if err := schema.Migrate(db); err != nil {
		return errors.Wrap(err, "migrate database")
	}

	fmt.Println("migrations complete")
	return nil
}
