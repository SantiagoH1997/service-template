package commands

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/business/auth"
	"github.com/santiagoh1997/service-template/business/service"
	"github.com/santiagoh1997/service-template/foundation/database"
)

// UserAdd adds new users into the database.
func UserAdd(traceID string, log *log.Logger, cfg database.Config, email, password string) error {
	if email == "" || password == "" {
		fmt.Println("help: useradd <email> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	us := service.New(log, db)

	nur := service.NewUserRequest{
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}

	usr, err := us.Create(ctx, traceID, nur, time.Now())
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
