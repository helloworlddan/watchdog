package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	downloadDirectory, downloadURL string
	watchInterval                  int
)

func init() {
	RootCmd.AddCommand(httpCmd)
	httpCmd.Flags().StringVarP(&downloadDirectory, "directory", "d", ".", "Directory to download to.")
	httpCmd.Flags().StringVarP(&downloadURL, "url", "u", "http://localhost/{extension}/{filename}", "URL to download from.")
	httpCmd.Flags().IntVarP(&watchInterval, "interval", "i", 60, "Watch interval in seconds.")
}

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Watch an URL and download to a filesystem directory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🐶")
	},
}
