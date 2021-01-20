package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/santiagoh1997/service-template/internal/business/auth"
)

type instrumentingUserService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	Service        UserService
}

// NewInstrumentingUserService returns an instance of an instrumenting UserService.
func NewInstrumentingUserService(requestCount metrics.Counter, requestLatency metrics.Histogram, us UserService) UserService {
	return &instrumentingUserService{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		Service:        us,
	}
}

func (s *instrumentingUserService) Create(ctx context.Context, traceID string, nur NewUserRequest, now time.Time) (user User, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "create").Add(1)
		s.requestLatency.With("method", "create", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.Create(ctx, traceID, nur, now)
}

func (s *instrumentingUserService) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uur UpdateUserRequest, now time.Time) (user User, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "update").Add(1)
		s.requestLatency.With("method", "update", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.Update(ctx, traceID, claims, userID, uur, now)
}

func (s *instrumentingUserService) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "delete").Add(1)
		s.requestLatency.With("method", "delete", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Delete(ctx, traceID, claims, userID)
}

func (s *instrumentingUserService) GetAll(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) (users []User, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "get_all").Add(1)
		s.requestLatency.With("method", "get_all", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.GetAll(ctx, traceID, pageNumber, rowsPerPage)
}

func (s *instrumentingUserService) GetByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (user User, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "get_by_id").Add(1)
		s.requestLatency.With("method", "get_by_id", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.GetByID(ctx, traceID, claims, userID)
}

func (s *instrumentingUserService) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (claims auth.Claims, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "authenticate").Add(1)
		s.requestLatency.With("method", "authenticate", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.Service.Authenticate(ctx, traceID, now, email, password)
}
