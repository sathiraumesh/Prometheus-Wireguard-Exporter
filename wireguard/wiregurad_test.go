package wireguard

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestConnectionsEmpty(t *testing.T) {
	testCases := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{`something random test 1
			something random test 2

			something test 3`, 0},
	}

	for _, tc := range testCases {
		conn := ListConnections(tc.input)
		if len(conn) != tc.expected {
			t.Error("wrong number of peers", len(conn))
		}
	}
}

func TestConnectionsNotEmpty(t *testing.T) {
	raw_output := `
		interface: wg0
		public key: YexUX3CRfPHSt7DYKV5gnRJWd8hDNkE2QxHIMZa5eEg=
		private key: (hidden)
		listening port: 51820

		peer: HYf+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSxxI=
			endpoint: 172.28.0.2:51820
			allowed ips: 10.8.0.1/32
			latest handshake: 43 seconds ago
			transfer: 180 B received, 400 B sent
			persistent keepalive: every 10 seconds
	
		peer: KY+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSxxI=
			endpoint: 172.28.0.2:51820
			allowed ips: 10.8.0.2/32
			latest handshake: 43 seconds ago
			transfer: 180 B received, 400 B sent
			persistent keepalive: every 10 seconds

		interface: wg1
		public key: YexUX3CRfPHSt7DYKV5gnRJWd8hDNkE2QxHIMZa5eEg=
		private key: (hidden)
		listening port: 51820
		
		peer: HYf+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSjkls
			endpoint: 172.28.0.2:51820
			allowed ips: 10.8.28.1/32
			latest handshake: 43 seconds ago
			transfer: 180 B received, 400 B sent
			persistent keepalive: every 10 seconds
			`

	conn := ListConnections(raw_output)
	if len(conn) != 3 {
		t.Error("wrong number of peers", len(conn))
	}

	expectedCons := []struct {
		interfaceName string
		publicKey     string
		allowedIps    string
	}{
		{"wg0", "HYf+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSxxI=", "10.8.0.1/32"},
		{"wg0", "KY+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSxxI=", "10.8.0.2/32"},
		{"wg1", "HYf+yNzgj3uhARFlNy3Pawuk/yLC+WYoY2qwjjlSjkls", "10.8.28.1/32"},
	}

	for _, exp := range expectedCons {
		con := conn[exp.publicKey]
		assertEqual(t, con.Interface, exp.interfaceName)
		assertEqual(t, con.AllowedIps, exp.allowedIps)

	}
}

func TestParseTime(t *testing.T) {
	currentTime := time.Now().Unix()

	testCases := []struct {
		input    string
		expected int
	}{
		{"    ", math.MaxInt},
		{"43 seconds ago", int(currentTime) - 43},
		{"1 minute, 59 seconds ago", int(currentTime) - 119},
		{"3 hours, 44 minutes, 51 seconds ago", int(currentTime) - 13491},
		{"7 days, 20 hours, 30 minutes, 15 seconds ago", int(currentTime) - 678615},
	}

	for _, tc := range testCases {
		secs, _ := parseTime(tc.input)
		assertDelta(t, tc.expected, secs, 100)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected: %v, but got: %v", expected, actual)
	}
}

func assertDelta(t *testing.T, expected, actual, delta int) {
	diff := expected - actual
	if math.Abs(float64(diff)) > float64(delta) {
		t.Errorf("Expected delta less than %v , but got: %v", delta, diff)
	}
}
