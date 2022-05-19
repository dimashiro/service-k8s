package middleware

import (
	"context"
	"net/http"

	"github.com/dimashiro/service/foundation/webapp"
	"go.uber.org/zap"
)

func Logger(log *zap.SugaredLogger) webapp.Middleware {
	m := func(handler webapp.Handler) webapp.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			log.Infow("request started", "method", r.Method, "path", r.URL.Path)
			err := handler(ctx, w, r)
			log.Infow("request ended", "method", r.Method, "path", r.URL.Path)
			return err
		}

		return h
	}
	return m
}
