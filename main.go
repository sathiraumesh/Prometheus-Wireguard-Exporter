package main

import (
	"bytes"
	"net/http"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wiregaurd_exporter/wireguard"
)

var (
	wireguardLatestHandshakeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_latest_handshake_seconds",
			Help: "The latest handshake time for Wireguard connections.",
		},
		[]string{"interface", "public_key"},
	)

	registry = prometheus.NewRegistry()
)

func init() {
	registry.MustRegister(wireguardLatestHandshakeSeconds)
}

func main() {
	go func() {
		for {
			stats, err := getStats()
			if err != nil {
				panic(err)
			}

			connections := wireguard.ListConnections(stats)
			updateMetrics(connections)
			time.Sleep(time.Second * 5)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	))
	http.ListenAndServe(":9011", nil)
}

func getStats() (string, error) {
	cmd := exec.Command("wg", "show")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}

func updateMetrics(connections map[string]*wireguard.Connection) {

	for publicKey, conn := range connections {
		wireguardLatestHandshakeSeconds.WithLabelValues(
			conn.Interface,
			publicKey,
		).Set(float64(conn.LatestHandshake))
	}
}
