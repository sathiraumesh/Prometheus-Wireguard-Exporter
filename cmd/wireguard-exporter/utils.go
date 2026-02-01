package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	minUserPort = 1024
	maxUserPort = 49151
)

func parsePort(port int) (string, error) {
	if port < minUserPort || port > maxUserPort {
		return "", fmt.Errorf("port must be between %d and %d, got %d",
			minUserPort, maxUserPort, port)
	}

	return ":" + strconv.Itoa(port), nil
}

func parseInterfaces(interfaceArg string) []string {
	interfaceArg = strings.TrimSpace(interfaceArg)

	if interfaceArg == "" {
		slog.Info("no interfaces specified, monitoring all WireGuard interfaces")
		return nil
	}

	return strings.Split(interfaceArg, ",")
}

func getEnvStr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid environment variable, using default", "key", key, "value", v, "default", fallback)
		return fallback
	}
	return i
}
