package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"golang.zx2c4.com/wireguard/wgctrl"
)

const SCRAP_INTERVAL = 5

var (
	wireguardLatestHandshake = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_latest_handshake_seconds",
			Help: "The latest handshake time for Wireguard connections.",
		},
		[]string{"interface", "public_key", "allowed_ips"},
	)

	wireguardTransmit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_transmitted_bytes",
			Help: "Transmitted data for Wireguard connections.",
		},
		[]string{"interface", "public_key", "allowed_ips"},
	)

	wireguardReceived = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wireguard_received_bytes",
			Help: "Received data for Wireguard connections.",
		},
		[]string{"interface", "public_key", "allowed_ips"},
	)

	registry = prometheus.NewRegistry()

	portPtr  = flag.Int("p", 9011, "the port to listen on")
	itemsStr = flag.String("i", "", "comma-separated list of interfaces")
)

func init() {
	registry.MustRegister(
		wireguardLatestHandshake,
		wireguardTransmit,
		wireguardReceived,
	)
}

func main() {
	flag.Parse()

	port := ":" + strconv.Itoa(*portPtr)
	fmt.Println(port)
	interfaces := strings.Split(*itemsStr, ",")

	client, err := wgctrl.New()
	if err != nil {
		log.Fatalf("Failed to create WireGuard client: %v\n", err)
	}

	defer client.Close()

	go scrapeConnectionStats(client, interfaces)

	http.Handle("/metrics", promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	))

	http.ListenAndServe(port, nil)
}

func scrapeConnectionStats(client *wgctrl.Client, intfTM []string) {

	for {

		interfaces, err := client.Devices()

		if err != nil {
			log.Fatalf("Failed to get interfaces %v", err)
		}

		for _, intf := range interfaces {

			shouldMonitorInF := false

			for _, intfToMonitor := range intfTM {
				if intfToMonitor == intf.Name {
					shouldMonitorInF = true
				}
			}

			if shouldMonitorInF {
				continue
			}

			for _, peer := range intf.Peers {
				wireguardLatestHandshake.WithLabelValues(
					intf.Name,
					peer.PublicKey.String(),
					fmt.Sprintf("%v", peer.AllowedIPs),
				).Set(float64(peer.LastHandshakeTime.Unix()))

				wireguardTransmit.WithLabelValues(
					intf.Name,
					peer.PublicKey.String(),
					fmt.Sprintf("%v", peer.AllowedIPs),
				).Set(float64(peer.TransmitBytes))

				wireguardReceived.WithLabelValues(
					intf.Name,
					peer.PublicKey.String(),
					fmt.Sprintf("%v", peer.AllowedIPs),
				).Set(float64(peer.ReceiveBytes))
			}
		}

		time.Sleep(time.Second * time.Duration(SCRAP_INTERVAL))
	}
}
