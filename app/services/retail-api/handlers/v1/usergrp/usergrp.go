package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dimashiro/service/business/auth"
	"github.com/dimashiro/service/business/core/user"
	userStorage "github.com/dimashiro/service/business/data/store/user"
	"github.com/dimashiro/service/business/database"
	"github.com/dimashiro/service/business/validate"
	"github.com/dimashiro/service/foundation/webapp"
)

// Handlers manages the set of user enpoints.
type Handlers struct {
	User user.Core
	Auth *auth.Auth
}

func (h Handlers) GetAll(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := webapp.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid page format [%s]", page), http.StatusBadRequest)
	}
	rows := webapp.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format [%s]", rows), http.StatusBadRequest)
	}

	users, err := h.User.GetAll(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for users: %w", err)
	}

	return webapp.Respond(ctx, w, users, http.StatusOK)
}

// QueryByID returns a user by its ID.
func (h Handlers) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("no claims in context")
	}

	userID := webapp.Param(r, "id")

	usr, err := h.User.GetByID(ctx, claims, userID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidID):
			return validate.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, database.ErrDBNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return webapp.Respond(ctx, w, usr, http.StatusOK)
}

func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := webapp.GetValues(ctx)
	if err != nil {
		return webapp.NewShutdownError("web value missing from context")
	}

	var nu userStorage.NewUserDTO
	if err := webapp.Decode(r, &nu); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	usr, err := h.User.Create(ctx, nu, v.Now)
	if err != nil {
		return fmt.Errorf("user[%+v]: %w", &usr, err)
	}

	return webapp.Respond(ctx, w, usr, http.StatusCreated)
}

func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := webapp.GetValues(ctx)
	if err != nil {
		return webapp.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("no claims in context")
	}

	var upd userStorage.UpdateUserDTO
	if err := webapp.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	userID := webapp.Param(r, "id")

	if err := h.User.Update(ctx, claims, userID, upd, v.Now); err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidID):
			return validate.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, database.ErrDBNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s] User[%+v]: %w", userID, &upd, err)
		}
	}

	return webapp.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("no claims in context")
	}

	userID := webapp.Param(r, "id")
	if err := h.User.Delete(ctx, claims, userID); err != nil {
		switch {
		case errors.Is(err, database.ErrInvalidID):
			return validate.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, database.ErrDBNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return webapp.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := webapp.GetValues(ctx)
	if err != nil {
		return webapp.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return validate.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := h.User.Authenticate(ctx, v.Now, email, pass)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrDBNotFound):
			return validate.NewRequestError(err, http.StatusNotFound)
		case errors.Is(err, database.ErrAuthenticationFailure):
			return validate.NewRequestError(err, http.StatusUnauthorized)
		default:
			return fmt.Errorf("authenticating: %w", err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return webapp.Respond(ctx, w, tkn, http.StatusOK)
}
