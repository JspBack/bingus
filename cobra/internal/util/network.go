package util

import (
	"encoding/binary"
	"fmt"
	"net"
)

func IPToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func Uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}

func GetIPNetForActiveInterface(logger *VerboseLogger) (net.Interface, *net.IPNet, error) {
	logger.Print("Discovering network interfaces...\n")

	interfaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, nil, fmt.Errorf("error retrieving interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			logger.Print("Skipping interface %s (flags: %v)\n", iface.Name, iface.Flags)
			continue
		}

		logger.Print("Checking interface %s (flags: %v)\n", iface.Name, iface.Flags)

		addrs, err := iface.Addrs()
		if err != nil {
			logger.Print("Error getting addresses for interface %s: %v\n", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				logger.Print("Selected active interface: %s with IPv4 address: %s\n", iface.Name, ipnet.IP)
				return iface, ipnet, nil
			}
		}
	}

	return net.Interface{}, nil, fmt.Errorf("no active network interface found")
}

const MaxIPsFromCIDR = 1000

func ParseCIDR(cidr string, logger *VerboseLogger) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ones, bits := ipNet.Mask.Size()
	totalIPs := 1 << (bits - ones)

	if totalIPs > MaxIPsFromCIDR {
		return nil, fmt.Errorf("CIDR range %s contains %d addresses, which exceeds the maximum of %d. Please use a smaller range",
			cidr, totalIPs, MaxIPsFromCIDR)
	}

	var ips []string
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	logger.Print("CIDR %s contains %d usable IP addresses\n", cidr, len(ips))
	return ips, nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
