package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int
	var dev bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Long: `Start the Stata web server to serve the dashboard.

In development mode (--dev), the server expects the React dev server
to be running separately and will proxy API requests.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServe(port, dev)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8081, "Port to listen on")
	cmd.Flags().BoolVar(&dev, "dev", false, "Run in development mode")

	return cmd
}

func runServe(port int, dev bool) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		// Auth routes (placeholder)
		r.Route("/auth", func(r chi.Router) {
			r.Get("/status", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"authenticated":false}`))
			})
		})
	})

	// Static file serving (placeholder for embedded files)
	r.Get("/*", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Stata</title></head>
<body>
<h1>Stata - Statistics for Strava</h1>
<p>Web UI will be served here. Run React dev server separately during development.</p>
<p><a href="/api/v1/health">API Health Check</a></p>
</body>
</html>`))
	})

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		close(done)
	}()

	slog.Info("starting server", "addr", addr, "dev", dev)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	slog.Info("server stopped")
	return nil
}
