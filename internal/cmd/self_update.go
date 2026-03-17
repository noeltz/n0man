package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update n0man to the latest version from the configured repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Strategy: rebuild from source using `go install`
		// This works robustly even if the current binary is running

		fmt.Println("🔄 Updating n0man...")

		// Check if Go is installed
		goPath, err := exec.LookPath("go")
		if err != nil {
			return fmt.Errorf("Go is not installed or not in PATH. Cannot self-update.\nInstall Go from https://go.dev/dl/ or update manually")
		}
		fmt.Printf("  Using Go at: %s\n", goPath)

		// Get current executable path for reference
		execPath, err := os.Executable()
		if err != nil {
			execPath = "unknown"
		}
		fmt.Printf("  Current binary: %s\n", execPath)
		fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		// Create context with timeout (5 minutes)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Use go install to fetch and build the latest version
		fmt.Println("  Fetching and building latest version (this may take a minute)...")
		installCmd := exec.CommandContext(ctx, "go", "install", "github.com/noeltz/n0man/cmd/n0man@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("self-update timed out after 5 minutes. Try again or update manually:\n  go install github.com/noeltz/n0man/cmd/n0man@latest")
			}
			return fmt.Errorf("self-update failed: %w", err)
		}

		// Determine where go install put the new binary
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				homeDir, _ := os.UserHomeDir()
				gopath = filepath.Join(homeDir, "go")
			}
			gobin = filepath.Join(gopath, "bin")
		}

		newBinary := filepath.Join(gobin, "n0man")

		// Verify the new binary exists
		if _, err := os.Stat(newBinary); os.IsNotExist(err) {
			return fmt.Errorf("self-update completed but new binary not found at: %s", newBinary)
		}

		fmt.Printf("\n✅ n0man updated successfully!\n")
		fmt.Printf("  New binary at: %s\n", newBinary)
		if execPath != "unknown" && execPath != newBinary {
			fmt.Println("\n  ⚠️  Note: The new binary is in a different location.")
			fmt.Printf("  To use the updated version, either:\n")
			fmt.Printf("    1. Move it: sudo mv %s %s\n", newBinary, execPath)
			fmt.Printf("    2. Or update your PATH to include: %s\n", gobin)
		}
		fmt.Println("  Run 'n0man version' to verify the update.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
}
