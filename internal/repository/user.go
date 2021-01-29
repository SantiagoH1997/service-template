package repository

import (
	"time"

	"github.com/lib/pq"
)

// User represents an individual user.
type User struct {
	ID           string         `db:"user_id"`
	Name         string         `db:"name"`
	LastName     string         `db:"last_name"`
	Email        string         `db:"email"`
	Country      string         `db:"country"`
	Roles        pq.StringArray `db:"roles" json:"roles"`
	PasswordHash []byte         `db:"password_hash" json:"-"`
	DateCreated  time.Time      `db:"date_created" json:"date_created"`
	DateUpdated  time.Time      `db:"date_updated" json:"date_updated"`
}
