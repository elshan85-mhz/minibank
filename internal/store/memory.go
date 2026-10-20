package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"minibank/internal/bank"
)

// Memory keeps everything in maps guarded by ONE mutex.
type Memory struct {
	mu        sync.Mutex
	accounts  map[int64]*bank.Account
	txs       []bank.Transaction
	nextAccID int64
	nextTxID  int64
}

var _ Store = (*Memory)(nil)

func NewMemory() *Memory {
	return &Memory{
		accounts:  make(map[int64]*bank.Account),
		nextAccID: 1,
		nextTxID:  1,
	}
}

func (m *Memory) CreateAccount(_ context.Context, owner string) (bank.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc := &bank.Account{
		ID:        m.nextAccID,
		Owner:     owner,
		CreatedAt: time.Now().UTC(),
	}
	m.nextAccID++
	m.accounts[acc.ID] = acc
	return *acc, nil
}

func (m *Memory) ListAccounts(_ context.Context) ([]bank.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]bank.Account, 0, len(m.accounts)) // never nil -> JSON []
	for _, a := range m.accounts {
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *Memory) GetAccount(_ context.Context, id int64) (bank.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, ok := m.accounts[id]
	if !ok {
		return bank.Account{}, bank.ErrAccountNotFound
	}
	return *acc, nil
}

func (m *Memory) Deposit(_ context.Context, id, amount int64) (bank.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, ok := m.accounts[id]
	if !ok {
		return bank.Account{}, bank.ErrAccountNotFound
	}
	acc.Balance += amount
	m.record(bank.TypeDeposit, 0, id, amount)
	return *acc, nil
}

func (m *Memory) Withdraw(_ context.Context, id, amount int64) (bank.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, ok := m.accounts[id]
	if !ok {
		return bank.Account{}, bank.ErrAccountNotFound
	}
	if acc.Balance < amount {
		return bank.Account{}, bank.ErrInsufficientFunds
	}
	acc.Balance -= amount
	m.record(bank.TypeWithdraw, id, 0, amount)
	return *acc, nil
}

// Transfer takes the lock ONCE and changes both accounts under it.
// Error order: 404 (either account missing) before 409 (insufficient funds).
func (m *Memory) Transfer(_ context.Context, from, to, amount int64) (bank.Transaction, error) {
	if from == to {
		return bank.Transaction{}, bank.ErrSameAccount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	src, okFrom := m.accounts[from]
	dst, okTo := m.accounts[to]
	if !okFrom || !okTo {
		return bank.Transaction{}, bank.ErrAccountNotFound
	}
	if src.Balance < amount {
		return bank.Transaction{}, bank.ErrInsufficientFunds
	}
	src.Balance -= amount
	dst.Balance += amount
	return m.record(bank.TypeTransfer, from, to, amount), nil
}

func (m *Memory) ListTransactions(_ context.Context, accountID int64) ([]bank.Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.accounts[accountID]; !ok {
		return nil, bank.ErrAccountNotFound
	}
	out := make([]bank.Transaction, 0) // never nil -> JSON []
	for _, t := range m.txs {          // txs is already oldest-first
		if t.From == accountID || t.To == accountID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *Memory) Ping(_ context.Context) error { return nil }

func (m *Memory) Close() error { return nil }

// record must be called with m.mu held.
func (m *Memory) record(typ string, from, to, amount int64) bank.Transaction {
	t := bank.Transaction{
		ID:        m.nextTxID,
		Type:      typ,
		From:      from,
		To:        to,
		Amount:    amount,
		CreatedAt: time.Now().UTC(),
	}
	m.nextTxID++
	m.txs = append(m.txs, t)
	return t
}
