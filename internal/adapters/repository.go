package adapters

import "context"

type Repository interface {
	GetServiceVersion(ctx context.Context, name string) (int, error)
	UpdateServiceVersion(ctx context.Context, name string, ver int) error
	Exec(ctx context.Context, sql string, arguments ...interface{}) error
}
