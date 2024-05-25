package wgprometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestMonitorSpecifiedInterfaces(t *testing.T) {
	monitorKeys := []string{"wg1", "wg2"}
	wgDevice1 := wgtypes.Device{Name: "wg1"}

	wgDevices := []*wgtypes.Device{&wgDevice1}
	result := monitorInfterface(monitorKeys, wgDevices)

	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0].Name, "wg1")
}

func TestMonitorAllInterfaces(t *testing.T) {
	monitorKeys := []string{}
	wgDevice1 := wgtypes.Device{Name: "wg1"}
	wgDevice2 := wgtypes.Device{Name: "wg2"}

	wgDevices := []*wgtypes.Device{&wgDevice1, &wgDevice2}
	result := monitorInfterface(monitorKeys, wgDevices)

	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[0].Name, "wg1")
	assert.Equal(t, result[1].Name, "wg2")
}

func TestGetIntefacesIntergation(t *testing.T) {
	// look in the setup/wireguard/wg0_host_1.conf
	interfaces, _ := getInterfaces()

	assert.Equal(t, len(interfaces), 1)
	assert.Equal(t, interfaces[0].Name, "wg0")
	assert.Equal(t, len(interfaces[0].Peers), 1)
}
