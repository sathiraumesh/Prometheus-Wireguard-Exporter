package wgprometheus

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

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

	registry *prometheus.Registry
)

func init() {
	registry = prometheus.NewRegistry()

	registry.MustRegister(
		wireguardLatestHandshake,
		wireguardTransmit,
		wireguardReceived,
	)
}

func GetRegistry() *prometheus.Registry {
	return registry
}

func ScrapConnectionStats(monitorKeys []string, scrapInterval time.Duration) {

	for {
		intfs, err := getInterfaces()

		if err != nil {
			log.Fatalf("failed to get interfaces from the device %v", err)
		}

		monitorIntfs := monitorInfterface(monitorKeys, intfs)

		for _, intf := range monitorIntfs {
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
		time.Sleep(time.Second * 5)
	}
}

func getInterfaces() (interfaces []*wgtypes.Device, err error) {
	wgClient, err := wgctrl.New()
	if err != nil {
		return
	}

	defer wgClient.Close()

	interfaces, err = wgClient.Devices()
	return
}

func monitorInfterface(monitorKeys []string, allIntf []*wgtypes.Device) (monitor []*wgtypes.Device) {

	if len(monitorKeys) == 0 {
		monitor = allIntf
		return
	}

	monitor = make([]*wgtypes.Device, 0)

	for _, key := range monitorKeys {
		for _, intf := range allIntf {
			if strings.TrimSpace(key) == intf.Name {
				monitor = append(monitor, intf)
			}
		}
	}

	return
}
