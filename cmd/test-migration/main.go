package main

import (
	"fmt"
	migration2 "github.com/webdevelop-pro/migration-service/internal/domain/migration"

	"github.com/pkg/errors"

	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

func main() {
	dbCfg := db.GetConfig()
	pg := db.NewPool(dbCfg)
	defer pg.Close()
	cfg := migration.GetConfig()
	set := migration.NewSet(pg)
	err := migration2.ReadDir(cfg.Dir, set)
	if err != nil {
		panic(errors.Wrap(err, "failed to read directory with migrations"))
	}

	n, err := set.ApplyAll(true)
	if err != nil {
		panic(errors.Wrap(err, "failed to apply all migrations"))
	}
	if n == 0 {
		fmt.Println("no migrations were applied")
	}
}
