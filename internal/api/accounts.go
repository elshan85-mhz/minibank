package api

import (
	"net/http"

	"minibank/internal/bank"
)

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Owner string `json:"owner"`
	}
	if err := decodeBody(w, r, &req); err != nil {
		s.fail(w, err)
		return
	}
	owner, err := bank.NormalizeOwner(req.Owner)
	if err != nil {
		s.fail(w, err)
		return
	}
	acc, err := s.store.CreateAccount(r.Context(), owner)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	accs, err := s.store.ListAccounts(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, accs)
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	acc, err := s.store.GetAccount(r.Context(), id)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}
