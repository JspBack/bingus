package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bingus",
		Short: "Bingus - A network scanning tool",
		Long: `Bingus is a command-line network scanning tool 
that allows you to discover hosts on your network and scan ports.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}

	rootCmd.AddCommand(NewPingCmd())
	rootCmd.AddCommand(NewPortCmd())
	rootCmd.AddCommand(NewHelpCmd())

	return rootCmd
}
