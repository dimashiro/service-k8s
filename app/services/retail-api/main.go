package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/dimashiro/service/app/services/retail-api/handlers"
	"github.com/dimashiro/service/business/auth"
	"github.com/dimashiro/service/business/database"
	"github.com/dimashiro/service/foundation/keystore"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var build = "develop"

func main() {

	log, err := initLogger("RETAIL-API")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("start", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}

}

func run(log *zap.SugaredLogger) error {

	//__________________________________________________________________________
	// GOMAXPROCS
	// set the number of threads available by qoutas
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("start", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	//__________________________________________________________________________
	// Config
	cfg := struct {
		APIHost         string        `env:"APIHOST" env-default:"0.0.0.0:3000"`
		DebugHost       string        `env:"DEBUGHOST" env-default:"0.0.0.0:4000"`
		ReadTimeout     time.Duration `env:"READTIMEOUT" env-default:"5s"`
		WriteTimeout    time.Duration `env:"WRITETIMEOUT" env-default:"10s"`
		IdleTimeout     time.Duration `env:"IDLETIMEOUT" env-default:"120s"`
		ShutdownTimeout time.Duration `env:"SHUTDOWNTIMEOUT" env-default:"20s"`
		AuthKeysFolder  string        `env:"AUTHKEYSFOLDER" env-default:"deploy/keys/"`
		AuthActiveKID   string        `env:"AUTHACTIVEKID" env-default:"developmentkeyid"`
		DBUser          string        `env:"DBUSER" env-default:"postgres"`
		DBPassword      string        `env:"DBPASSWORD" env-default:"postgres,mask"`
		DBHost          string        `env:"DBHOST" env-default:"localhost"`
		DBName          string        `env:"DBNAME" env-default:"postgres"`
		DBMaxIdleConns  int           `env:"DBMAXIDLECONNS" env-default:"0"`
		DBMaxOpenConns  int           `env:"DBMAXOPENCONNS" env-default:"0"`
		DBDisableTLS    bool          `env:"DBDISABLETLS" env-default:"true"`
	}{}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return fmt.Errorf("loading conf: %w", err)
	}

	//__________________________________________________________________________
	// Database

	log.Infow("start", "status", "initializing database support", "host", cfg.DBHost)

	db, err := database.Open(database.Config{
		User:         cfg.DBUser,
		Password:     cfg.DBPassword,
		Host:         cfg.DBHost,
		Name:         cfg.DBName,
		MaxIdleConns: cfg.DBMaxIdleConns,
		MaxOpenConns: cfg.DBMaxOpenConns,
		DisableTLS:   cfg.DBDisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DBHost)
		db.Close()
	}()

	//__________________________________________________________________________
	// Start Debug Service

	log.Infow("start", "status", "debug router started", "host", cfg.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This includes the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log, db)

	// Start the service listening for debug requests.
	go func() {
		if err := http.ListenAndServe(cfg.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.DebugHost, "ERROR", err)
		}
	}()

	//__________________________________________________________________________
	// App start
	log.Infow("start", "version", build)

	//__________________________________________________________________________
	// Initialize authentication support

	log.Infow("start", "status", "initializing authentication support")

	// Construct a key store based on the key files stored in
	// the specified directory.
	ks, err := keystore.NewFS(os.DirFS(cfg.AuthKeysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	auth, err := auth.New(cfg.AuthActiveKID, ks)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	//__________________________________________________________________________
	// Start service
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the mux for the API calls.
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		Auth:     auth,
		DB:       db,
	})

	api := http.Server{
		Addr:         cfg.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Infow("start", "status", "start router", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Blocking main select
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Infow("shutdown", "status", "start shutdown", "signal", sig)
		defer log.Infow("shutdown", "status", "finish shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server: %w", err)
		}
	}

	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}
