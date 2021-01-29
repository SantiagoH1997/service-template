package service

import (
	"context"
	"time"
)

// Repository represents a persistance layer.
type Repository interface {
	Create(ctx context.Context, u User, now time.Time) (User, error)
	GetByID(ctx context.Context, userID string) (User, error)
	Update(ctx context.Context, userID, name, lastName, country string, now time.Time) error
	Delete(ctx context.Context, userID string) error
	GetByEmail(ctx context.Context, email string) (User, error)
	CheckEmailInUse(ctx context.Context, email string) (bool, error)
}
