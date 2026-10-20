// Package api contains the HTTP layer: routes, handlers and error mapping.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"minibank/internal/bank"
	"minibank/internal/store"
)

const maxBodyBytes = 1 << 20 // 1 MiB

var errBadJSON = errors.New("invalid json body")

type Server struct {
	store store.Store
	log   *slog.Logger
}

// New builds the router and wraps it with the middlewares.
func New(s store.Store, log *slog.Logger) http.Handler {
	srv := &Server{store: s, log: log}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthz)
	mux.HandleFunc("GET /readyz", srv.readyz)
	mux.HandleFunc("POST /accounts", srv.createAccount)
	mux.HandleFunc("GET /accounts", srv.listAccounts)
	mux.HandleFunc("GET /accounts/{id}", srv.getAccount)
	mux.HandleFunc("POST /accounts/{id}/deposit", srv.deposit)
	mux.HandleFunc("POST /accounts/{id}/withdraw", srv.withdraw)
	mux.HandleFunc("POST /transfers", srv.transfer)
	mux.HandleFunc("GET /accounts/{id}/transactions", srv.listTransactions)

	return logging(log, recovery(log, mux))
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// statusFor is the ONE place that maps errors to HTTP status codes.
func statusFor(err error) int {
	switch {
	case errors.Is(err, errBadJSON),
		errors.Is(err, bank.ErrInvalidOwner),
		errors.Is(err, bank.ErrInvalidAmount),
		errors.Is(err, bank.ErrInvalidID),
		errors.Is(err, bank.ErrSameAccount):
		return http.StatusBadRequest
	case errors.Is(err, bank.ErrAccountNotFound):
		return http.StatusNotFound
	case errors.Is(err, bank.ErrInsufficientFunds):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	status := statusFor(err)
	if status == http.StatusInternalServerError {
		s.log.Error("internal error", "err", err)
		writeError(w, status, "internal error")
		return
	}
	writeError(w, status, err.Error())
}

// decodeBody limits the body size and decodes JSON.
// Wrong types (string for a number, 10.5 for an int64) become errBadJSON.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errBadJSON
	}
	return nil
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, bank.ErrInvalidID
	}
	if err := bank.ValidateID(id); err != nil {
		return 0, err
	}
	return id, nil
}

// ---------- health ----------

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		s.log.Warn("readyz: ping failed", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
