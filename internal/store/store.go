// Package store defines the persistence interface and its implementations.
package store

import (
	"context"

	"minibank/internal/bank"
)

// Store is implemented by the in-memory store and the PostgreSQL store.
type Store interface {
	CreateAccount(ctx context.Context, owner string) (bank.Account, error)
	ListAccounts(ctx context.Context) ([]bank.Account, error)
	GetAccount(ctx context.Context, id int64) (bank.Account, error)
	Deposit(ctx context.Context, id, amount int64) (bank.Account, error)
	Withdraw(ctx context.Context, id, amount int64) (bank.Account, error)
	Transfer(ctx context.Context, from, to, amount int64) (bank.Transaction, error)
	ListTransactions(ctx context.Context, accountID int64) ([]bank.Transaction, error)
	Ping(ctx context.Context) error
	Close() error
}
