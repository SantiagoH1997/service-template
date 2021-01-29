package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/pkg/errors"
	"github.com/santiagoh1997/service-template/internal/auth"
)

type instrumentingDecorator struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	Service        UserService
}

// NewInstrumentingDecorator returns a UserService with instrumentation features.
func NewInstrumentingDecorator(requestCount metrics.Counter, requestLatency metrics.Histogram, us UserService) (UserService, error) {
	if requestCount == nil {
		return nil, errors.New("requestCount can't be nil")
	}
	if requestLatency == nil {
		return nil, errors.New("requestLatency can't be nil")
	}
	if us == nil {
		return nil, errors.New("UserService can't be nil")
	}

	return &instrumentingDecorator{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		Service:        us,
	}, nil
}

func (d *instrumentingDecorator) Create(ctx context.Context, traceID string, nur NewUserRequest, now time.Time) (user User, err error) {
	defer func(begin time.Time) {
		d.requestCount.With("method", "create").Add(1)
		d.requestLatency.With("method", "create", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return d.Service.Create(ctx, traceID, nur, now)
}

func (d *instrumentingDecorator) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uur UpdateUserRequest, now time.Time) (err error) {
	defer func(begin time.Time) {
		d.requestCount.With("method", "update").Add(1)
		d.requestLatency.With("method", "update", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return d.Service.Update(ctx, traceID, claims, userID, uur, now)
}

func (d *instrumentingDecorator) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) (err error) {
	defer func(begin time.Time) {
		d.requestCount.With("method", "delete").Add(1)
		d.requestLatency.With("method", "delete", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return d.Service.Delete(ctx, traceID, claims, userID)
}

func (d *instrumentingDecorator) GetByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (user User, err error) {
	defer func(begin time.Time) {
		d.requestCount.With("method", "get_by_id").Add(1)
		d.requestLatency.With("method", "get_by_id", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return d.Service.GetByID(ctx, traceID, claims, userID)
}

func (d *instrumentingDecorator) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (claims auth.Claims, err error) {
	defer func(begin time.Time) {
		d.requestCount.With("method", "authenticate").Add(1)
		d.requestLatency.With("method", "authenticate", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return d.Service.Authenticate(ctx, traceID, now, email, password)
}
