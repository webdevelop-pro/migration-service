package app

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
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

func (a *App) ApplyAll() {
	n, err := a.set.ApplyAll(a.cfg.ForceApply)
	if err != nil {
		a.log.Fatal().Err(err).Msg("failed to apply all migrations")
	}

	a.log.Info().Int("n", n).Msg("applied migrations")
}

func (a *App) Apply(ctx context.Context, serviceName string) (int, error) {
	if serviceName == "" || !a.set.ServiceExists(serviceName) {
		return 0, fmt.Errorf("service '%s' not found", serviceName)
	}

	ver, err := a.repo.GetServiceVersion(ctx, serviceName)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current service version")
	}

	n, lastVersion, err := a.set.Apply(serviceName, -1, ver, false, true)
	if err != nil {
		return 0, errors.Wrap(err, "failed to apply migrations")
	}

	if lastVersion > ver {
		if err := a.repo.UpdateServiceVersion(ctx, serviceName, lastVersion); err != nil {
			a.log.Error().Err(err).Msg("failed to bump service version")
		}
	}

	if err != nil {
		return 0, errors.Wrap(err, "failed to encode response")
	}

	return n, nil
}
