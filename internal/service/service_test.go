package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/santiagoh1997/service-template/internal/auth"
	"github.com/santiagoh1997/service-template/internal/data/schema"
	"github.com/santiagoh1997/service-template/internal/repository"
	"github.com/santiagoh1997/service-template/internal/service"
	"github.com/santiagoh1997/service-template/internal/tests"
)

func TestCreate(t *testing.T) {
	_, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	ur, _ := repository.NewRepository(db)
	us, _ := service.NewBasicService(ur)
	ctx := context.Background()
	now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
	traceID := "00000000-0000-0000-0000-000000000000"

	t.Run("Success case", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}

		u, err := us.Create(ctx, traceID, nur, now)
		if err != nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
		}

		// Retrieving saved user and comparing...
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   u.ID,
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleUser},
		}

		saved, err := us.GetByID(ctx, traceID, claims, u.ID)
		if err != nil {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, nil)
		}

		if diff := cmp.Diff(u, saved); diff != "" {
			t.Fatalf("\t%s\tGetByID() user = %v, diff:\n%s", tests.Failed, saved, diff)
		}
	})

	t.Run("Duplicated email case", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Fernández",
			Country:         "Colombia",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}

		_, err := us.Create(ctx, traceID, nur, now)
		if err == nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want err", tests.Failed, err)
		}
	})

}

func TestUpdate(t *testing.T) {
	_, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	ur, _ := repository.NewRepository(db)
	us, _ := service.NewBasicService(ur)
	ctx := context.Background()
	now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
	traceID := "00000000-0000-0000-0000-000000000000"

	t.Run("Success case", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}

		u, err := us.Create(ctx, traceID, nur, now)
		if err != nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
		}

		uur := service.UpdateUserRequest{
			Name:     "Jorgito",
			LastName: "Porcel",
			Country:  "Argentina",
		}

		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   u.ID,
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleUser},
		}

		if err = us.Update(ctx, traceID, claims, u.ID, uur, now); err != nil {
			tt.Fatalf("\t%s\tUpdate() err = %v, want %v", tests.Failed, err, nil)
		}
	})

	t.Run("Invalid ID case", func(tt *testing.T) {
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		uur := service.UpdateUserRequest{
			Name:     "Jorgito",
			LastName: "Porcel",
			Country:  "Argentina",
		}

		if err := us.Update(ctx, traceID, claims, "invalidID", uur, now); err != service.ErrInvalidID {
			tt.Fatalf("\t%s\tUpdate() err = %v, want %v", tests.Failed, err, service.ErrInvalidID)
		}
	})
}

func TestDelete(t *testing.T) {
	_, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	ur, _ := repository.NewRepository(db)
	us, _ := service.NewBasicService(ur)
	ctx := context.Background()
	now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
	traceID := "00000000-0000-0000-0000-000000000000"

	t.Run("Success case (user deleting themself)", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleUser},
			Password:        "password",
			PasswordConfirm: "password",
		}

		// Creating a User to be deleted...
		u, err := us.Create(ctx, traceID, nur, now)
		if err != nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
		}

		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   u.ID,
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleUser},
		}

		// Deleting User...
		if err = us.Delete(ctx, traceID, claims, u.ID); err != nil {
			tt.Fatalf("\t%s\tDelete() err = %v, want %v", tests.Failed, err, nil)
		}

		// Trying to retrieve the deleted User...
		_, err = us.GetByID(ctx, traceID, claims, u.ID)
		if err != service.ErrNotFound {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, service.ErrNotFound)
		}
	})

	t.Run("Success case (admin deleting user)", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}

		// Creating a User to be deleted...
		u, err := us.Create(ctx, traceID, nur, now)
		if err != nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
		}

		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		// Deleting User using Claims with Admin role...
		if err = us.Delete(ctx, traceID, claims, u.ID); err != nil {
			tt.Fatalf("\t%s\tDelete() err = %v, want %v", tests.Failed, err, nil)
		}

		// Trying to retrieve the deleted User...
		_, err = us.GetByID(ctx, traceID, claims, u.ID)
		if err != service.ErrNotFound {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, service.ErrNotFound)
		}
	})

	t.Run("Not authorized case", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}

		// Creating test User...
		u, err := us.Create(ctx, traceID, nur, now)
		if err != nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
		}

		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleUser},
		}

		// Attempting to delete User with invalid claims...
		if err = us.Delete(ctx, traceID, claims, u.ID); err != service.ErrForbidden {
			tt.Fatalf("\t%s\tDelete() err = %v, want %v", tests.Failed, err, service.ErrForbidden)
		}

		// Trying to retrieve the User with valid credentials...
		claims.StandardClaims.Subject = u.ID

		_, err = us.GetByID(ctx, traceID, claims, u.ID)
		if err != nil {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, nil)
		}
	})

	t.Run("Invalid ID case", func(tt *testing.T) {
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		// Attempting to delete User with invalid ID...
		if err := us.Delete(ctx, traceID, claims, "invalidID"); err != service.ErrInvalidID {
			tt.Fatalf("\t%s\tDelete() err = %v, want %v", tests.Failed, err, service.ErrInvalidID)
		}
	})
}

