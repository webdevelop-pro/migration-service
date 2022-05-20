package main

import (
	"testing"

	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

func TestIntegrity(t *testing.T) {
	dbCfg := db.GetConfig()
	pg := db.NewPool(dbCfg)
	defer pg.Close()
	cfg := migration.GetConfig()
	set := migration.NewSet(pg)
	err := migration.ReadDir(cfg.Dir, set)
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
