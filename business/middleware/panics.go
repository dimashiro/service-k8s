package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/dimashiro/service/business/metrics"
	"github.com/dimashiro/service/foundation/webapp"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics() webapp.Middleware {

	m := func(handler webapp.Handler) webapp.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			defer func() {
				if rec := recover(); rec != nil {

					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

					metrics.AddPanics(ctx)

				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
