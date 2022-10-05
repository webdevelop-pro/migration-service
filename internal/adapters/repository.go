package adapters

import "context"

type Repository interface {
	GetServiceVersion(ctx context.Context, name string) (int, error)
	UpdateServiceVersion(ctx context.Context, name string, ver int) error
	CreateMigrationTable(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...interface{}) error
}
