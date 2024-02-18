package wireguard

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	SEPERATOR                    = ":"
	INTERFACE_PLACEHOLDER        = "interface"
	PEER_PLACHOLDER              = "peer"
	LATEST_HANDSHAKE_PLACEHOLDER = "lastest handshake"
	ALLOWED_IPS_PLACEHOLDER      = "allowed ips"
	TRANSFER_PLACEHOLDER         = "transfer"

	/* time in seocnds */
	MINUTE = 60
	HOUR   = 60 * MINUTE
	DAY    = 24 * HOUR
	YEAR   = 365 * DAY

	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
)

var timeFactors = [5]int{1, MINUTE, HOUR, DAY, YEAR}

type Connection struct {
	Interface        string
	LatestHandshake  int
	TransferReceived *Transfer
	TransferSent     *Transfer
	AllowedIps       string
}

type Connections map[string]*Connection

type Transfer struct {
	Unit string
	Size float64
}

func ListConnections(data string) Connections {
	conn := make(Connections)

	lines := strings.Split(data, "\n")

	var currentConnectionKey string
	var currentInteface string

	for _, line := range lines {
		lineFormated := strings.TrimSpace(line)

		if lineFormated == "" {
			continue
		}

		if strings.HasPrefix(strings.ToLower(lineFormated), INTERFACE_PLACEHOLDER+SEPERATOR) {
			parts := strings.Split(lineFormated, ":")
			currentInteface = strings.TrimSpace(parts[1])
			continue
		}

		if strings.HasPrefix(strings.ToLower(lineFormated), PEER_PLACHOLDER+SEPERATOR) {
			parts := strings.Fields(lineFormated)
			currentConnectionKey = strings.TrimSpace(parts[1])

			peer := Connection{Interface: currentInteface, LatestHandshake: math.MaxInt}
			conn[currentConnectionKey] = &peer
			continue
		}

		if currentConnectionKey != "" {

			parts := strings.Split(lineFormated, SEPERATOR)
			field := strings.ToLower(strings.TrimSpace(parts[0]))
			entry := parts[1]
			currentConn := conn[currentConnectionKey]

			switch field {
			case LATEST_HANDSHAKE_PLACEHOLDER:
				lastHandShakeEpochs, err := parseTime(entry)
				if err != nil {
					return nil
				}
				currentConn.LatestHandshake = lastHandShakeEpochs
			case TRANSFER_PLACEHOLDER:
				recived, sent, err := parseDataExchanged(entry)
				if err != nil {
					return nil
				}
				currentConn.TransferReceived = recived
				currentConn.TransferSent = sent
			case ALLOWED_IPS_PLACEHOLDER:
				currentConn.AllowedIps = strings.TrimSpace(entry)
			default:
				// do nothing
			}
		}
	}
	return conn
}

func parseTime(input string) (int, error) {

	formattedInput := strings.ToLower(strings.TrimSpace(input))

	if formattedInput == "" {
		return math.MaxInt, nil
	}

	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindAllString(formattedInput, -1)

	if len(matches) <= 5 {
		seconds := 0
		timeFactorIndex := 0

		for i := len(matches) - 1; i >= 0; i-- {
			num, err := strconv.Atoi(matches[i])

			if err != nil {
				return math.MaxInt, err
			}
			seconds += num * timeFactors[timeFactorIndex]
			timeFactorIndex++
		}

		currentTime := time.Now().Add(-time.Duration(seconds))
		lastHandShakeEpochs := int(currentTime.Unix()) - seconds
		return lastHandShakeEpochs, nil
	}
	return math.MaxInt, errors.New("invalid input string parse time")
}

func parseDataExchanged(input string) (*Transfer, *Transfer, error) {
	var dataRecived, dataSent float64
	var dataRecivedUnit, dataSentUnit string
	formattedInput := strings.ToLower(strings.TrimSpace(input))

	_, err := fmt.Sscanf(formattedInput, "%f %s received, %f %s sent",
		&dataRecived, &dataRecivedUnit, &dataSent, &dataSentUnit)
	if err != nil {
		return nil, nil, err
	}

	tarnsferRecived := Transfer{dataRecivedUnit, dataRecived}
	tarnsferSent := Transfer{dataRecivedUnit, dataRecived}

	return &tarnsferRecived, &tarnsferSent, nil
}
