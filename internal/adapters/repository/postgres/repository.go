package postgres

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/db"
	"github.com/webdevelop-pro/migration-service/internal/domain/migration_log"
)

const NO_TABLE_CODE = "42P01"

type Repository struct {
	db *db.DB
}

// New returns new DB instance.
func New(c *configurator.Configurator) *Repository {
	return &Repository{
		db: db.New(c),
	}
}

// UpdateServiceVersion updates service version.
func (r *Repository) UpdateServiceVersion(ctx context.Context, name string, ver int) error {
	const query = `INSERT INTO migration_services (name, version) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET version=$2`
	_, err := r.db.Exec(ctx, query, name, ver)

	if err != nil {
		return errors.Wrapf(err, "query %s failed, params: %s %d", query, name, ver)
	}
	return nil
}

// GetServiceVersion returns currently deployed version of the service.
func (r *Repository) GetServiceVersion(ctx context.Context, name string) (int, error) {
	const query = `SELECT version FROM migration_services WHERE name=$1`

	var pgErr *pgconn.PgError
	var ver int

	err := r.db.QueryRow(ctx, query, name).Scan(&ver)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		if errors.As(err, &pgErr) {
			if pgErr.Code == NO_TABLE_CODE {
				if err := r.CreateMigrationTable(context.Background()); err != nil {
					return 0, errors.Wrapf(err, "query %s failed, %s ", query, pgErr.Message)
				}
				return r.GetServiceVersion(ctx, name)
			}
			return 0, errors.Wrapf(err, "query %s failed, %s ", query, pgErr.Message)
		}
		return 0, errors.Wrapf(err, "query %s failed, %s ", query, name)
	}

	return ver, nil
}

// Exec executes query
func (r *Repository) Exec(ctx context.Context, sql string, arguments ...interface{}) error {
	return r.db.BeginFunc(
		ctx,
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, sql, arguments...)

			return err
		},
	)
}

// CreateMigrationTable will create a migration table
func (r *Repository) CreateMigrationTable(ctx context.Context) error {
	const query = `CREATE TABLE IF NOT EXISTS migration_services (
	id serial NOT NULL PRIMARY KEY,
	name varchar NOT NULL UNIQUE,
	version int NOT NULL DEFAULT 0,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone NOT NULL DEFAULT NOW()
);
CREATE OR REPLACE FUNCTION update_at_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER set_timestamp_migration_services
  BEFORE UPDATE ON migration_services
  FOR EACH ROW
  EXECUTE PROCEDURE update_at_set_timestamp();
COMMIT;

CREATE TABLE IF NOT EXISTS migration_services_log
(
    id                      SERIAL PRIMARY KEY,

    -- required
    migration_services_name character varying(255) NOT NULL,
    priority                integer                NOT NULL,
    version                 integer                NOT NULL,
    file_name               character varying(255) NOT NULL,
    sql                     text                   NOT NULL,
    hash                    character varying(255) NOT NULL,

    -- dates
    created_at              timestamptz            NOT NULL DEFAULT now(),
    updated_at              timestamptz            NOT NULL DEFAULT now()
);

ALTER TABLE public.migration_services_log DROP CONSTRAINT IF EXISTS migration_services_log_pk;
ALTER TABLE public.migration_services_log
    ADD CONSTRAINT migration_services_log_pk
        UNIQUE (migration_services_name, priority, version, file_name);

CREATE OR REPLACE TRIGGER migration_services_log_updated_at_timestamp
    BEFORE UPDATE
    ON migration_services_log
    FOR EACH ROW
EXECUTE PROCEDURE update_at_set_timestamp();

CREATE INDEX IF NOT EXISTS migration_services_log_hash_index
    on migration_services_log (hash);
COMMIT;
`
	_, err := r.db.Exec(ctx, query)

	if err != nil {
		return errors.Wrapf(err, "query %s failed.", query)
	}

	return nil
}

// WriteMigrationServiceLog inserts row to migration_services_log
func (r *Repository) WriteMigrationServiceLog(ctx context.Context, log migration_log.MigrationServicesLog) error {
	const query = `INSERT INTO migration_services_log (migration_services_name, priority, version, file_name, "sql", hash) 
		VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT(migration_services_name, priority, version, file_name) DO UPDATE 
		SET "sql"=$5, hash=$6`
	_, err := r.db.Exec(ctx, query, log.MigrationServiceName, log.Priority, log.Version, log.FileName, log.SQL, log.Hash)

	if err != nil {
		return errors.Wrapf(err, "query %s failed, params: MigrationServiceName = %s, Priority = %d, "+
			"Version = %d, FileName = %s, SQL = %s, Hash = %s", query, log.MigrationServiceName, log.Priority,
			log.Version, log.FileName, log.SQL, log.Hash)
	}
	return nil
}
