package main

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

type config struct {
	Host             string `default:""`
	Port             string `default:"8085"`
	DbDatabase       string `required:"true" split_words:"true"`
	DbHost           string `required:"true" split_words:"true"`
	DbPort           uint16 `default:"5432" split_words:"true"`
	DbUser           string `required:"true" split_words:"true"`
	DbPassword       string `split_words:"true"`
	DbMaxConnections int    `default:"5" split_words:"true"`
	MigrationDir     string `split_words:"true"`
	ForceApply       bool   `split_words:"true"`
	ApplyOnly        bool   `split_words:"true"`
}

func pgConnect(cfg config) (*pgx.ConnPool, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var errCnt int
	for ; ; <-ticker.C {
		pgConfig := new(pgx.ConnConfig)
		pgConfig.TLSConfig = nil
		connPoolConfig := pgx.ConnPoolConfig{
			ConnConfig: pgx.ConnConfig{
				Host:     cfg.DbHost,
				Port:     cfg.DbPort,
				User:     cfg.DbUser,
				Password: cfg.DbPassword,
				Database: cfg.DbDatabase,
			},
			AcquireTimeout: 10 * time.Second,
			MaxConnections: cfg.DbMaxConnections,
		}
		pg, err := pgx.NewConnPool(connPoolConfig)
		if err != nil {
			if errCnt > 60 {
				return nil, errors.Wrap(err, "failed to connect to db")
			}
			errCnt++
			continue
		}
		return pg, nil
	}
}

func initConfig() (cfg config, pg *pgx.ConnPool, err error) {
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

func main() {
	cfg, pg, err := initConfig()
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize"))
	}

	defer pg.Close()
	set := migration.NewSet(pg)
	err = migration.ReadDir(cfg.MigrationDir, set)
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
