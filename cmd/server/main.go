package main

import (
	"context"
	"os"

	"github.com/jackc/pgx"
	"github.com/webdevelop-pro/migration-service/internal/cli"
	"github.com/webdevelop-pro/migration-service/internal/config"
	"github.com/webdevelop-pro/migration-service/internal/http"
	"github.com/webdevelop-pro/migration-service/internal/logger"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		// Provide default logger for fx
		fx.Logger(logger.NewLogger("fx", os.Stdout, config.GetConfig())),

		// Provide dependencies
		fx.Provide(
			// Configuration
			config.GetConfig,
			// Postgres
			cli.ConnnectToDB,
			// Migration
			migration.NewSet,
		),
		fx.Invoke(
			// Start migration
			func(lc fx.Lifecycle, shutdowner fx.Shutdowner, cfg *config.Config, pg *pgx.ConnPool, mSet *migration.Set) {
				l := logger.NewLogger("main", os.Stdout, cfg)
				// read migration directory and initializate data
				err := migration.ReadDir(cfg.MigrationDir, mSet)
				if err != nil {
					l.Fatal().Err(err).Msg("failed to read directory with migrations")
					return
				}
				lc.Append(fx.Hook{
					OnStart: func(_ context.Context) error {
						argsWithoutProg := os.Args[1:]
						if len(argsWithoutProg) == 0 || argsWithoutProg[0] == "-cli" {
							cli.StartApp(cfg, pg, mSet)
							if err := shutdowner.Shutdown(); err != nil {
								l.Fatal().Err(err).Msg("cannot stop application")
							}
						} else if argsWithoutProg[0] == "-http-server" {
							l.Debug().Msgf("start http app %s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
							go http.StartApp(cfg, pg, mSet)
						} else {
							l.Fatal().Msgf(`
	wrong argument %s
	use no arguments or -cli to execute migrations
	or -http-server to run as http-server
	`, argsWithoutProg[0])
						}
						return nil
					},
					OnStop: func(_ context.Context) error {
						l.Debug().Msg("end app")
						return nil
					},
				},
				)
			},
		),
	).Run()
}
