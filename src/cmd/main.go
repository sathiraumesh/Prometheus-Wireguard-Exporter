package main

import (
	"flag"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sathiraumesh/wireguard_exporter/wgprometheus"
)

const SCRAP_INTERVAL = time.Second * 5
const DEFALULT_PORT = 9011

var port = flag.Int("p", DEFALULT_PORT, "the port to listen on")
var interfaces = flag.String("i", "", "comma-separated list of interfaces")

func main() {

	flag.Parse()

	interfaces, port, _ := validateReturnFlags(*interfaces, *port)

	registry := wgprometheus.GetRegistry()

	go wgprometheus.ScrapConnectionStats(interfaces, SCRAP_INTERVAL)
	http.Handle("/metrics", promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{},
	))

	http.ListenAndServe(port, nil)
}

func validateReturnFlags(interfaceArg string, portArg int) (interfaces []string, port string, configPath string) {

	port = ":" + strconv.Itoa(portArg)

	if strings.TrimSpace(interfaceArg) != "" {
		interfaces = strings.Split(interfaceArg, ",")
	}

	return
}
