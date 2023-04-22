package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
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
		),

		fx.Invoke(
			// Run application
			RunApp,
		),
	)

	if err := a.Start(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed")
	}

	a.Done()

	log.Info().Msg("done")
}

func RunApp(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator) {
	init := flag.Bool("init", false, "initialize service by creating migration table at DB")
	finalSql := flag.String("final-sql", "", "if provided - program return final SQL for migrations without applying it. Argument = service name")
	flag.Parse()

	if *init {
		RunInit(sd, _app)
		return
	}
	if *finalSql != "" {
		GetFinalSQL(sd, _app, c, *finalSql)
		return
	}

	RunMigrations(sd, _app, c)
}

func RunMigrations(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator) {
	cfg := c.New("migration", &app.Config{}, "migration").(*app.Config)
	if err := _app.ApplyAll(cfg.Dir); err != nil {
		log := logger.NewDefault()
		log.Error().Err(err).Msg("error during migrations")
	}

	sd.Shutdown()
}

func GetFinalSQL(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator, serviceName string) {
	cfg := c.New("migration", &app.Config{}, "migration").(*app.Config)
	sql, err := _app.GetSQL(context.Background(), cfg.Dir, serviceName)
	if err != nil {
		log := logger.NewDefault()
		log.Error().Err(err).Msg("error during forming sql for migration")
	}
	fmt.Println(sql)
	sd.Shutdown()
}

func RunInit(sd fx.Shutdowner, _app *app.App) {
	err := _app.Init(context.Background())
	log := logger.NewDefault()
	if err != nil {
		log.Error().Err(err).Msg("error during creating migration table")
	}
	log.Info().Msg("successfully initialized")
	sd.Shutdown()
}
