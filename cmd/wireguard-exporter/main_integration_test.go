//go:build integration

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/internal/wgprometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsEndpointIntegration(t *testing.T) {
	collector := wgprometheus.NewCollector(nil)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Retry until WireGuard interface is fully up instead of a fixed sleep
	var responseText string
	require.Eventually(t, func() bool {
		resp, err := http.Get(testServer.URL + "/metrics")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}
		responseText = string(body)
		return strings.Contains(responseText, `interface="wg0"`)
	}, 30*time.Second, 1*time.Second, "WireGuard metrics not available within timeout")

	// peer public key from wg0_host_2.conf -> host 1's public key
	assert.Contains(t, responseText, `public_key="HYf+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSxxI="`)

	// all metric families should be present
	assert.Contains(t, responseText, "wireguard_latest_handshake_seconds")
	assert.Contains(t, responseText, "wireguard_transmitted_bytes")
	assert.Contains(t, responseText, "wireguard_received_bytes")
	assert.Contains(t, responseText, "wireguard_peer_up")
	assert.Contains(t, responseText, "wireguard_interface_info")
	assert.Contains(t, responseText, "wireguard_scrape_success")
	assert.Contains(t, responseText, "wireguard_scrape_duration_seconds")
}
