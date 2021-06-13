package http

import (
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/webdevelop-pro/migration-service/internal/api"
	"github.com/webdevelop-pro/migration-service/internal/config"
	"github.com/webdevelop-pro/migration-service/internal/logger"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

// StartApp is function that registers start of http server in lifecycle
func StartApp(cfg *config.Config, pg *pgx.ConnPool, mSet *migration.Set) {
	l := logger.NewLogger("http", os.Stdout, cfg)
	defer pg.Close()

	if cfg.HTTP.Host == "" || cfg.HTTP.Port == "" {
		l.Fatal().Msg("please HOST and PORT envs")
	}

	svc := api.NewAPI(l, mSet)
	mux := http.NewServeMux()
	// ToDo
	// Add Authentication key to check request
	mux.HandleFunc("/apply", svc.HandleApplyMigrations)
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(cfg.HTTP.Host+":"+cfg.HTTP.Port, mux); err != nil {
		l.Fatal().Err(err).Msg("failed to start REST API listener")
	}
}
