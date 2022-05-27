package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dimashiro/service/buiseness/auth"
	"github.com/dimashiro/service/buiseness/validate"
	"github.com/dimashiro/service/foundation/webapp"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) webapp.Middleware {

	m := func(handler webapp.Handler) webapp.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// Expecting: bearer <token>
			authStr := r.Header.Get("authorization")

			// Parse the authorization header.
			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: bearer <token>")
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role
func Authorize(roles ...string) webapp.Middleware {

	m := func(handler webapp.Handler) webapp.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			claims, err := auth.GetClaims(ctx)
			if err != nil {
				return validate.NewRequestError(
					fmt.Errorf("you are not authorized for that action, no claims"),
					http.StatusForbidden,
				)
			}

			if !claims.Authorized(roles...) {
				return validate.NewRequestError(
					fmt.Errorf("you are not authorized for that action, claims[%v] roles[%v]", claims.Roles, roles),
					http.StatusForbidden,
				)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
