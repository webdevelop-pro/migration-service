package main

import (
	"context"

	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/go-common/server"
	"github.com/webdevelop-pro/migration-service/internal/adapters"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
	"github.com/webdevelop-pro/migration-service/internal/ports/http"
	"github.com/webdevelop-pro/migration-service/internal/services"
	"go.uber.org/fx"
)

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
			// Http Server
			server.New,
		),

		fx.Invoke(
			http.InitHandlers,
			// Run HTTP server
			RunHttpServer,
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

func RunHttpServer(c *configurator.Configurator, lc fx.Lifecycle, srv *server.HttpServer) {
	cfg := c.New("main", &Config{}, "main").(*Config)

	if !cfg.ApplyOnly {
		return
	}

	server.StartServer(lc, srv)
}

func RunMigrations(sd fx.Shutdowner, app *app.App, c *configurator.Configurator) {
	cfg := c.New("main", &Config{}, "main").(*Config)

	app.ApplyAll()

	if cfg.ApplyOnly {
		sd.Shutdown()
	}
}
