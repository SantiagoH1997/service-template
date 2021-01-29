package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/service"
)

// UserRepository is in charge of communicating with the persistence layer.
type UserRepository struct {
	db *sqlx.DB
}

// NewRepository returns a new instance of the repository struct.
// This struct implements the service.Repository interface.
func NewRepository(db *sqlx.DB) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("db parameter can't be nil")
	}
	return &UserRepository{db}, nil
}

// Create saves a User in the DB.
func (ur *UserRepository) Create(ctx context.Context, u service.User, now time.Time) (service.User, error) {
	u.DateCreated = now.UTC()
	u.DateUpdated = now.UTC()

	const q = `INSERT INTO users
	(user_id, email, password_hash, roles, name, last_name, country, date_created, date_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`
	if _, err := ur.db.ExecContext(ctx, q, u.ID, u.Email, u.PasswordHash, u.Roles, u.Name, u.LastName, u.Country, u.DateCreated, u.DateUpdated); err != nil {
		return service.User{}, errors.Wrap(err, "inserting user")
	}
	return u, nil
}

// Update changes certain fields of a saved User.
func (ur *UserRepository) Update(ctx context.Context, userID, name, lastName, country string, now time.Time) error {
	if _, err := uuid.Parse(userID); err != nil {
		return service.ErrInvalidID
	}

	const q = `
	UPDATE
		users
	SET
		"name" = $1,
		"last_name" = $2,
		"country" = $3,
		"date_updated" = $4
	WHERE
		user_id=$5`

	if _, err := ur.db.ExecContext(ctx, q, name, lastName, country, now.UTC(), userID); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete deletes a User from the DB.
func (ur *UserRepository) Delete(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return service.ErrInvalidID
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = $1`

	if _, err := ur.db.ExecContext(ctx, q, userID); err != nil {
		return errors.Wrapf(err, "deleting user %s", userID)
	}

	return nil
}

// GetByID retrieves a User from the DB by its ID.
func (ur *UserRepository) GetByID(ctx context.Context, userID string) (service.User, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return service.User{}, service.ErrInvalidID
	}

	const q = `SELECT * FROM users WHERE user_id = $1`

	var u service.User
	if err := ur.db.GetContext(ctx, &u, q, userID); err != nil {
		if err == sql.ErrNoRows {
			return service.User{}, service.ErrNotFound
		}
		return service.User{}, errors.Wrapf(err, "selecting user %q", userID)
	}

	return u, nil
}

// CheckEmailInUse returns true if a given email is already being used.
func (ur UserRepository) CheckEmailInUse(ctx context.Context, email string) (bool, error) {
	// TODO: improve...
	const q1 = `SELECT COUNT(*) FROM users AS numUsers WHERE email=$1`

	var numUsers int
	if err := ur.db.QueryRowContext(ctx, q1, email).Scan(&numUsers); err != nil {
		return false, errors.Wrapf(err, "looking for users with the email %s", email)
	}

	if numUsers != 0 {
		return true, nil
	}

	return false, nil
}

// GetByEmail retrieves a User by its email.
func (ur UserRepository) GetByEmail(ctx context.Context, email string) (service.User, error) {
	const q = `SELECT * FROM users WHERE email = $1`

	var u service.User
	if err := ur.db.GetContext(ctx, &u, q, email); err != nil {
		if err == sql.ErrNoRows {
			return service.User{}, service.ErrNotFound
		}
		return service.User{}, errors.Wrapf(err, "selecting user %q", email)
	}

	return u, nil
}
