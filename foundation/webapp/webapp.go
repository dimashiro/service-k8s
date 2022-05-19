// Package web contains a small web framework extension.
package webapp

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

// our own Handler
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
	}
}

// SignalShutdown is used to shutdown the app
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler Handler) {

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := handler(ctx, w, r); err != nil {
			return
		}

	}

	fullPath := path
	if group != "" {
		fullPath = "/" + group + path
	}
	a.ContextMux.Handle(method, fullPath, h)
}
