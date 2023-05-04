package main

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
	"github.com/webdevelop-pro/migration-service/internal/domain/migration"
)

type sqlFiles struct {
	filename string
	sql      string
}

func testInit() (logger.Logger, *configurator.Configurator, *postgres.Repository, *app.App, *db.DB, context.Context) {
	_log := logger.NewDefault()
	c := configurator.New()
	pg := postgres.New(c)
	_migration := app.New(c, pg)
	rawPG := db.New(c)
	ctx := context.Background()

	_, err := rawPG.Exec(context.Background(), "DROP TABLE IF EXISTS email_emails")
	if err != nil {
		_log.Fatal().Err(err).Msg("can't drop table email_emails from DB")
	}
	_, err = rawPG.Exec(context.Background(), "DROP TABLE IF EXISTS user_users")
	if err != nil {
		_log.Fatal().Err(err).Msg("can't drop table user_users from DB")
	}
	_, err = rawPG.Exec(context.Background(), "DROP TABLE IF EXISTS migration_services")
	if err != nil {
		_log.Fatal().Err(err).Msg("can't drop table migration_services from DB")
	}

	return _log, c, pg, _migration, rawPG, ctx
}

func checkResults(t *testing.T, rawPG *db.DB, log logger.Logger, expName string, expVer int) {
	ctx := context.Background()
	name := ""
	ver := 0
	query := "SELECT name, version FROM migration_services ORDER by id DESC LIMIT 1"
	if err := rawPG.QueryRow(ctx, query).Scan(&name, &ver); err != nil {
		log.Fatal().Err(err).Msg("cannot get values for migration service")
		t.Error()
	}
	if name != expName || ver != expVer {
		log.Fatal().Msgf("data does not match %s!=%s or %d!=%d", name, expName, ver, expVer)
		t.Error()
	}
}

// TestIgnoreNonSQLFiles checks if only *.sql files are applied
func TestIgnoreNonSQLFiles(t *testing.T) {
	_log, _, pg, _migration, rawPG, ctx := testInit()

	err := pg.CreateMigrationTable(ctx)
	if err != nil {
		_log.Fatal().Err(err).Msg("cannot create migration table")
	}

	if err := _migration.ApplyAll("./migrations/TestIgnoreNonSQLFiles"); err != nil {
		_log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	checkResults(t, rawPG, _log, "user_seeds", 2)
}

// TestServicePriorities checks if services executed in correct order
func TestServicePriorities(t *testing.T) {
	// we will create new _migration for email service
	// and verify if _migration will be applied in correct order
	// first user and then _migration
	_log, _, _, _migration, rawPG, _ := testInit()

	// Create two different services with different indexes
	// make sure _migration executed in correct order

	// ToDo
	// check uniqueness of services numbers
	/*
		// broken file record, _migration should not be applied create new email table
		if err := os.WriteFile("./migrations/02_email/01_init-second-time.sql", []byte(SQL), 0644); err != nil {
			_log.Fatal().Err(err).Msg("cannot create a file")
		}
	*/
	if err := _migration.ApplyAll("./migrations/TestServicePriorities"); err != nil {
		_log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	checkResults(t, rawPG, _log, "email", 1)
}

func TestAllowError(t *testing.T) {
	// we will create new migration for email service
	// and verify if migration will be applied in correct order
	// first user and then migration
	mig := migration.NewMigration([]string{`--
	-- --   allow_error: true  
	--
	THIS IS SQL WITH AN ERROR`,
	}, "./migration")

	if mig.AllowError != true {
		log.Error().Msg("allow error should be true")
		t.Fail()
	}
}

// TestMigrationPriorities checks if files executed in correct order
func TestMigrationPriorities(t *testing.T) {
	// we will create new _migration for email service
	// and verify if _migration will be applied in correct order
	// first user and then _migration
	_log, _, _, _migration, rawPG, _ := testInit()

	_log.Debug().Msg("trying to apply _migration")

	// ToDo
	// check uniqueness of services numbers
	/*
		// broken file record, _migration should not be applied since we have order duplication
		if err := os.WriteFile("./migrations/01_user/03_add_bitint-for-second-time.sql", []byte(SQL), 0644); err != nil {
			_log.Fatal().Err(err).Msg("cannot create a file")
		}
	*/

	if err := _migration.ApplyAll("./migrations/TestMigrationPriorities"); err != nil {
		_log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	checkResults(t, rawPG, _log, "email_emails", 2)
}

// TestMigrationApplied checks applied migrations commited to db
func TestMigrationCommited(t *testing.T) {
	// we will create new _migration for email service
	// and verify if _migration will be applied in correct order
	// first user and then _migration
	_log, _, _, _migration, rawPG, _ := testInit()

	_log.Debug().Msg("trying to apply _migration")

	// ToDo
	// check uniqueness of services numbers
	/*
		// broken file record, _migration should not be applied since we have order duplication
		if err := os.WriteFile("./migrations/01_user/03_add_bitint-for-second-time.sql", []byte(SQL), 0644); err != nil {
			_log.Fatal().Err(err).Msg("cannot create a file")
		}
	*/

	if err := _migration.ApplyAll("./migrations/TestMigrationCommited"); err == nil {
		_log.Fatal().Msg("last _migration should fail")
		t.Fail()
	}

	checkResults(t, rawPG, _log, "user_users", 1)
}
