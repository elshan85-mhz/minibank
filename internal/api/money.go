package api

import (
	"net/http"

	"minibank/internal/bank"
)

type amountRequest struct {
	Amount int64 `json:"amount"`
}

func (s *Server) deposit(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var req amountRequest
	if err := decodeBody(w, r, &req); err != nil {
		s.fail(w, err)
		return
	}
	if err := bank.ValidateAmount(req.Amount); err != nil {
		s.fail(w, err)
		return
	}
	acc, err := s.store.Deposit(r.Context(), id, req.Amount)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (s *Server) withdraw(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var req amountRequest
	if err := decodeBody(w, r, &req); err != nil {
		s.fail(w, err)
		return
	}
	if err := bank.ValidateAmount(req.Amount); err != nil {
		s.fail(w, err)
		return
	}
	acc, err := s.store.Withdraw(r.Context(), id, req.Amount)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From   int64 `json:"from"`
		To     int64 `json:"to"`
		Amount int64 `json:"amount"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		s.fail(w, err)
		return
	}
	// All 400-class checks first, then the store decides 404 / 409.
	if err := bank.ValidateID(req.From); err != nil {
		s.fail(w, err)
		return
	}
	if err := bank.ValidateID(req.To); err != nil {
		s.fail(w, err)
		return
	}
	if req.From == req.To {
		s.fail(w, bank.ErrSameAccount)
		return
	}
	if err := bank.ValidateAmount(req.Amount); err != nil {
		s.fail(w, err)
		return
	}
	tx, err := s.store.Transfer(r.Context(), req.From, req.To, req.Amount)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (s *Server) listTransactions(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	txs, err := s.store.ListTransactions(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, txs)
}
