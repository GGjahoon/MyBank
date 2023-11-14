package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provided all function to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	CreateUserTX(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	VerifyEmailTX(ctx context.Context, arg VerifyEmailTXParams) (VerifyEmailTXResult, error)
}

// SQLStore provided all function to execute db queries and transactions
// the real implementation of Store
type SQLStore struct {
	*Queries
	connPool *pgxpool.Pool
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		Queries:  New(connPool),
		connPool: connPool,
	}
}
