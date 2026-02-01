package wgprometheus

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type mockDeviceLister struct {
	devices []*wgtypes.Device
	err     error
}

func (m *mockDeviceLister) Devices() ([]*wgtypes.Device, error) {
	return m.devices, m.err
}

func newTestPeer(pubKeyByte byte, transmit, received int64, handshake time.Time) wgtypes.Peer {
	var key wgtypes.Key
	key[0] = pubKeyByte
	return wgtypes.Peer{
		PublicKey:          key,
		LastHandshakeTime:  handshake,
		TransmitBytes:     transmit,
		ReceiveBytes:       received,
		AllowedIPs:         []net.IPNet{{IP: net.IPv4(10, 0, 0, 1), Mask: net.CIDRMask(32, 32)}},
	}
}

func collectMetrics(t *testing.T, c *Collector) []*dto.MetricFamily {
	t.Helper()
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)
	families, err := reg.Gather()
	require.NoError(t, err)
	return families
}

func familyMap(families []*dto.MetricFamily) map[string]*dto.MetricFamily {
	m := make(map[string]*dto.MetricFamily, len(families))
	for _, f := range families {
		m[f.GetName()] = f
	}
	return m
}

func TestCollectSpecifiedInterfaces(t *testing.T) {
	peer1 := newTestPeer(1, 100, 200, time.Unix(1000, 0))
	peer2 := newTestPeer(2, 300, 400, time.Unix(2000, 0))

	mock := &mockDeviceLister{
		devices: []*wgtypes.Device{
			{Name: "wg0", Peers: []wgtypes.Peer{peer1}},
			{Name: "wg1", Peers: []wgtypes.Peer{peer2}},
		},
	}

	c := NewCollectorWithDevices([]string{"wg0"}, mock)
	families := collectMetrics(t, c)
	fm := familyMap(families)

	assert.Equal(t, 7, len(families))

	// Per-peer metrics should only contain wg0
	for _, name := range []string{
		"wireguard_latest_handshake_seconds",
		"wireguard_transmitted_bytes",
		"wireguard_received_bytes",
		"wireguard_peer_up",
	} {
		require.Contains(t, fm, name)
		assert.Equal(t, 1, len(fm[name].GetMetric()))
		for _, metric := range fm[name].GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "interface" {
					assert.Equal(t, "wg0", label.GetValue())
				}
			}
		}
	}

	// Interface info should only have wg0
	require.Contains(t, fm, "wireguard_interface_info")
	assert.Equal(t, 1, len(fm["wireguard_interface_info"].GetMetric()))

	// Scrape metrics always present
	require.Contains(t, fm, "wireguard_scrape_success")
	require.Contains(t, fm, "wireguard_scrape_duration_seconds")
}

func TestCollectAllInterfaces(t *testing.T) {
	peer1 := newTestPeer(1, 100, 200, time.Unix(1000, 0))
	peer2 := newTestPeer(2, 300, 400, time.Unix(2000, 0))

	mock := &mockDeviceLister{
		devices: []*wgtypes.Device{
			{Name: "wg0", Peers: []wgtypes.Peer{peer1}},
			{Name: "wg1", Peers: []wgtypes.Peer{peer2}},
		},
	}

	c := NewCollectorWithDevices(nil, mock)
	families := collectMetrics(t, c)
	fm := familyMap(families)

	assert.Equal(t, 7, len(families))

	// Per-peer metrics should have 2 entries (one per device/peer)
	for _, name := range []string{
		"wireguard_latest_handshake_seconds",
		"wireguard_transmitted_bytes",
		"wireguard_received_bytes",
		"wireguard_peer_up",
	} {
		require.Contains(t, fm, name)
		assert.Equal(t, 2, len(fm[name].GetMetric()))
	}

	// Interface info: 2 entries (one per device)
	require.Contains(t, fm, "wireguard_interface_info")
	assert.Equal(t, 2, len(fm["wireguard_interface_info"].GetMetric()))

	// Scrape metrics: 1 entry each
	assert.Equal(t, 1, len(fm["wireguard_scrape_success"].GetMetric()))
	assert.Equal(t, 1, len(fm["wireguard_scrape_duration_seconds"].GetMetric()))
}

func TestCollectDeviceError(t *testing.T) {
	mock := &mockDeviceLister{
		err: errors.New("wgctrl: permission denied"),
	}

	c := NewCollectorWithDevices(nil, mock)
	families := collectMetrics(t, c)
	fm := familyMap(families)

	// Only scrape metrics emitted on error
	assert.Equal(t, 2, len(families))
	require.Contains(t, fm, "wireguard_scrape_success")
	assert.Equal(t, 0.0, fm["wireguard_scrape_success"].GetMetric()[0].GetGauge().GetValue())
}

func TestCollectMetricValues(t *testing.T) {
	handshakeTime := time.Unix(1700000000, 0)
	peer := newTestPeer(1, 500, 1000, handshakeTime)

	mock := &mockDeviceLister{
		devices: []*wgtypes.Device{
			{Name: "wg0", Peers: []wgtypes.Peer{peer}},
		},
	}

	c := NewCollectorWithDevices(nil, mock)
	families := collectMetrics(t, c)

	values := make(map[string]float64)
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			values[family.GetName()] = metric.GetGauge().GetValue()
		}
	}

	assert.Equal(t, float64(1700000000), values["wireguard_latest_handshake_seconds"])
	assert.Equal(t, float64(500), values["wireguard_transmitted_bytes"])
	assert.Equal(t, float64(1000), values["wireguard_received_bytes"])
	assert.Equal(t, 0.0, values["wireguard_peer_up"]) // handshake is far in the past
	assert.Equal(t, 1.0, values["wireguard_interface_info"])
	assert.Equal(t, 1.0, values["wireguard_scrape_success"])
}

func TestPeerUpWithRecentHandshake(t *testing.T) {
	recentHandshake := time.Now().Add(-1 * time.Minute) // 1 minute ago
	peer := newTestPeer(1, 100, 200, recentHandshake)

	mock := &mockDeviceLister{
		devices: []*wgtypes.Device{
			{Name: "wg0", Peers: []wgtypes.Peer{peer}},
		},
	}

	c := NewCollectorWithDevices(nil, mock)
	families := collectMetrics(t, c)
	fm := familyMap(families)

	require.Contains(t, fm, "wireguard_peer_up")
	assert.Equal(t, 1.0, fm["wireguard_peer_up"].GetMetric()[0].GetGauge().GetValue())
}
