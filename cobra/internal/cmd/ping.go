package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jspback/bingus/cobra/internal/ping"
	"github.com/jspback/bingus/cobra/internal/util"
	"github.com/spf13/cobra"
)

func NewPingCmd() *cobra.Command {
	var timeout time.Duration
	var maxHosts int
	var verbose bool

	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Scan for hosts on your network",
		Long:  `Scan for hosts on your network using ICMP echo requests`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Scanning for hosts on the network...")

			ctx := context.WithValue(context.Background(), "verbose", verbose)
			logger := util.NewVerboseLogger(ctx)

			logger.Print("Using timeout of %v per host\n", timeout)
			logger.Print("Maximum hosts to scan: %d\n", maxHosts)
			logger.Print("Starting scan at %v\n", time.Now().Format(time.RFC3339))

			hostFoundCh := make(chan string)
			go func() {
				for host := range hostFoundCh {
					fmt.Printf("Host found: %s\n", host)
				}
			}()

			hosts, err := ping.HostDiscovery(ctx, timeout, hostFoundCh, maxHosts)
			if err != nil {
				return fmt.Errorf("error during host discovery: %w", err)
			}

			close(hostFoundCh)
			logger.Print("Scan completed at %v\n", time.Now().Format(time.RFC3339))

			fmt.Printf("\nScan complete. Found %d hosts on the network.\n", len(hosts))
			for i, host := range hosts {
				fmt.Printf("%d. %s\n", i+1, host)
			}

			return nil
		},
	}

	pingCmd.Flags().DurationVarP(&timeout, "timeout", "t", 500*time.Millisecond, "Timeout for each host ping (default: 500ms)")
	pingCmd.Flags().IntVarP(&maxHosts, "max-hosts", "m", 50, "Maximum number of hosts to scan (default: 50)")
	pingCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return pingCmd
}
