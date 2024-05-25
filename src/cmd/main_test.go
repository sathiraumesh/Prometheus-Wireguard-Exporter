package main

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/wgprometheus"
	"github.com/stretchr/testify/assert"
)

// mocked default flag values that we expect for exporter
var defaultPort = flag.Int("test-default-p", DEFALULT_PORT, "the port to listen on")
var defaultInterface = flag.String("test-default-i", "", "comma-separated list of interfaces")

// mocked custom flag values that we expect for exporter
var customPort = flag.Int("test-custom-i", 8080, "the port to listen on")
var customInterface = flag.String("test-custom-p", "wg0,wg1", "comma-separated list of interfaces")

func TestValidatesDefaultFlags(t *testing.T) {
	interfaces, port, _ := validateReturnFlags(*defaultInterface, *defaultPort)

	assert.Empty(t, interfaces, "default flags for interface should be empty")

	expectedPort := ":" + strconv.Itoa(*defaultPort)
	assert.Equalf(t, port, expectedPort, "invalid default port %s", port)
}

func TestValidateCustomFlags(t *testing.T) {
	interfaces, port, _ := validateReturnFlags(*customInterface, *customPort)

	assert.NotEmpty(t, interfaces, "custom flags for (-i) interface should not be empty")
	assert.Equal(t, len(interfaces), 2, "invalid interface count")

	expectedPort := ":" + strconv.Itoa(*customPort)
	assert.Equalf(t, port, expectedPort, "invalid custom port %s", port)
}

func TestMetricsEndpoint(t *testing.T) {

	registry := wgprometheus.GetRegistry()

	go wgprometheus.ScrapConnectionStats([]string{}, SCRAP_INTERVAL)

	handler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	)

	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	// we wait for some time for some time until the connection stats goroutine is run
	time.Sleep(2 * time.Second)

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

	assert.Contains(t, responseText, `SS"`)
}
