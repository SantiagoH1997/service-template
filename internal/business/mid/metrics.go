package mid

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/santiagoh1997/service-template/internal/pkg/web"
	"go.opentelemetry.io/otel/trace"
)

// Metrics updates program counters.
func Metrics(errorCount metrics.Counter, redMetrics metrics.Histogram) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.mid.metrics")
			defer span.End()

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			// Don't count anything on /debug routes towards metrics.
			// Call the next handler to continue processing.
			if strings.HasPrefix(r.URL.Path, "/debug") {
				return handler(ctx, w, r)
			}

			// Call the next handler.
			err := handler(ctx, w, r)

			// Increment the request counter.
			s := time.Since(v.Now).Seconds()
			redMetrics.With("method", r.Method, "path", r.URL.Path, "status_code", strconv.Itoa(v.StatusCode)).Observe(s)

			// Increment the errors counter if an error occurred on this request.
			if err != nil {
				errorCount.With("method", r.Method, "path", r.URL.Path).Add(1)
			}

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
