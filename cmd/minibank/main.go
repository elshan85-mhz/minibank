package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"minibank/internal/api"
	"minibank/internal/store"
)

type config struct {
	port        string
	storeKind   string
	databaseURL string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadConfig reads env vars and fails fast on invalid configuration.
func loadConfig() (config, error) {
	cfg := config{
		port:        getenv("PORT", "8080"),
		storeKind:   getenv("STORE", "postgres"),
		databaseURL: os.Getenv("DATABASE_URL"),
	}
	switch cfg.storeKind {
	case "memory":
	case "postgres":
		if cfg.databaseURL == "" {
			return cfg, errors.New("DATABASE_URL is required when STORE=postgres")
		}
	default:
		return cfg, fmt.Errorf("unknown STORE %q (use memory or postgres)", cfg.storeKind)
	}
	return cfg, nil
}

func openStore(cfg config) (store.Store, error) {
	switch cfg.storeKind {
	case "memory":
		return store.NewMemory(), nil
	default:
		return nil, errors.New("postgres store is not implemented yet")
	}
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := loadConfig()
	if err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	st, err := openStore(cfg)
	if err != nil {
		log.Error("cannot open store", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.port,
		Handler:           api.New(st, log),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("server starting", "addr", srv.Addr, "store", cfg.storeKind)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			st.Close()
			os.Exit(1)
		}
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "err", err)
	}
	if err := st.Close(); err != nil {
		log.Error("store close failed", "err", err)
	}
	log.Info("server stopped")
}
