package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/dimashiro/service/foundation/webapp"
	"go.uber.org/zap"
)

func Logger(log *zap.SugaredLogger) webapp.Middleware {
	m := func(handler webapp.Handler) webapp.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := webapp.GetValues(ctx)
			if err != nil {
				return err // web.NewShutdownError("web value missing from context")
			}

			log.Infow("request started", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr)

			err = handler(ctx, w, r)

			log.Infow("request completed", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now))

			return err
		}

		return h
	}
	return m
}
