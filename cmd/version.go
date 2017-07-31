package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information on watchdog
const Version string = "0.0.1"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns version information.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🐶")
		fmt.Println("Watchdog v" + Version)
	},
}
