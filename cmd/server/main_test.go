package main

import (
	"testing"

	"github.com/jackc/pgx"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/webdevelop-pro/migration-service/internal/config"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

func initConfig() (cfg config.Config, pg *pgx.ConnPool, err error) {
	err = envconfig.Process("", &cfg)
	if err != nil {
		err = errors.Wrap(err, "failed to parse config")
		return
	}
	pg, err = pgConnect(cfg)
	if err != nil {
		err = errors.Wrap(err, "failed to connect to db")
		return
	}
	return
}

func TestIntegrity(t *testing.T) {
	cfg, pg, err := initConfig()
	if err != nil {
		t.Errorf("failed to initialize: %v", err)
		return
	}
	defer pg.Close()
	set := migration.NewSet(pg)
	err = migration.ReadDir(cfg.MigrationDir, set)
	if err != nil {
		t.Errorf("failed to read directory with migrations: %v", err)
		return
	}

	n, err := set.ApplyAll(true)
	if err != nil {
		t.Errorf("failed to apply all migrations: %v", err)
		return
	}
	if n == 0 {
		t.Errorf("no migrations were applied")
	}
}
