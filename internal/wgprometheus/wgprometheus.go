package wgprometheus

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var peerLabels = []string{"interface", "public_key", "allowed_ips"}

var (
	handshakeDesc = prometheus.NewDesc(
		"wireguard_latest_handshake_seconds",
		"Unix timestamp of the latest handshake for a WireGuard peer.",
		peerLabels, nil,
	)

	transmitDesc = prometheus.NewDesc(
		"wireguard_transmitted_bytes",
		"Total bytes transmitted to a WireGuard peer.",
		peerLabels, nil,
	)

	receivedDesc = prometheus.NewDesc(
		"wireguard_received_bytes",
		"Total bytes received from a WireGuard peer.",
		peerLabels, nil,
	)

	peerUpDesc = prometheus.NewDesc(
		"wireguard_peer_up",
		"Whether a WireGuard peer has had a recent handshake (1 = up, 0 = down).",
		peerLabels, nil,
	)

	interfaceInfoDesc = prometheus.NewDesc(
		"wireguard_interface_info",
		"Information about a WireGuard interface.",
		[]string{"interface", "public_key", "listen_port"}, nil,
	)

	scrapeSuccessDesc = prometheus.NewDesc(
		"wireguard_scrape_success",
		"Whether the last scrape of WireGuard metrics was successful (1 = success, 0 = failure).",
		nil, nil,
	)

	scrapeDurationDesc = prometheus.NewDesc(
		"wireguard_scrape_duration_seconds",
		"Duration of the last WireGuard metrics scrape in seconds.",
		nil, nil,
	)
)

// PeerHandshakeTimeout is the duration after which a peer is considered down
// if no handshake has occurred.
const PeerHandshakeTimeout = 5 * time.Minute

// DeviceLister abstracts WireGuard device enumeration for testability.
type DeviceLister interface {
	Devices() ([]*wgtypes.Device, error)
}

// wgDeviceLister creates a new wgctrl client per call.
type wgDeviceLister struct{}

func (w *wgDeviceLister) Devices() ([]*wgtypes.Device, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.Devices()
}

// Collector implements prometheus.Collector and fetches WireGuard
// metrics on each Prometheus scrape.
type Collector struct {
	devices    DeviceLister
	monitorSet map[string]struct{}
}

// NewCollector creates a Collector that monitors the given interfaces.
// If monitorKeys is empty, all WireGuard interfaces are monitored.
func NewCollector(monitorKeys []string) *Collector {
	set := make(map[string]struct{}, len(monitorKeys))
	for _, key := range monitorKeys {
		set[strings.TrimSpace(key)] = struct{}{}
	}
	return &Collector{
		devices:    &wgDeviceLister{},
		monitorSet: set,
	}
}

// NewCollectorWithDevices creates a Collector with a custom DeviceLister,
// useful for testing.
func NewCollectorWithDevices(monitorKeys []string, devices DeviceLister) *Collector {
	c := NewCollector(monitorKeys)
	c.devices = devices
	return c
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- handshakeDesc
	ch <- transmitDesc
	ch <- receivedDesc
	ch <- peerUpDesc
	ch <- interfaceInfoDesc
	ch <- scrapeSuccessDesc
	ch <- scrapeDurationDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	devices, err := c.devices.Devices()
	if err != nil {
		slog.Error("failed to list WireGuard devices", "error", err)
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds())
		return
	}

	for _, dev := range devices {
		if !c.shouldMonitor(dev.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			interfaceInfoDesc, prometheus.GaugeValue, 1,
			dev.Name, dev.PublicKey.String(), fmt.Sprintf("%d", dev.ListenPort),
		)

		for _, peer := range dev.Peers {
			ifName := dev.Name
			pubKey := peer.PublicKey.String()
			allowedIPs := fmt.Sprintf("%v", peer.AllowedIPs)

			ch <- prometheus.MustNewConstMetric(
				handshakeDesc, prometheus.GaugeValue,
				float64(peer.LastHandshakeTime.Unix()),
				ifName, pubKey, allowedIPs,
			)
			ch <- prometheus.MustNewConstMetric(
				transmitDesc, prometheus.GaugeValue,
				float64(peer.TransmitBytes),
				ifName, pubKey, allowedIPs,
			)
			ch <- prometheus.MustNewConstMetric(
				receivedDesc, prometheus.GaugeValue,
				float64(peer.ReceiveBytes),
				ifName, pubKey, allowedIPs,
			)

			up := 0.0
			if !peer.LastHandshakeTime.IsZero() && time.Since(peer.LastHandshakeTime) < PeerHandshakeTimeout {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				peerUpDesc, prometheus.GaugeValue,
				up,
				ifName, pubKey, allowedIPs,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds())
}

func (c *Collector) shouldMonitor(name string) bool {
	if len(c.monitorSet) == 0 {
		return true
	}
	_, ok := c.monitorSet[name]
	return ok
}
