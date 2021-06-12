package cli

import (
	"os"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/webdevelop-pro/migration-service/internal/config"
	"github.com/webdevelop-pro/migration-service/internal/logger"
	"github.com/webdevelop-pro/migration-service/pkg/migration"
)

func pgConnect(cfg *config.Config) (*pgx.ConnPool, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var errCnt int
	for ; ; <-ticker.C {
		pgConfig := new(pgx.ConnConfig)
		pgConfig.TLSConfig = nil
		connPoolConfig := pgx.ConnPoolConfig{
			ConnConfig: pgx.ConnConfig{
				Host:     cfg.Database.Host,
				Port:     cfg.Database.Port,
				User:     cfg.Database.User,
				Password: cfg.Database.Password,
				Database: cfg.Database.Database,
			},
			AcquireTimeout: 10 * time.Second,
			MaxConnections: cfg.Database.MaxConnections,
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

func ConnnectToDB(cfg *config.Config) *pgx.ConnPool {
	l := logger.NewLogger("cliApp", os.Stdout, cfg)
	l.Info().Interface("cfg", cfg).Msg("connecting to db")
	var pg *pgx.ConnPool
	var err error
	i := 0

	ticker := time.NewTicker(time.Second)
	for ; ; <-ticker.C {
		i++
		pg, err = pgConnect(cfg)
		if err == nil || i > 60 {
			break
		}
		l.Warn().Err(err).Msg("failed to connect to db")
	}
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to db")
	}
	ticker.Stop()
	return pg
}

// StartApp is function that registers start of http server in lifecycle
func StartApp(cfg *config.Config, pg *pgx.ConnPool, mSet *migration.Set) {
	l := logger.NewLogger("cliApp", os.Stdout, cfg)

	err := migration.ReadDir(cfg.MigrationDir, mSet)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to read directory with migrations")
		return
	}

	n, err := mSet.ApplyAll(cfg.ForceApply)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to apply all migrations")
		return
	}
	l.Info().Int("n", n).Msg("applied migrations")
	pg.Close()
}
