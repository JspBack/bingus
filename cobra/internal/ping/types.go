package ping

import (
	"time"
)

type PingResult struct {
	IP  string
	RTT time.Duration
}
