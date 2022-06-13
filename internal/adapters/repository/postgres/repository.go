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
	const query = `INSERT INTO migration_service (name, version) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET version=$2`
	_, err := r.db.Exec(ctx, query, name, ver)

	if err != nil {
		return errors.Wrap(err, "query failed")
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

		return 0, errors.Wrap(err, "query failed")
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
