package mid

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/santiagoh1997/service-template/internal/foundation/web"
	"go.opentelemetry.io/otel/trace"
)

// Metrics updates program counters.
func Metrics(requests, errors metrics.Counter, duration metrics.Histogram) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.mid.metrics")
			defer span.End()

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
			requests.Add(1)
			defer func() {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(v.Now).Seconds())
			}()

			// Increment the errors counter if an error occurred on this request.
			if err != nil {
				errors.Add(1)
			}

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
