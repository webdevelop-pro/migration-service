package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/webdevelop-pro/go-common/configurator"
	"github.com/webdevelop-pro/go-common/db"
)

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

	var ver int

	err := r.db.QueryRow(ctx, query, name).Scan(&ver)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
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

// UpdateServiceVersion updates service version.
func (r *Repository) CreateMigrationTable(ctx context.Context) error {
	const query = `CREATE TABLE migration_service (
		id serial NOT NULL PRIMARY KEY,
		name varchar NOT NULL UNIQUE,
		version int NOT NULL DEFAULT 0,
		created_at timestamp with time zone DEFAULT now() NOT NULL,
		UNIQUE (name, version)
	);`
	_, err := r.db.Exec(ctx, query)

	if err != nil {
		return errors.Wrapf(err, "query %s failed.", query)
	}

	return nil
}
