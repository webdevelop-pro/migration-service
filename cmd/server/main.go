package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/app"
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
			migration.GetConfig,
			migration.NewSet,
		),

		fx.Invoke(
			// Run HTTP server
			registerHooks,
		),
	).Run()
}

func registerHooks(
	lifecycle fx.Lifecycle, log logger.Logger, dbConfig *db.Config,
	dbPool *pgxpool.Pool, migrationCfg *migration.Config, set *migration.Set,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				err := migration.ReadDir(migrationCfg.Dir, set)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to read directory with migrations")
					return err
				}
				log.Info().Msgf("started on: %s", time.Now().String())

				n, err := set.ApplyAll(migrationCfg.ForceApply)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to apply all migrations")
					return err
				}
				log.Info().Int("n", n).Msg("applied migrations")

				if migrationCfg.ApplyOnly {
					// clean up here
					dbPool.Close()
					os.Exit(3)
					return nil
				}

				appCfg := app.GetConfig()
				myApp := app.New(
					logger.NewDefaultComponent("app"),
					appCfg,
					set,
				)
				myApp.StartServer()
				return nil
			},
			OnStop: func(context.Context) error {
				dbPool.Close()
				return nil
			},
		},
	)
}
