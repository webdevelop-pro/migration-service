package app

import (
	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters"
	"github.com/webdevelop-pro/migration-service/internal/domain/migration"
)

const pkgName = "app"

type App struct {
	log  logger.Logger
	repo adapters.Repository
	set  *migration.Set
	cfg  *Config
}

func New(c *configurator.Configurator, repo adapters.Repository) *App {
	cfg := c.New(pkgName, &Config{}, pkgName).(*Config)

	return &App{
		log:  logger.NewDefaultComponent(pkgName),
		repo: repo,
		cfg:  cfg,
		set:  migration.New(cfg.Dir, repo),
	}
}

func (a *App) RunMigration() {
	n, err := a.set.ApplyAll(a.cfg.ForceApply)
	if err != nil {
		a.log.Fatal().Err(err).Msg("failed to apply all migrations")
	}

	a.log.Info().Int("n", n).Msg("applied migrations")
}
