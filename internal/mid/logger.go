package mid

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/santiagoh1997/service-template/internal/pkg/web"
	"go.opentelemetry.io/otel/trace"
)

// Logger writes some information about the request to the logs in the
// format: TraceID : (200) GET /foo -> IP ADDR (latency)
func Logger(log *log.Logger) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.mid.logger")
			defer span.End()

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			log.Printf("%s: started: %s %s -> %s",
				v.TraceID,
				r.Method, r.URL.Path, r.RemoteAddr,
			)

			err := handler(ctx, w, r)

			log.Printf("%s: completed: %s %s -> %s (%d) (%s)",
				v.TraceID,
				r.Method, r.URL.Path, r.RemoteAddr,
				v.StatusCode, time.Since(v.Now),
			)

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
