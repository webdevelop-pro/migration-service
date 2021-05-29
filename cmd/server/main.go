package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/webdevelop-pro/migration-service/internal/api"
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

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		_ = envconfig.Usage("", &cfg)
		log.Fatal().Err(err).Msg("failed to parse config")
		return
	}
	var pg *pgx.ConnPool
	i := 0
	log.Info().Interface("cfg", cfg).Msg("connecting to db")
	ticker := time.NewTicker(time.Second)
	for ; ; <-ticker.C {
		i++
		pg, err = pgConnect(cfg)
		if err == nil || i > 60 {
			break
		}
		log.Warn().Err(err).Msg("failed to connect to db")
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
		return
	}
	defer pg.Close()
	ticker.Stop()

	set := migration.NewSet(pg)
	var migrationDir string
	if cfg.MigrationDir != "" {
		migrationDir = cfg.MigrationDir
	} else {
		migrationDir = filepath.Join(binaryPath(), "migrations")
	}
	err = migration.ReadDir(migrationDir, set)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read directory with migrations")
		return
	}

	n, err := set.ApplyAll(cfg.ForceApply)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to apply all migrations")
		return
	}
	log.Info().Int("n", n).Msg("applied migrations")

	if cfg.ApplyOnly {
		return
	}
	svc := api.NewAPI(log.With().Str("module", "api").Logger(), set)
	mux := http.NewServeMux()
	mux.HandleFunc("/apply", svc.HandleApplyMigrations)
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	if err = http.ListenAndServe(cfg.Host+":"+cfg.Port, mux); err != nil {
		log.Fatal().Err(err).Msg("failed to start REST API listener")
	}
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

func binaryPath() string {
	ex, err := os.Executable()
	if err == nil && !strings.HasPrefix(ex, "/var/folders/") && !strings.HasPrefix(ex, "/tmp/go-build") {
		return path.Dir(ex)
	}
	_, callerFile, _, _ := runtime.Caller(1)
	ex = filepath.Dir(callerFile)
	return ex
}
