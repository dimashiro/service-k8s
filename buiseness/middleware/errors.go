package middleware

import (
	"context"
	"net/http"

	"github.com/dimashiro/service/buiseness/validate"
	"github.com/dimashiro/service/foundation/webapp"
	"go.uber.org/zap"
)

func Errors(log *zap.SugaredLogger) webapp.Middleware {

	m := func(handler webapp.Handler) webapp.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v, err := webapp.GetValues(ctx)
			if err != nil {
				return webapp.NewShutdownError("webapp values missing from context")
			}

			if err := handler(ctx, w, r); err != nil {
				// Log the error.
				log.Errorw("ERROR", "traceid", v.TraceID, "ERROR", err)

				// build error response
				var er validate.ErrorResponse
				var status int
				switch {
				case validate.IsFieldErrors(err):
					er = validate.ErrorResponse{
						Error:  "data validation error",
						Fields: err.Error(),
					}
					status = http.StatusBadRequest
				case validate.IsRequestError(err):
					reqErr := validate.GetRequestError(err)
					er = validate.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = validate.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// Respond back to the client.
				if err := webapp.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// check for shutdown
				if ok := webapp.IsShutdown(err); ok {
					return err
				}

			}
			return nil
		}
		return h
	}
	return m
}
