package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/business/auth"
	"github.com/santiagoh1997/service-template/internal/business/service"
	"github.com/santiagoh1997/service-template/internal/foundation/web"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrWebValuesMissing is returned whenever the web.Values are missing
	// from the context passed to a handler.
	// This issues a service shutdown as it represents a serious integrity problem.
	ErrWebValuesMissing = web.NewShutdownError("web value missing from context")
)

type userHandler struct {
	svc  service.UserService
	auth *auth.Auth
}

func (uh userHandler) getAll(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.userHandler.getAll")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return ErrWebValuesMissing
	}

	params := web.Params(r)
	pageNumber, err := strconv.Atoi(params["page"])
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid page format: %s", params["page"]), http.StatusBadRequest)
	}
	rowsPerPage, err := strconv.Atoi(params["rows"])
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid rows format: %s", params["rows"]), http.StatusBadRequest)
	}

	users, err := uh.svc.GetAll(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query for users")
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

func (uh userHandler) getByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.userHandler.getByID")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return ErrWebValuesMissing
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	params := web.Params(r)
	usr, err := uh.svc.GetByID(ctx, v.TraceID, claims, params["id"])
	if err != nil {
		switch err {
		case service.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case service.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case service.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (uh userHandler) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.userHandler.create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return ErrWebValuesMissing
	}

	var nur service.NewUserRequest
	if err := web.Decode(r, &nur); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	usr, err := uh.svc.Create(ctx, v.TraceID, nur, v.Now)
	if err != nil {
		switch err {
		case service.ErrDuplicatedEmail:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrap(err, "creating user")
		}
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

func (uh userHandler) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.userHandler.update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return ErrWebValuesMissing
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	var uur service.UpdateUserRequest
	if err := web.Decode(r, &uur); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	params := web.Params(r)
	u, err := uh.svc.Update(ctx, v.TraceID, claims, params["id"], uur, v.Now)
	if err != nil {
		switch err {
		case service.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case service.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case service.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s  User: %+v", params["id"], &uur)
		}
	}

	return web.Respond(ctx, w, u, http.StatusOK)
}

func (uh userHandler) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.userHandler.delete")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return ErrWebValuesMissing
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	params := web.Params(r)
	err := uh.svc.Delete(ctx, v.TraceID, claims, params["id"])
	if err != nil {
		switch err {
		case service.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case service.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (uh userHandler) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.userHandler.token")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return ErrWebValuesMissing
	}

	var lr service.LoginRequest
	if err := web.Decode(r, &lr); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	claims, err := uh.svc.Authenticate(ctx, v.TraceID, v.Now, lr.Email, lr.Password)
	if err != nil {
		switch err {
		case service.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	params := web.Params(r)

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = uh.auth.GenerateToken(params["kid"], claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
