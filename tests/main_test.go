package main

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/adapters/repository/postgres"
	"github.com/webdevelop-pro/migration-service/internal/app"
)

type sqlFiles struct {
	filename string
	sql      string
}

func testInit() (logger.Logger, *configurator.Configurator, *postgres.Repository, *app.App, *db.DB, context.Context) {
	// migration folder is required
	exec.Command("mkdir", "./migrations/").Output()
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

func setUp(log logger.Logger, files []sqlFiles) {
	// old left overs
	exec.Command("rm", "-rf", "./migrations/").Output()
	// ToDo
	// flush system cache?
	time.Sleep(2 * time.Second)
	for _, file := range files {
		dir, err := exec.Command("dirname", file.filename).Output()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot create a dicrectory")
		}

		if _, err := exec.Command("mkdir", "-p", string(dir[0:len(dir)-1])).Output(); err != nil {
			log.Fatal().Err(err).Msg("cannot create a dicrectory")
		}

		if err := os.WriteFile(file.filename, []byte(file.sql), 0644); err != nil {
			log.Fatal().Err(err).Msg("cannot create a file")
		}
	}
}

func checkResults(t *testing.T, rawPG *db.DB, log logger.Logger, expName string, expVer int) {
	ctx := context.Background()
	name := ""
	ver := 0
	query := "SELECT name, version FROM migration_service ORDER by id DESC LIMIT 1"
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
	log, _, pg, migration, rawPG, ctx := testInit()

	err := pg.CreateMigrationTable(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migration table")
	}

	initSqls := []sqlFiles{
		{
			"./migrations/01_user/01_init.sql",
			`--- some comment
CREATE TABLE user_users (
    id serial not null primary key,
    name varchar(150) not null default ''
);`,
		},
		{
			"./migrations/01_user/01_seed_test.yaml",
			`service: user
	migrations:
	- version: 1
	  allowError: false
	  queries:
	  - |
		DELETE FROM user_users;
		INSERT INTO user_users (name)  VALUES ('Maria');`,
		},
		{
			"./migrations/01_user/seeds/02_seed.sql",
			`insert into user_users(name) values('tamata')`,
		},
	}
	setUp(log, initSqls)

	if err := migration.ApplyAll(); err != nil {
		log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	checkResults(t, rawPG, log, "user", 2)
}

// TestServicePriorities checks if services executed in correct order
func TestServicePriorities(t *testing.T) {
	// we will create new migration for email service
	// and verify if migration will be applied in correct order
	// first user and then migration
	log, _, _, migration, rawPG, _ := testInit()

	// Create two different services with different indexes
	// make sure migration executed in correct order
	initSqls := []sqlFiles{
		{
			"./migrations/02_email/01_init.sql",
			`
	CREATE TABLE email_emails (
	id serial primary key,
	user_id integer not null,
		FOREIGN KEY (user_id) REFERENCES user_users ("id")
		ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED);
	CREATE INDEX email_emails_user_fk ON email_emails USING btree ("user_id");`,
		},
		{
			"./migrations/01_user/01_init.sql",
			`--- some comment
CREATE TABLE user_users (
    id serial not null primary key,
    name varchar(150) not null default ''
);`,
		},
	}
	setUp(log, initSqls)

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

	checkResults(t, rawPG, log, "email", 1)
}

func TestAllowError(t *testing.T) {
	// we will create new migration for email service
	// and verify if migration will be applied in correct order
	// first user and then migration
	log, _, _, migration, rawPG, _ := testInit()

	// Create two different services with different indexes
	// make sure migration executed in correct order
	initSqls := []sqlFiles{
		{
			"./migrations/03_error_errors/01_test-allow-error.sql",
			`--- allow_error: true
	THIS IS SQL WITH AN ERROR`,
		},
	}
	setUp(log, initSqls)

	if err := migration.ApplyAll(); err != nil {
		log.Fatal().Err(err).Msg("cannot apply migrations")
	}

	checkResults(t, rawPG, log, "error_errors", 1)
}

// TestMigrationPriorities checks if files executed in correct order
func TestMigrationPriorities(t *testing.T) {
	// we will create new migration for email service
	// and verify if migration will be applied in correct order
	// first user and then migration
	log, _, _, migration, rawPG, _ := testInit()

	log.Debug().Msg("trying to apply migration")

	initSqls := []sqlFiles{
		{
			"./migrations/01_user_user/04_add_bitint.sql",
			`ALTER TABLE user_users ADD COLUMN external_id bigint default 0;`,
		},
		{
			"./migrations/01_user_user/01_init.sql",
			`--- some comment
CREATE TABLE user_users (
    id serial not null primary key,
    name varchar(150) not null default ''
);`,
		},
		{
			"./migrations/02_email_emails/02_add_id.sql",
			`ALTER TABLE email_emails ADD COLUMN external_id bigint default 0;`,
		},
		{
			"./migrations/02_email_emails/01_create.sql",
			`CREATE TABLE email_emails (id serial not null primary key);`,
		},
		{
			"./migrations/01_user_user/02_add_email.sql",
			`--- some comment
	ALTER TABLE user_users ADD email varchar(150) not null default '' UNIQUE;`,
		},
	}
	setUp(log, initSqls)

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

	checkResults(t, rawPG, log, "email_emails", 2)
}
