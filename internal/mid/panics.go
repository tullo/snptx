package mid

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/platform/web"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics(log *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// If the context is missing this value,
			// request the service to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if r := recover(); r != nil {
					err = errors.Errorf("panic: %v", r)

					// Log the Go stack trace for this panic's goroutine.
					log.Printf("%s :\n%s", v.TraceID, debug.Stack())
				}
			}()

			// Call the next Handler and set its return value in the err variable.
			return after(ctx, w, r)
		}

		return h
	}

	return f
}

/*
// recoverPanic recovers the panic and logs the cause
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// called last on the way up in the middleware chain while Go unwinds the stack
		defer func() {
			// check if there has been a panic or not
			if err := recover(); err != nil {
				// Trigger the Go server to automatically close the current connection
				// after a response has been sent.
				w.Header().Set("Connection", "close")
				// format error with default textual representation
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
*/
