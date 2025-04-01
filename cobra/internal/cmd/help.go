package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewHelpCmd() *cobra.Command {
	helpCmd := &cobra.Command{
		Use:   "help",
		Short: "Display help information about Bingus",
		Long:  `Display detailed help information about Bingus and its commands`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(`Bingus - Network Scanner Tool

COMMANDS:
  ping        Scan for hosts on your network using ICMP echo requests
  port        Scan for open ports on specified hosts using TCP connections
  help        Display this help information

EXAMPLES:
  # Scan for hosts on the network
  bingus ping --timeout 500ms --max-hosts 100

  # Scan specific ports on a host
  bingus port --hosts 192.168.1.1 --ports 80,443,8080

  # Scan a range of ports
  bingus port --hosts 192.168.1.1 --ports 1-1000

  # Scan multiple hosts
  bingus port --hosts 192.168.1.1,192.168.1.2 --ports 22,80,443

For more details on each command, use:
  bingus [command] --help`)
		},
	}

	return helpCmd
}
