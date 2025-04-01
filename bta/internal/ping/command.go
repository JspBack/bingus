package ping

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}

func ping(ctx context.Context, host string, timeout time.Duration) (*PingResult, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %w", err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
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
		return nil, fmt.Errorf("error marshalling ICMP message: %w", err)
	}

	start := time.Now()
	if _, err := conn.WriteTo(msgBytes, ipAddr); err != nil {
		return nil, fmt.Errorf("error sending ICMP packet: %w", err)
	}

	select {
	case <-ctxWithTimeout.Done():
		return nil, ctxWithTimeout.Err()
	default:
	}

	conn.SetReadDeadline(time.Now().Add(timeout))

	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return nil, err
	}
	duration := time.Since(start)

	rm, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
		return nil, fmt.Errorf("error parsing ICMP message: %w", err)
	}
	if rm.Type == ipv4.ICMPTypeEchoReply {
		return &PingResult{IP: ipAddr.String(), RTT: duration}, nil
	}
	return nil, fmt.Errorf("unexpected ICMP message type: %v", rm.Type)
}

func hostDiscovery(timeout time.Duration, hostFoundCh chan string, maxHosts int) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var result []string
	var activeInterface net.Interface
	var found bool

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("error retrieving interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				activeInterface = iface
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no active network interface found")
	}

	var ipNet *net.IPNet
	addrs, err := activeInterface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("error retrieving addresses for interface %s: %w", activeInterface.Name, err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ipNet = ipnet
			break
		}
	}
	if ipNet == nil {
		return nil, fmt.Errorf("no IPv4 address found for interface %s", activeInterface.Name)
	}

	localIP := ipNet.IP.Mask(ipNet.Mask)
	mask := net.IP(ipNet.Mask).To4()
	ipUint := ipToUint32(localIP)
	maskUint := ipToUint32(mask)
	broadcastUint := ipUint | ^maskUint

	hostCount := int(broadcastUint - ipUint - 1)
	if maxHosts > 0 && maxHosts < hostCount {
		hostCount = maxHosts
	}

	var wg sync.WaitGroup
	resultsMutex := sync.Mutex{}

	concurrency := min(50, hostCount)

	sem := make(chan struct{}, concurrency)

	scanned := 0

	resultBuffer := make([]string, 0, hostCount)

	for candidate := ipUint + 1; candidate < broadcastUint && (maxHosts <= 0 || scanned < maxHosts); candidate++ {
		candidateIP := uint32ToIP(candidate).String()
		wg.Add(1)
		scanned++

		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			continue
		}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			if res, err := ping(ctx, ip, timeout); err == nil && res != nil {
				select {
				case hostFoundCh <- ip:
				default:
				}

				resultsMutex.Lock()
				resultBuffer = append(resultBuffer, ip)
				resultsMutex.Unlock()
			}
		}(candidateIP)
	}

	wg.Wait()
	close(sem)

	resultsMutex.Lock()
	result = append(result, resultBuffer...)
	resultsMutex.Unlock()

	return result, nil
}
