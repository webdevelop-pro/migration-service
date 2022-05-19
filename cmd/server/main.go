package main

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/api"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
	"go.uber.org/fx"
)

// @schemes https
func main() {
	fx.New(
		fx.Logger(logger.NewDefaultComponent("fx")),
		fx.Provide(
			// Default logger
			logger.NewDefault,
			// Database connection
			db.GetConfig,
			db.NewPool,
			// Repository
			migration.GetConfig(),
		),
		// Init http handlers
		// http.InitServer(),

		fx.Invoke(
			// Run HTTP server
			registerHooks,
		),
	).Run()
}

func registerHooks(
	lifecycle fx.Lifecycle, log logger.Logger, dbConfig *db.Config,
	dbPool *pgxpool.Pool, migrationCfg *migration.Config,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				set := migration.NewSet(dbPool)
				var migrationDir string
				if migrationCfg.MigrationDir != "" {
					migrationDir = migrationCfg.MigrationDir
				} else {
					migrationDir = filepath.Join(binaryPath(), "migrations")
				}
				err := migration.ReadDir(migrationDir, set)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to read directory with migrations")
					return err
				}

				n, err := set.ApplyAll(migrationCfg.ForceApply)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to apply all migrations")
					return err
				}
				log.Info().Int("n", n).Msg("applied migrations")

				if migrationCfg.ApplyOnly {
					return nil
				}
				svc := api.NewAPI(log.With().Str("module", "api").Logger(), set)
				mux := http.NewServeMux()
				mux.HandleFunc("/apply", svc.HandleApplyMigrations)
				mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
				if err = http.ListenAndServe(migrationCfg.Host+":"+migrationCfg.Port, mux); err != nil {
					log.Fatal().Err(err).Msg("failed to start REST API listener")
				}
				log.Info().Msgf("started on: %s", time.Now().String())
				return nil
			},
			OnStop: func(context.Context) error {
				return nil
			},
		},
	)
}

func binaryPath() string {
	ex, err := os.Executable()
	if err == nil && !strings.HasPrefix(ex, "/var/folders/") && !strings.HasPrefix(ex, "/tmp/go-build") {
		return path.Dir(ex)
	}
	_, callerFile, _, _ := runtime.Caller(1)
	ex = filepath.Dir(callerFile)
	return ex
}
