package port

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

func scanPort(ctx context.Context, host string, port int, timeout time.Duration) PortResult {
	dialer := net.Dialer{Timeout: timeout}

	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return PortResult{Port: port, Open: false, Error: err}
	}

	defer conn.Close()

	return PortResult{Port: port, Open: true, Error: nil}
}

func portDiscovery(ctx context.Context, hosts []string, portsToScan []int, timeout time.Duration, portFoundCh chan PortResult) (map[string][]int, error) {
	results := make(map[string][]int)

	var hostWg sync.WaitGroup

	resultsMutex := sync.Mutex{}

	hostSem := make(chan struct{}, len(hosts))

	for _, host := range hosts {
		hostWg.Add(1)

		select {
		case hostSem <- struct{}{}:
		case <-ctx.Done():
			hostWg.Done()
			continue
		}

		go func(host string) {
			defer hostWg.Done()
			defer func() { <-hostSem }()

			var portWg sync.WaitGroup

			portSem := make(chan struct{}, 100)

			openPorts := make([]int, 0)
			openPortsMutex := sync.Mutex{}

			for _, port := range portsToScan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				portWg.Add(1)

				select {
				case portSem <- struct{}{}:
				case <-ctx.Done():
					portWg.Done()
					continue
				}

				go func(port int) {
					defer portWg.Done()
					defer func() { <-portSem }()

					result := scanPort(ctx, host, port, timeout)

					select {
					case portFoundCh <- result:
					default:
					}

					if result.Open {
						openPortsMutex.Lock()
						openPorts = append(openPorts, port)
						openPortsMutex.Unlock()
					}
				}(port)
			}

			portWg.Wait()
			close(portSem)

			resultsMutex.Lock()
			results[host] = openPorts
			resultsMutex.Unlock()

		}(host)
	}

	hostWg.Wait()
	close(hostSem)

	return results, nil
}
