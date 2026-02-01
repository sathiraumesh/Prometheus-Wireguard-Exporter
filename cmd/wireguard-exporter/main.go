package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/internal/wgprometheus"
)

var (
	version = "dev"
	commit  = "unknown"
)

const DefaultPort = 9011

var port = flag.Int("p", getEnvInt("WIREGUARD_EXPORTER_PORT", DefaultPort), "the port to listen on (env: WIREGUARD_EXPORTER_PORT)")
var interfaces = flag.String("i", getEnvStr("WIREGUARD_EXPORTER_INTERFACES", ""), "comma-separated list of interfaces (env: WIREGUARD_EXPORTER_INTERFACES)")

func main() {
	flag.Parse()

	addr, err := parsePort(*port)
	if err != nil {
		slog.Error("invalid port", "error", err)
		os.Exit(1)
	}

	interfacesList := parseInterfaces(*interfaces)

	slog.Info("starting wireguard exporter",
		"address", addr,
		"version", version,
		"commit", commit,
	)

	collector := wgprometheus.NewCollector(interfacesList)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
