// Package service contains user related CRUD functionality.
package service

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/metrics"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/business/auth"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrDuplicatedEmail is used whenever someone attempts to create a User
	// with an email that's already being used.
	ErrDuplicatedEmail = errors.New("email already in use")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// UserService manages the set of API's for user access.
type UserService interface {
	Create(ctx context.Context, traceID string, nur NewUserRequest, now time.Time) (User, error)
	Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uur UpdateUserRequest, now time.Time) error
	Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error
	GetByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (User, error)
	Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error)
}

type userService struct {
	repo Repository
}

// NewBasicService constructs a UserService for api access.
func NewBasicService(repo Repository) (UserService, error) {
	if repo == nil {
		return nil, errors.New("repo can't be nil")
	}

	return userService{
		repo: repo,
	}, nil
}

// New returns a UserService with instrumentation features.
func New(repo Repository, requestCount metrics.Counter, requestLatency metrics.Histogram) (UserService, error) {
	us, err := NewBasicService(repo)
	if err != nil {
		return nil, errors.Wrap(err, "creating service")
	}

	us, err = NewInstrumentingDecorator(requestCount, requestLatency, us)
	if err != nil {
		return nil, errors.Wrap(err, "creating service")
	}

	return us, nil
}

// Create creates a new user, generating a password hash.
func (us userService) Create(ctx context.Context, traceID string, nur NewUserRequest, now time.Time) (User, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.service.create")
	defer span.End()

	isInUse, err := us.repo.CheckEmailInUse(ctx, nur.Email)
	if err != nil {
		return User{}, errors.Wrap(err, "creating user")
	}
	if isInUse {
		return User{}, ErrDuplicatedEmail
	}

	// Generate hash from passowrd.
	hash, err := bcrypt.GenerateFromPassword([]byte(nur.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         nur.Name,
		LastName:     nur.LastName,
		Email:        nur.Email,
		Country:      nur.Country,
		PasswordHash: hash,
		Roles:        nur.Roles,
	}

	saved, err := us.repo.Create(ctx, u, now)
	if err != nil {
		return User{}, errors.Wrap(err, "inserting user")
	}

	return saved, nil
}

// Update allows a client to update certain fields of a saved User.
func (us userService) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uur UpdateUserRequest, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.service.update")
	defer span.End()

	u, err := us.GetByID(ctx, traceID, claims, userID)
	if err != nil {
		return err
	}

	if err = us.repo.Update(ctx, u.ID, uur.Name, uur.LastName, uur.Country, now); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete deletes a User by its ID.
func (us userService) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.service.delete")
	defer span.End()

	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return ErrForbidden
	}

	if err := us.repo.Delete(ctx, userID); err != nil {
		switch err {
		case ErrInvalidID:
			return ErrInvalidID
		default:
			return errors.Wrapf(err, "deleting user %s", userID)
		}
	}

	return nil
}

// GetByID retrieves a User by its ID.
func (us userService) GetByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (User, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.service.getById")
	defer span.End()

	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return User{}, ErrForbidden
	}

	u, err := us.repo.GetByID(ctx, userID)
	if err != nil {
		switch err {
		case ErrInvalidID:
			return User{}, ErrInvalidID
		case ErrNotFound:
			return User{}, ErrNotFound
		default:
			return User{}, errors.Wrapf(err, "searching for user %q", userID)
		}
	}

	return u, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims representing the user. The claims can be
// used to generate a token for future authentication.
func (us userService) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.service.authenticate")
	defer span.End()

	u, err := us.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == ErrNotFound {
			return auth.Claims{}, ErrAuthenticationFailure
		}
		return auth.Claims{}, errors.Wrap(err, "selecting single user")
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	claims := auth.Claims{
		// TODO: Customize claims to suit the project.
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service template",
			Subject:   u.ID,
			Audience:  "clients",
			ExpiresAt: now.Add(time.Hour).Unix(),
			IssuedAt:  now.Unix(),
		},
		Roles: u.Roles,
	}

	return claims, nil
}
