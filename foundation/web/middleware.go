package web

// Middleware is a function that runs some code before and/or after
// another Handler.
type Middleware func(Handler) Handler

// wrapMiddleware creates a new Handler by wrapping middleware around a final
// Handler. The middlewares' Handlers will be executed by requests in the order
// they are provided.
func wrapMiddleware(mw []Middleware, handler Handler) Handler {

	// Looping backwards ensures that the first middleware
	// of the slice is the first to be executed by requests.
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
