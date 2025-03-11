package storage

import "context"

type StatusChecker interface {
	CheckStatus(ctx context.Context) error
}
