package main

import (
	"context"
	"os"

	"github.com/jackc/pgx"
	"github.com/webdevelop-pro/migration-service/internal/cli"
	"github.com/webdevelop-pro/migration-service/internal/config"
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
			// Start syncing
			func(lc fx.Lifecycle, shutdowner fx.Shutdowner, cfg *config.Config, pg *pgx.ConnPool, mSet *migration.Set) {
				l := logger.NewLogger("cliApp", os.Stdout, cfg)
				lc.Append(fx.Hook{
					OnStart: func(_ context.Context) error {
						l.Debug().Msg("start app")
						cli.StartApp(cfg, pg, mSet)
						shutdowner.Shutdown()
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

/*
func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	cfg, err := config.NewConfig()
	if err != nil {
		_ = envconfig.Usage("", &cfg)
		log.Fatal().Err(err).Msg("failed to parse config")
		return
	}

	var pg *pgx.ConnPool
	i := 0
	log.Info().Interface("cfg", cfg).Msg("connecting to db")
	ticker := time.NewTicker(time.Second)
	for ; ; <-ticker.C {
		i++
		pg, err = pgConnect(cfg)
		if err == nil || i > 60 {
			break
		}
		log.Warn().Err(err).Msg("failed to connect to db")
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
		return
	}
	defer pg.Close()
	ticker.Stop()

	set := migration.NewSet(pg)
	var migrationDir string
	if cfg.MigrationDir != "" {
		migrationDir = cfg.MigrationDir
	} else {
		migrationDir = filepath.Join(binaryPath(), "migrations")
	}
	err = migration.ReadDir(migrationDir, set)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read directory with migrations")
		return
	}

	n, err := set.ApplyAll(cfg.ForceApply)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to apply all migrations")
		return
	}
	log.Info().Int("n", n).Msg("applied migrations")

	if cfg.ApplyOnly {
		return
	}
	svc := api.NewAPI(log.With().Str("module", "api").Logger(), set)
	mux := http.NewServeMux()
	mux.HandleFunc("/apply", svc.HandleApplyMigrations)
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	if err = http.ListenAndServe(cfg.Host+":"+cfg.Port, mux); err != nil {
		log.Fatal().Err(err).Msg("failed to start REST API listener")
	}
}
*/
