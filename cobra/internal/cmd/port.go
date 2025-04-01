package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jspback/bingus/cobra/internal/port"
	"github.com/jspback/bingus/cobra/internal/util"
	"github.com/spf13/cobra"
)

func NewPortCmd() *cobra.Command {
	var timeout time.Duration
	var hostsFlag []string
	var portsFlag string
	var verbose bool

	portCmd := &cobra.Command{
		Use:   "port",
		Short: "Scan for open ports on specified hosts",
		Long:  `Scan for open ports on specified hosts using TCP connections`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(hostsFlag) == 0 {
				return fmt.Errorf("at least one host must be specified")
			}

			ctx := context.WithValue(context.Background(), "verbose", verbose)
			logger := util.NewVerboseLogger(ctx)

			var hosts []string
			for _, host := range hostsFlag {
				if strings.Contains(host, "/") {
					logger.Print("Parsing CIDR range: %s\n", host)
					cidrs, err := util.ParseCIDR(host, logger)
					if err != nil {
						return fmt.Errorf("invalid CIDR notation: %s: %w", host, err)
					}
					logger.Print("Resolved %s to %d IP addresses\n", host, len(cidrs))
					hosts = append(hosts, cidrs...)
				} else {
					hosts = append(hosts, host)
				}
			}

			portsToScan, err := util.ParsePortRange(portsFlag, logger)
			if err != nil {
				return err
			}

			fmt.Printf("Scanning %d ports on %d hosts (%d total port scans)...\n",
				len(portsToScan), len(hosts), len(hosts)*len(portsToScan))

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			bufferSize := 100
			if len(hosts)*len(portsToScan) < bufferSize {
				bufferSize = len(hosts) * len(portsToScan)
			}
			portFoundCh := make(chan port.PortResult, bufferSize)

			done := make(chan struct{})
			go func() {
				defer close(done)
				for result := range portFoundCh {
					if result.Open {
						fmt.Printf("Found open port %d on host %s\n", result.Port, result.Host)
					} else if verbose {
						fmt.Printf("Port %d on host %s is closed: %v\n", result.Port, result.Host, result.Error)
					}
				}
			}()

			if timeout == 0 {
				timeout = 500 * time.Millisecond
			}

			logger.Print("Using timeout of %v per connection\n", timeout)
			logger.Print("Starting scan at %v\n", time.Now().Format(time.RFC3339))

			results, err := port.PortDiscovery(ctx, hosts, portsToScan, timeout, portFoundCh)
			if err != nil {
				return fmt.Errorf("error during port discovery: %w", err)
			}

			close(portFoundCh)
			<-done

			logger.Print("Scan completed at %v\n", time.Now().Format(time.RFC3339))

			fmt.Println("\nScan complete. Found open ports:")
			openHostCount := 0
			for host, openPorts := range results {
				if len(openPorts) > 0 {
					openHostCount++
					fmt.Printf("%s: %v\n", host, openPorts)
				} else if verbose {
					fmt.Printf("%s: No open ports found\n", host)
				}
			}

			if openHostCount == 0 {
				fmt.Println("No open ports found on any hosts")
			}

			return nil
		},
	}

	portCmd.Flags().DurationVarP(&timeout, "timeout", "t", 500*time.Millisecond, "Timeout for each port scan")
	portCmd.Flags().StringSliceVarP(&hostsFlag, "hosts", "H", []string{}, "Hosts to scan (comma-separated, CIDR notation supported, e.g., 192.168.1.0/24)")
	portCmd.Flags().StringVarP(&portsFlag, "ports", "p", "21,22,23,25,53,80,110,139,143,443,445,993,995,3306,3389,5900,8080", "Ports to scan (comma-separated, ranges allowed e.g. 80-100)")
	portCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	portCmd.MarkFlagRequired("hosts")

	return portCmd
}
