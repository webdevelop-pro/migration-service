package main

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
)

func testInit() (logger.Logger, *configurator.Configurator, *postgres.Repository, *app.App, *db.DB, context.Context) {
	log := logger.NewDefault()
	c := configurator.New()
	pg := postgres.New(c)
	migration := app.New(c, pg)
	rawPG := db.New(c)
	ctx := context.Background()

	rawPG.Exec(context.Background(), "DROP TABLE IF EXISTS email_emails")
	rawPG.Exec(context.Background(), "DROP TABLE IF EXISTS user_users")
	rawPG.Exec(context.Background(), "DROP TABLE IF EXISTS migration_service")

	return log, c, pg, migration, rawPG, ctx
}

func TestApplyAllMigration(t *testing.T) {
	log, _, pg, migration, rawPG, ctx := testInit()

	err := pg.CreateMigrationTable(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migration table")
	}

	if err := migration.ApplyAll(); err != nil {
		log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	name := ""
	ver := 0
	query := "SELECT name, version FROM migration_service ORDER by id DESC LIMIT 1"
	if err = rawPG.QueryRow(ctx, query).Scan(&name, &ver); err != nil {
		log.Fatal().Err(err).Msg("cannot get values for migration service")
	}

	if name != "user" || ver != 2 {
		log.Fatal().Msgf("data does not match 'user'!=%s or 1!=%d", name, ver)
	}
}

func TestServicePriorities(t *testing.T) {
	// we will create new migration for email service
	// and verify if migration will be applied in correct order
	// first user and then migration
	log, _, pg, migration, rawPG, ctx := testInit()

	err := pg.CreateMigrationTable(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migration table")
	}

	// old left overs
	exec.Command("rm", "-rf", "./migrations/02_email").Output()

	// create new email table
	SQL := `
	CREATE TABLE email_emails (
	id serial primary key,
	user_id integer not null,
		FOREIGN KEY (user_id) REFERENCES user_users ("id")
		ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED);
	CREATE INDEX email_emails_user_fk ON email_emails USING btree ("user_id");`
	if err := os.Mkdir("./migrations/02_email", 0755); err != nil {
		log.Fatal().Err(err).Msg("cannot create a dicrectory")
	}
	if err := os.WriteFile("./migrations/02_email/01_init.sql", []byte(SQL), 0644); err != nil {
		log.Fatal().Err(err).Msg("cannot create a file")
	}

	// ToDo
	// check uniqueness of services numbers
	/*
		// broken file record, migration should not be applied create new email table
		if err := os.WriteFile("./migrations/02_email/01_init-second-time.sql", []byte(SQL), 0644); err != nil {
			log.Fatal().Err(err).Msg("cannot create a file")
		}
	*/
	if err := migration.ApplyAll(); err != nil {
		log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	name := ""
	ver := 0
	query := "SELECT name, version FROM migration_service ORDER by id DESC LIMIT 1"
	if err = rawPG.QueryRow(ctx, query).Scan(&name, &ver); err != nil {
		log.Fatal().Err(err).Msg("cannot get values for migration service")
	}
	if name != "email" || ver != 1 {
		log.Fatal().Msgf("data does not match 'email'!=%s or 1!=%d", name, ver)
	}
}

func TestMigrationPriorities(t *testing.T) {
	// we will create new migration for email service
	// and verify if migration will be applied in correct order
	// first user and then migration
	log, _, pg, migration, rawPG, ctx := testInit()

	err := pg.CreateMigrationTable(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migration table")
	}
	log.Debug().Msg("trying to apply migration")

	// create new email table
	SQL := `ALTER TABLE user_users ADD COLUMN external_id bigint default 0`
	if err := os.WriteFile("./migrations/01_user/03_add_bitint.sql", []byte(SQL), 0644); err != nil {
		log.Fatal().Err(err).Msg("cannot create a file")
	}

	// ToDo
	// check uniqueness of services numbers
	/*
		// broken file record, migration should not be applied since we have order duplication
		if err := os.WriteFile("./migrations/01_user/03_add_bitint-for-second-time.sql", []byte(SQL), 0644); err != nil {
			log.Fatal().Err(err).Msg("cannot create a file")
		}
	*/

	if err := migration.ApplyAll(); err != nil {
		log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	name := ""
	ver := 0
	query := "SELECT name, version FROM migration_service ORDER by id ASC LIMIT 1"
	if err = rawPG.QueryRow(ctx, query).Scan(&name, &ver); err != nil {
		log.Fatal().Err(err).Msg("cannot get values for migration service")
	}
	if name != "user" || ver != 3 {
		log.Fatal().Msgf("data does not match 'user'!=%s or 3!=%d", name, ver)
	}
}
