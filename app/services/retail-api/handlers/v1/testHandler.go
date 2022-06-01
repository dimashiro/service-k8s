package v1_test

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/dimashiro/service/business/validate"
	"github.com/dimashiro/service/foundation/webapp"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		return validate.NewRequestError(errors.New("request error"), http.StatusBadRequest)
	}
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return webapp.Respond(ctx, w, status, http.StatusOK)
}
