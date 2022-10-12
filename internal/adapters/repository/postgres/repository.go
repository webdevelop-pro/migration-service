package postgres

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/db"
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
	const query = `INSERT INTO migration_service (name, version) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, query, name, ver)

	if err != nil {
		return errors.Wrapf(err, "query %s failed, params: %s %d", query, name, ver)
	}
	return nil
}

// GetServiceVersion returns currently deployed version of the service.
func (r *Repository) GetServiceVersion(ctx context.Context, name string) (int, error) {
	const query = `SELECT version FROM migration_service WHERE name=$1`

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
	const query = `CREATE TABLE migration_service (
	id serial NOT NULL PRIMARY KEY,
	name varchar NOT NULL UNIQUE,
	version int NOT NULL DEFAULT 0,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone NOT NULL DEFAULT NOW(),
	UNIQUE (name, version)
);
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_email_emails
  BEFORE UPDATE ON migration_service
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_set_timestamp();
COMMIT;
`
	_, err := r.db.Exec(ctx, query)

	if err != nil {
		return errors.Wrapf(err, "query %s failed.", query)
	}

	return nil
}
