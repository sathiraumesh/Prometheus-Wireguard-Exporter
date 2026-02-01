package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/internal/wgprometheus"
	"github.com/stretchr/testify/assert"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type mockDeviceLister struct {
	devices []*wgtypes.Device
	err     error
}

func (m *mockDeviceLister) Devices() ([]*wgtypes.Device, error) {
	return m.devices, m.err
}

func TestMetricsEndpoint(t *testing.T) {
	var key wgtypes.Key
	key[0] = 0x1

	mock := &mockDeviceLister{
		devices: []*wgtypes.Device{
			{
				Name: "wg0",
				Peers: []wgtypes.Peer{
					{
						PublicKey:         key,
						LastHandshakeTime: time.Unix(1700000000, 0),
						TransmitBytes:     500,
						ReceiveBytes:      1000,
						AllowedIPs:        []net.IPNet{{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(32, 32)}},
					},
				},
			},
		},
	}

	collector := wgprometheus.NewCollectorWithDevices(nil, mock)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	responseText := string(body)

	assert.Contains(t, responseText, "wireguard_latest_handshake_seconds")
	assert.Contains(t, responseText, "wireguard_transmitted_bytes")
	assert.Contains(t, responseText, "wireguard_received_bytes")
	assert.Contains(t, responseText, "wireguard_peer_up")
	assert.Contains(t, responseText, "wireguard_interface_info")
	assert.Contains(t, responseText, "wireguard_scrape_success")
	assert.Contains(t, responseText, "wireguard_scrape_duration_seconds")
	assert.Contains(t, responseText, `interface="wg0"`)
}

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	assert.Equal(t, "ok\n", string(body))
}
