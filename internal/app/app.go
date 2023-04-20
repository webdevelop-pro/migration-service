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

const pkgName = "migration"

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
		set:  migration.New(repo),
	}
}

func (a *App) ApplyAll(dir string) error {
	if dir == "" {
		dir = a.cfg.Dir
	}
	a.set.ClearData()
	err := migration.ReadDir(dir, "", a.set)
	if err != nil {
		a.log.Error().Err(err).Msgf("can't get migration data from directory: %s", a.cfg.Dir)
		panic(err)
	}
	n, err := a.set.ApplyAll()
	if err != nil {
		a.log.Error().Err(err).Msg("failed to apply all migrations")
		return err
	}
	a.log.Info().Int("n", n).Msg("applied migrations")
	return nil
}

func (a *App) Apply(ctx context.Context, serviceName string) (int, error) {
	if serviceName == "" || !a.set.ServiceExists(serviceName) {
		return 0, fmt.Errorf("service '%s' not found", serviceName)
	}

	ver, err := a.repo.GetServiceVersion(ctx, serviceName)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current service version")
	}

	n, lastVersion, err := a.set.Apply(serviceName, -1, ver)
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
