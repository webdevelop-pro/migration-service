package app

import (
	"net/http"

	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/api"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

type Config struct {
	Host string `required:"true"`
	Port string `required:"true"`
}

const PkgName = "app"

type App struct {
	log logger.Logger
	set *migration.Set
	cfg *Config
}

func New(log logger.Logger, cfg *Config, set *migration.Set) *App {
	return &App{
		log: log,
		set: set,
		cfg: cfg,
	}
}

// GetConfig return config from envs
func GetConfig() *Config {
	cfg := &Config{}

	if err := configurator.NewConfiguration(cfg, PkgName); err != nil {
		log := logger.NewDefaultComponent(PkgName)
		log.Fatal().Err(err).Msgf("failed to get configuration of %s", PkgName)
	}

	return cfg
}

func (app *App) StartServer() {
	svc := api.NewAPI(app.log.With().Str("module", "api").Logger(), app.set)
	mux := http.NewServeMux()
	mux.HandleFunc("/apply", svc.HandleApplyMigrations)
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(app.cfg.Host+":"+app.cfg.Port, mux); err != nil {
		app.log.Fatal().Err(err).Msg("failed to start REST API listener")
	}
}
