package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/business/data/schema"
	"github.com/santiagoh1997/service-template/internal/pkg/database"
)

// Seed loads test data into the database.
func Seed(cfg database.Config) error {
	db, err := database.NewDBClient(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	if err := schema.Seed(db); err != nil {
		return errors.Wrap(err, "seed database")
	}

	fmt.Println("seed data complete")
	return nil
}
