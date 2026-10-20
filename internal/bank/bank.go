// Package bank contains the domain types and validation rules of MiniBank.
package bank

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinOwnerRunes = 2
	MaxOwnerRunes = 40
	MaxAmount     = 1_000_000_000
)

var (
	ErrInvalidOwner      = errors.New("invalid owner")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrSameAccount       = errors.New("cannot transfer to the same account")
	ErrInvalidID         = errors.New("invalid id")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
)

const (
	TypeDeposit  = "deposit"
	TypeWithdraw = "withdraw"
	TypeTransfer = "transfer"
)

type Account struct {
	ID        int64     `json:"id"`
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

// From/To are 0 when absent; ids are always positive, so omitempty is safe.
type Transaction struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	From      int64     `json:"from,omitempty"`
	To        int64     `json:"to,omitempty"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// NormalizeOwner trims spaces and checks the length in runes (not bytes).
func NormalizeOwner(owner string) (string, error) {
	owner = strings.TrimSpace(owner)
	n := utf8.RuneCountInString(owner)
	if n < MinOwnerRunes || n > MaxOwnerRunes {
		return "", ErrInvalidOwner
	}
	return owner, nil
}

// ValidateAmount accepts 1..MaxAmount.
func ValidateAmount(amount int64) error {
	if amount <= 0 || amount > MaxAmount {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateID accepts positive ids only.
func ValidateID(id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return nil
}
