package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "watchdog",
	Short: "Simple agent watch dog to watch for HTTP objects and filesystem changes",
	Long: `This is a simple watchdog to monitor HTTP endpoints (as in REST-facing object stores) 
and filesystem directories. Once a change has been noted, the watchdog can then 
download an object from an HTTP source or upload a file to an HTTP target.`,
}

// Execute adds all child commands to the root command and sets flags appropriately. // This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
