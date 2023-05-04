package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/webdevelop-pro/lib/configurator"
	"github.com/webdevelop-pro/lib/logger"
	"github.com/webdevelop-pro/lib/server"
	"github.com/webdevelop-pro/migration-service/internal/adapters"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
	"github.com/webdevelop-pro/migration-service/internal/ports"
	"github.com/webdevelop-pro/migration-service/internal/services"
	"go.uber.org/fx"
)

// @schemes https
func main() {
	log := logger.NewComponentLogger("fx", nil)

	fx.New(
		fx.Logger(log),
		fx.Provide(
			// Configurator
			configurator.NewConfigurator,
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
			// InitHandlers
			ports.InitHandlers,
			// Run application
			RunApp,
		),
	).Run()

	/*
		if err := a.Start(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("failed")
		}

		a.Done()
	*/

	log.Info().Msg("done")
}

func RunHttpServer(lc fx.Lifecycle, srv *server.HttpServer) {
	server.StartServer(lc, srv)
}

func RunApp(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator, lc fx.Lifecycle, srv *server.HttpServer) {
	init := flag.Bool("init", false, "initialize service by creating migration table at DB")
	finalSql := flag.String("final-sql", "", "if provided - program return final SQL for migrations without applying it. Argument = service name")
	force := flag.Bool("force", false, "force apply migration without version checking. Accept files or dir paths. Will not update service version if applied version is lower, then already applied")
	skip := flag.Bool("fake", false, "fake do not apply any migration but mark according migrations in migration_services table as completed")
	check := flag.Bool("check", false, "check verifies if all hashes of migrations are equal to those in migration table. If no - returns list of files with migrations, that have differences. Can accept files or dirs of migrations as arguments")
	checkApply := flag.Bool("check-apply", false, "check-apply compares hashes of all migrations with hashes in DB and try to apply those, that have differences. Can accept files or dirs of migrations as arguments")
	applyOnly := flag.Bool("apply-only", false, "apply and shutdown migration service, do not start web service")

	flag.Parse()

	if *init {
		RunInit(sd, _app)
		return
	}
	if *force {
		args := flag.Args()
		RunForceApply(sd, _app, args)
		return
	}
	if *skip {
		args := flag.Args()
		RunFakeApply(sd, _app, args)
		return
	}
	if *check {
		args := flag.Args()
		RunCheck(sd, _app, args, c)
		return
	}
	if *checkApply {
		args := flag.Args()
		RunCheckApply(sd, _app, args, c)
		return
	}
	if *finalSql != "" {
		GetFinalSQL(sd, _app, c, *finalSql)
		return
	}

	RunMigrations(sd, _app, c)

	if *applyOnly == false {
		// Run server
		RunHttpServer(lc, srv)
	} else {
		sd.Shutdown()
	}
	// Run server
	RunHttpServer(lc, srv)
}

func RunMigrations(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator) {
	cfg := c.New("migration", &app.Config{}, "migration").(*app.Config)
	if err := _app.ApplyAll(cfg.Dir); err != nil {
		log := logger.NewComponentLogger("RunMigrations", nil)
		log.Error().Err(err).Msg("error during migrations")
	}

	// sd.Shutdown()
}

func GetFinalSQL(sd fx.Shutdowner, _app *app.App, c *configurator.Configurator, serviceName string) {
	cfg := c.New("migration", &app.Config{}, "migration").(*app.Config)
	sql, err := _app.GetSQL(context.Background(), cfg.Dir, serviceName)
	if err != nil {
		log := logger.NewComponentLogger("GetFinalSQL", nil)
		log.Error().Err(err).Msg("error during forming sql for migration")
	}
	fmt.Println(sql)
	sd.Shutdown()
}

func RunInit(sd fx.Shutdowner, _app *app.App) {
	err := _app.Init(context.Background())
	log := logger.NewComponentLogger("RunInit", nil)
	if err != nil {
		log.Error().Err(err).Msg("error during creating migration table")
	}
	log.Info().Msg("successfully initialized")
	sd.Shutdown()
}

func RunForceApply(sd fx.Shutdowner, _app *app.App, args []string) {
	err := _app.ForceApply(args)
	log := logger.NewComponentLogger("RunForceApply", nil)
	if err != nil {
		log.Error().Err(err).Msg("error during force apply migrations")
	}
	log.Info().Msg("successfully force applied")
	sd.Shutdown()
}

func RunFakeApply(sd fx.Shutdowner, _app *app.App, args []string) {
	err := _app.FakeApply(args)
	log := logger.NewComponentLogger("RunSkipApply", nil)
	if err != nil {
		log.Error().Err(err).Msg("error during skip migrations")
	}
	log.Info().Msg("successfully skipped and marked as finished")
	sd.Shutdown()
}

func RunCheck(sd fx.Shutdowner, _app *app.App, args []string, c *configurator.Configurator) {
	cfg := c.New("migration", &app.Config{}, "migration").(*app.Config)
	if len(args) == 0 {
		args = append(args, cfg.Dir)
	}
	_, _, err := _app.CheckMigrationHash(args)
	log := logger.NewComponentLogger("RunCheck", nil)
	if err != nil {
		log.Error().Err(err).Msg("error during checking migrations")
	}
	sd.Shutdown()
}

func RunCheckApply(sd fx.Shutdowner, _app *app.App, args []string, c *configurator.Configurator) {
	cfg := c.New("migration", &app.Config{}, "migration").(*app.Config)
	if len(args) == 0 {
		args = append(args, cfg.Dir)
	}
	err := _app.CheckAndApplyMigrations(args)
	log := logger.NewComponentLogger("RunCheckApply", nil)
	if err != nil {
		log.Error().Err(err).Msg("error during checking and applying migrations")
	}

	sd.Shutdown()
}
