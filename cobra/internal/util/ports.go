package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePortRange(portsFlag string, logger *VerboseLogger) ([]int, error) {
	var portsToScan []int
	portRanges := strings.SplitSeq(portsFlag, ",")

	for portRange := range portRanges {
		portRange = strings.TrimSpace(portRange)
		if strings.Contains(portRange, "-") {
			rangeParts := strings.Split(portRange, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", portRange)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %s", rangeParts[1])
			}

			logger.Print("Adding port range %d-%d (%d ports)\n", start, end, end-start+1)
			for p := start; p <= end; p++ {
				portsToScan = append(portsToScan, p)
			}
		} else {
			port, err := strconv.Atoi(portRange)
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %s", portRange)
			}
			portsToScan = append(portsToScan, port)
		}
	}

	return portsToScan, nil
}
