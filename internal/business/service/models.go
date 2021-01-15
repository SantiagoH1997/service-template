package service

import (
	"time"

	"github.com/lib/pq"
)

// User represents an individual user.
type User struct {
	ID           string         `db:"user_id" json:"id"`
	Name         string         `db:"name" json:"name"`
	LastName     string         `db:"last_name"`
	Email        string         `db:"email" json:"email"`
	Country      string         `db:"country"`
	Roles        pq.StringArray `db:"roles" json:"roles"`
	PasswordHash []byte         `db:"password_hash" json:"-"`
	DateCreated  time.Time      `db:"date_created" json:"date_created"`
	DateUpdated  time.Time      `db:"date_updated" json:"date_updated"`
}

// NewUserRequest contains all the needed data to create a User.
type NewUserRequest struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	LastName        string   `json:"last_name" validate:"required"`
	Country         string   `json:"country" validate:"required"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

// UpdateUserRequest contains the information needed to modify an existing User.
type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email" validate:"omitempty,email"`
	LastName string `json:"last_name" validate:"required"`
	Country  string `json:"country" validate:"required"`
}

// LoginRequest is used in order to authenticate a client.
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}
