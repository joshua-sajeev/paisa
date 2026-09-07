package ports

import "context"

// TxManager executes a function within a database transaction.
type TxManager interface {
	WithinTransaction(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
}
