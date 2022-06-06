package user

import (
	"context"
	"fmt"
	"time"

	"github.com/dimashiro/service/business/auth"
	"github.com/dimashiro/service/business/data/store/user"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Core struct {
	log  *zap.SugaredLogger
	user user.Store
}

func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		log:  log,
		user: user.NewStore(log, db),
	}
}

func (c Core) Create(ctx context.Context, nu user.NewUserDTO, now time.Time) (user.User, error) {

	usr, err := c.user.Create(ctx, nu, now)
	if err != nil {
		return user.User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

func (c Core) Update(ctx context.Context, claims auth.Claims, userID string, uu user.UpdateUserDTO, now time.Time) error {

	if err := c.user.Update(ctx, claims, userID, uu, now); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

func (c Core) Delete(ctx context.Context, claims auth.Claims, userID string) error {

	if err := c.user.Delete(ctx, claims, userID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (c Core) GetAll(ctx context.Context, pageNumber int, rowsPerPage int) ([]user.User, error) {

	users, err := c.user.GetAll(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}

	return users, nil
}

func (c Core) GetByID(ctx context.Context, claims auth.Claims, userID string) (user.User, error) {

	usr, err := c.user.GetByID(ctx, claims, userID)
	if err != nil {
		return user.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return usr, nil
}

func (c Core) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {

	claims, err := c.user.Authenticate(ctx, now, email, password)
	if err != nil {
		return auth.Claims{}, fmt.Errorf("authenticate: %w", err)
	}

	return claims, nil
}
