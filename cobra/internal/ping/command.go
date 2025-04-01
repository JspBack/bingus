package ping

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jspback/bingus/cobra/internal/util"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func ping(ctx context.Context, host string, timeout time.Duration) (*PingResult, error) {
	logger := util.NewVerboseLogger(ctx)
	logger.Print("Pinging host %s (timeout: %v)...\n", host, timeout)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		logger.Print("Failed to resolve host %s: %v\n", host, err)
		return nil, fmt.Errorf("failed to resolve host: %w", err)
	}

	logger.Print("Resolved %s to %s\n", host, ipAddr.String())

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		logger.Print("Error listening for ICMP packets: %v\n", err)
		return nil, fmt.Errorf("error listening for ICMP packets: %w", err)
	}
	defer conn.Close()

	id := os.Getpid() & 0xffff
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  1,
			Data: []byte("PING"),
		},
	}
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		logger.Print("Error marshalling ICMP message: %v\n", err)
		return nil, fmt.Errorf("error marshalling ICMP message: %w", err)
	}

	start := time.Now()
	if _, err := conn.WriteTo(msgBytes, ipAddr); err != nil {
		logger.Print("Error sending ICMP packet to %s: %v\n", ipAddr.String(), err)
		return nil, fmt.Errorf("error sending ICMP packet: %w", err)
	}

	logger.Print("ICMP packet sent to %s, waiting for reply...\n", ipAddr.String())

	select {
	case <-ctxWithTimeout.Done():
		logger.Print("Timeout waiting for response from %s\n", ipAddr.String())
		return nil, ctxWithTimeout.Err()
	default:
	}

	conn.SetReadDeadline(time.Now().Add(timeout))

	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		logger.Print("Error reading response from %s: %v\n", ipAddr.String(), err)
		return nil, err
	}
	duration := time.Since(start)

	logger.Print("Received reply from %s in %v\n", ipAddr.String(), duration)

	rm, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
		logger.Print("Error parsing ICMP message from %s: %v\n", ipAddr.String(), err)
		return nil, fmt.Errorf("error parsing ICMP message: %w", err)
	}
	if rm.Type == ipv4.ICMPTypeEchoReply {
		return &PingResult{IP: ipAddr.String(), RTT: duration}, nil
	}

	logger.Print("Unexpected ICMP message type from %s: %v\n", ipAddr.String(), rm.Type)
	return nil, fmt.Errorf("unexpected ICMP message type: %v", rm.Type)
}

func HostDiscovery(ctx context.Context, timeout time.Duration, hostFoundCh chan string, maxHosts int) ([]string, error) {
	logger := util.NewVerboseLogger(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var result []string

	_, ipNet, err := util.GetIPNetForActiveInterface(logger)
	if err != nil {
		return nil, err
	}

	localIP := ipNet.IP.Mask(ipNet.Mask)
	mask := net.IP(ipNet.Mask).To4()
	ipUint := util.IPToUint32(localIP)
	maskUint := util.IPToUint32(mask)
	broadcastUint := ipUint | ^maskUint

	hostCount := int(broadcastUint - ipUint - 1)
	if maxHosts > 0 && maxHosts < hostCount {
		hostCount = maxHosts
	}

	logger.Print("Network information:\n")
	logger.Print("  Local IP: %s\n", localIP)
	logger.Print("  Netmask: %s\n", mask)
	logger.Print("  Broadcast: %s\n", util.Uint32ToIP(broadcastUint))
	logger.Print("  Host count: %d (limited to %d)\n", broadcastUint-ipUint-1, hostCount)

	concurrency := min(50, hostCount)
	logger.Print("Using concurrency of %d\n", concurrency)

	limiter := util.NewConcurrencyLimiter(ctx, concurrency)
	defer limiter.Close()

	var resultsMutex sync.Mutex
	resultBuffer := make([]string, 0, hostCount)
	scanned := 0

	logger.Print("Starting host scan from %s to %s\n", util.Uint32ToIP(ipUint+1), util.Uint32ToIP(broadcastUint-1))

	for candidate := ipUint + 1; candidate < broadcastUint && (maxHosts <= 0 || scanned < maxHosts); candidate++ {
		candidateIP := util.Uint32ToIP(candidate).String()
		scanned++

		if err := limiter.Execute(func() {
			if res, err := ping(ctx, candidateIP, timeout); err == nil && res != nil {
				select {
				case hostFoundCh <- candidateIP:
				default:
				}

				resultsMutex.Lock()
				resultBuffer = append(resultBuffer, candidateIP)
				resultsMutex.Unlock()
			} else {
				logger.Print("Host %s is not reachable: %v\n", candidateIP, err)
			}
		}); err != nil {
			break
		}
	}

	logger.Print("Waiting for all ping operations to complete...\n")
	limiter.Wait()

	resultsMutex.Lock()
	result = append(result, resultBuffer...)
	resultsMutex.Unlock()

	logger.Print("Host discovery complete, found %d active hosts\n", len(result))

	return result, nil
}
