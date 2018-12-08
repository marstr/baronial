package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	version string
	revision string
)

var versionCmd = &cobra.Command{
	Use: "version",
	Short: "Prints version information about ledger.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: ", version)
		fmt.Printf("System: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Go: ", runtime.Version())
		fmt.Println("Source Revision: ", revision)
	},
	Args: cobra.NoArgs,
}

func init() {
	if version == "" {
		version = "unknown"
	}

	if revision == "" {
		revision = "unknown"
	}

	rootCmd.AddCommand(versionCmd)
}