func TestGetByID(t *testing.T) {
	_, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	if err := schema.Seed(db); err != nil {
		t.Fatalf("\tschema.Seed() err = %v", err)
	}

	ur, _ := repository.NewRepository(db)
	us, _ := service.NewBasicService(ur)
	now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	traceID := "00000000-0000-0000-0000-000000000000"

	nur := service.NewUserRequest{
		Name:            "Santiago",
		Email:           "santiago@santiago.com",
		LastName:        "Hernández",
		Country:         "Argentina",
		Roles:           []string{auth.RoleAdmin},
		Password:        "password",
		PasswordConfirm: "password",
	}

	// Creating User to be retrieved...
	u, err := us.Create(ctx, traceID, nur, now)
	if err != nil {
		t.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
	}

	t.Run("Success case (same user ID and Claims Subject)", func(tt *testing.T) {
		// Creating Claims with the user ID as the Subject...
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   u.ID,
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleUser},
		}

		// Retrieving User...
		saved, err := us.GetByID(ctx, traceID, claims, u.ID)
		if err != nil {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, nil)
		}
		if diff := cmp.Diff(u, saved); diff != "" {
			t.Fatalf("\t%s\tGetByID() user = %v, diff:\n%s", tests.Failed, saved, diff)
		}
	})

	t.Run("Success case (admin requesting User)", func(tt *testing.T) {
		// Creating Claims with Admin role...
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		// Retrieving User...
		saved, err := us.GetByID(ctx, traceID, claims, u.ID)
		if err != nil {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, nil)
		}
		if diff := cmp.Diff(u, saved); diff != "" {
			t.Fatalf("\t%s\tGetByID() user = %v, diff:\n%s", tests.Failed, saved, diff)
		}
	})

	t.Run("Forbidden", func(tt *testing.T) {
		// Creating Claims with User as the role...
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleUser},
		}

		// Retrieving User...
		if _, err := us.GetByID(ctx, traceID, claims, u.ID); err != service.ErrForbidden {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, service.ErrForbidden)
		}
	})

	t.Run("User not found", func(tt *testing.T) {
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		// Retrieving User...
		if _, err := us.GetByID(ctx, traceID, claims, uuid.New().String()); err != service.ErrNotFound {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, service.ErrNotFound)
		}
	})

	t.Run("Invalid ID", func(tt *testing.T) {
		claims := auth.Claims{
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   uuid.New().String(),
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
			Roles: []string{auth.RoleAdmin},
		}

		// Retrieving User...
		if _, err := us.GetByID(ctx, traceID, claims, "invalid ID"); err != service.ErrInvalidID {
			tt.Fatalf("\t%s\tGetByID() err = %v, want %v", tests.Failed, err, service.ErrInvalidID)
		}
	})
}

func TestAuthenticate(t *testing.T) {
	_, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	ctx := context.Background()
	now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
	traceID := "00000000-0000-0000-0000-000000000000"
	ur, _ := repository.NewRepository(db)
	us, _ := service.NewBasicService(ur)

	t.Run("Success case", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			LastName:        "Hernández",
			Email:           "santiago@santiago.com",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}

		u, err := us.Create(ctx, traceID, nur, now)
		if err != nil {
			tt.Fatalf("\t%s\tCreate() err = %v, want %v", tests.Failed, err, nil)
		}

		claims, err := us.Authenticate(ctx, traceID, now, "santiago@santiago.com", "password")
		if err != nil {
			tt.Fatalf("\t%s\tAuthenticate() err = %v, want %v", tests.Failed, err, nil)
		}

		// Compare Claims returned by Authenticate() with expected Claims...
		want := auth.Claims{
			Roles: u.Roles,
			StandardClaims: jwt.StandardClaims{
				Issuer:    "service template",
				Subject:   u.ID,
				Audience:  "clients",
				ExpiresAt: now.Add(time.Hour).Unix(),
				IssuedAt:  now.Unix(),
			},
		}
		if diff := cmp.Diff(want, claims); diff != "" {
			t.Fatalf("\t%s\tAuthenticate() claims = %v, diff:\n%s", tests.Failed, claims, diff)
		}
	})

	t.Run("User not found", func(tt *testing.T) {
		_, err := us.Authenticate(ctx, traceID, now, "some@email.com", "some_password")
		if err != service.ErrAuthenticationFailure {
			tt.Fatalf("\t%s\tAuthenticate() err = %v, want %v", tests.Failed, err, service.ErrAuthenticationFailure)
		}
	})
}
