package middleware

import (
	"context"
	"net/http"

	"github.com/dimashiro/service/buiseness/metrics"
	"github.com/dimashiro/service/foundation/webapp"
)

func Metrics() webapp.Middleware {

	m := func(handler webapp.Handler) webapp.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			ctx = metrics.Set(ctx)

			err := handler(ctx, w, r)

			// Increment the request and goroutines counter.
			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}

	return m
}
