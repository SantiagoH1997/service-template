package schema

import (
	"github.com/jmoiron/sqlx"
)

// Seed seeds the DB with example data.
// The queries are ran in a transaction and rolled back if any fail.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

const seeds = `
-- Create admin and regular User with password "password"
INSERT INTO users (user_id, name, last_name, email, country, roles, password_hash, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Example', 'Admin', 'admin@example.com', 'Albania', '{ADMIN,USER}', '$2a$10$v79.Q7kMpIZFH0QYi.IoieRZikqOIr8a7Mo5Xk59sjeexFeuG22Oq', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'Example', 'User', 'user@example.com', 'Algeria', '{USER}', '$2a$10$v79.Q7kMpIZFH0QYi.IoieRZikqOIr8a7Mo5Xk59sjeexFeuG22Oq', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
	ON CONFLICT DO NOTHING;
`

// DeleteAll is used to clean the database between tests.
// It runs the set of Drop-table queries against db.
// The queries are ran in a transaction and rolled back if any fail.
func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteAll); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

const deleteAll = `
DELETE FROM users;`
