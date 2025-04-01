package port

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/jspback/bingus/cobra/internal/util"
)

func scanPort(ctx context.Context, host string, port int, timeout time.Duration) PortResult {
	logger := util.NewVerboseLogger(ctx)
	logger.Print("Scanning port %d on host %s (timeout: %v)...\n", port, host, timeout)

	dialer := net.Dialer{Timeout: timeout}
	address := fmt.Sprintf("%s:%d", host, port)

	startTime := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", address)

	if err != nil {
		logger.Print("Port %d on host %s is closed: %v (in %v)\n", port, host, err, time.Since(startTime))
		return PortResult{Host: host, Port: port, Open: false, Error: err}
	}

	logger.Print("Port %d on host %s is OPEN (connected in %v)\n", port, host, time.Since(startTime))
	defer conn.Close()
	return PortResult{Host: host, Port: port, Open: true, Error: nil}
}

func PortDiscovery(ctx context.Context, hosts []string, portsToScan []int, timeout time.Duration, portFoundCh chan PortResult) (map[string][]int, error) {
	logger := util.NewVerboseLogger(ctx)

	results := make(map[string][]int)
	var resultsMutex sync.Mutex

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	maxHostConcurrency := min(len(hosts), 50)
	logger.Print("Starting port discovery with %d hosts and %d ports\n", len(hosts), len(portsToScan))
	logger.Print("Using max host concurrency of %d\n", maxHostConcurrency)

	hostLimiter := util.NewConcurrencyLimiter(ctx, maxHostConcurrency)
	defer hostLimiter.Close()

	for _, host := range hosts {
		host := host
		if err := hostLimiter.Execute(func() {
			logger.Print("Starting scan for host %s (%d ports)\n", host, len(portsToScan))

			resultsMutex.Lock()
			results[host] = []int{}
			resultsMutex.Unlock()

			maxPortConcurrency := 100
			logger.Print("Using max port concurrency of %d for host %s\n", maxPortConcurrency, host)

			portLimiter := util.NewConcurrencyLimiter(ctx, maxPortConcurrency)
			defer portLimiter.Close()

			var openPortsMutex sync.Mutex

			for _, port := range portsToScan {
				port := port

				if err := portLimiter.Execute(func() {
					result := scanPort(ctx, host, port, timeout)

					select {
					case <-ctx.Done():
						return
					case portFoundCh <- result:
					default:
						logger.Print("Warning: result channel full, skipped reporting port %d on %s\n", port, host)
					}

					if result.Open {
						openPortsMutex.Lock()
						resultsMutex.Lock()
						results[host] = append(results[host], port)
						resultsMutex.Unlock()
						openPortsMutex.Unlock()
					}
				}); err != nil {
					return
				}
			}

			logger.Print("Waiting for all port scans to complete for host %s\n", host)
			portLimiter.Wait()

			resultsMutex.Lock()
			openPorts := results[host]
			resultsMutex.Unlock()
			logger.Print("Scan complete for host %s: found %d open ports\n", host, len(openPorts))
		}); err != nil {
			return results, err
		}
	}

	logger.Print("Waiting for all host scans to complete...\n")
	hostLimiter.Wait()
	logger.Print("Port discovery complete.\n")

	return results, nil
}
