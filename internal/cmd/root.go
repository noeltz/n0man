package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information set at build time
var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// rootContext holds the global context for cancellation support
var rootContext = context.Background()

// SetContext sets the global context for command cancellation
func SetContext(ctx context.Context) {
	rootContext = ctx
}

// GetContext returns the global context
func GetContext() context.Context {
	return rootContext
}

var rootCmd = &cobra.Command{
	Use:     "n0man",
	Short:   "n0man is a safe and simple dotfiles manager",
	Long:    `n0man provides real bidirectional synchronization, backup, and security for your dotfiles.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// VersionCmd prints version information
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("n0man version %s\n", version)
		fmt.Printf("  Go version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		if gitCommit != "unknown" {
			fmt.Printf("  Git commit: %s\n", gitCommit)
		}
		if buildTime != "unknown" {
			fmt.Printf("  Built: %s\n", buildTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(VersionCmd)
}
