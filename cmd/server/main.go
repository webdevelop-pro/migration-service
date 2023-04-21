package main

import (
	"context"

	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
	"github.com/webdevelop-pro/migration-service/internal/services"
	"go.uber.org/fx"
)

const pkgName = "migration"

// @schemes https
func main() {
	log := logger.NewDefault()

	a := fx.New(
		fx.Logger(logger.NewDefaultComponent("fx")),
		fx.Provide(
			// Default logger
			logger.NewDefault,
			// Configurator
			configurator.New,
			// Database connection
			postgres.New,
			// Bind DB with Repository interface
			func(repo *postgres.Repository) adapters.Repository { return repo },
			// app
			app.New,
			// Bind App with service interface
			func(mig *app.App) services.Migration { return mig },
		),

		fx.Invoke(
			// Run migrations
			RunMigrations,
		),
	)

	if err := a.Start(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed")
	}

	a.Done()

	log.Info().Msg("done")
}

func RunMigrations(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator) {
	cfg := c.New(pkgName, &app.Config{}, "migration").(*app.Config)
	if err := _app.ApplyAll(cfg.Dir); err != nil {
		log := logger.NewDefault()
		log.Error().Err(err).Msg("error during migrations")
	}

	sd.Shutdown()
}
