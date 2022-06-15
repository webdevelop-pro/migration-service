package services

import "context"

type Migration interface {
	Apply(ctx context.Context, serviceName string) (int, error)
}
