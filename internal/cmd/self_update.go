package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

const (
	// Lock timeout - if lock is older than this, it's considered stale
	lockTimeout = 30 * time.Minute
)

var (
	// Repository information (configurable via environment variables)
	repoOwner = getEnvOrDefault("N0MAN_REPO_OWNER", "noeltz")
	repoName  = getEnvOrDefault("N0MAN_REPO_NAME", "n0man")
)

// GitHubRelease represents a GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update n0man to the latest version",
	Long: `Update n0man to the latest version from the configured repository.
	
This command will:
- Check for a newer version
- Download and build the latest version
- Replace the current binary atomically
- Keep a backup that can be restored if needed`,
	RunE: runSelfUpdate,
}

// UpdateError represents an error that occurred during the update process
type UpdateError struct {
	Step       string
	Err        error
	CanRecover bool
	Suggestion string
}

func (e *UpdateError) Error() string {
	return fmt.Sprintf("%s: %v", e.Step, e.Err)
}

func (e *UpdateError) Unwrap() error {
	return e.Err
}

// runSelfUpdate executes the self-update process
func runSelfUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Updating n0man...")
	fmt.Println()

	// Step 1: Acquire lock to prevent concurrent updates
	releaseLock, err := acquireUpdateLock()
	if err != nil {
		return fmt.Errorf("failed to acquire update lock: %w", err)
	}
	defer releaseLock()

	// Step 2: Detect current binary location
	currentPath, err := getExecutablePath()
	if err != nil {
		return &UpdateError{
			Step:       "Binary location detection",
			Err:        err,
			CanRecover: false,
			Suggestion: "Ensure n0man is installed correctly",
		}
	}
	fmt.Printf("  Current binary: %s\n", currentPath)
	fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Step 3: Check if Go is installed
	if err := checkGoInstalled(); err != nil {
		return &UpdateError{
			Step:       "Prerequisite check",
			Err:        err,
			CanRecover: false,
			Suggestion: "Install Go from https://go.dev/dl/",
		}
	}

	// Step 4: Check for updates
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		return &UpdateError{
			Step:       "Version check",
			Err:        err,
			CanRecover: true,
			Suggestion: "Check your network connection and try again",
		}
	}

	currentVersion := getCurrentVersion()
	fmt.Printf("  Current version: %s\n", currentVersion)
	fmt.Printf("  Latest version: %s\n", latestVersion)

	if currentVersion == latestVersion {
		fmt.Println()
		fmt.Println("✅ Already up to date!")
		return nil
	}

	// Step 5: Build new version to temporary location
	fmt.Println()
	fmt.Println("  Building new version...")
	tempPath, err := buildNewVersion()
	if err != nil {
		return &UpdateError{
			Step:       "Build",
			Err:        err,
			CanRecover: true,
			Suggestion: "Check your Go installation and try again",
		}
	}
	// Clean up temp file if we fail later
	defer func() {
		if _, err := os.Stat(tempPath); err == nil {
			os.Remove(tempPath)
		}
	}()

	// Step 6: Verify the new binary
	fmt.Println("  Verifying new binary...")
	if err := verifyBinary(tempPath); err != nil {
		return &UpdateError{
			Step:       "Verification",
			Err:        err,
			CanRecover: true,
			Suggestion: "The downloaded binary may be corrupted",
		}
	}

	// Step 7: Atomically replace the old binary
	fmt.Println("  Replacing binary...")
	if err := atomicReplace(currentPath, tempPath); err != nil {
		return &UpdateError{
			Step:       "Replacement",
			Err:        err,
			CanRecover: true,
			Suggestion: "Check file permissions and disk space",
		}
	}

	// Success!
	fmt.Println()
	fmt.Println("✅ n0man updated successfully!")
	fmt.Printf("  New version: %s\n", latestVersion)
	fmt.Println()
	fmt.Println("  Run 'n0man --help' to verify the update.")

	return nil
}

// getExecutablePath returns the path of the currently running binary
func getExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to determine executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(realPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// checkGoInstalled verifies that Go is installed and accessible
func checkGoInstalled() error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go is not installed or not in PATH")
	}
	fmt.Printf("  Using Go at: %s\n", goPath)

	// Check Go version
	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check Go version: %w", err)
	}
	fmt.Printf("  %s", strings.TrimSpace(string(output)))

	return nil
}

// getCurrentVersion returns the current version of n0man
func getCurrentVersion() string {
	// First, try with --short flag for simpler output
	cmd := exec.Command(os.Args[0], "version", "--short")
	output, err := cmd.CombinedOutput()
	if err == nil {
		// If --short flag is supported, output should be just the version
		version := strings.TrimSpace(string(output))
		if version != "" && version != "unknown" {
			return version
		}
	}

	// Fall back to parsing full output
	cmd = exec.Command(os.Args[0], "version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}

	// Parse "n0man version X.Y.Z" to extract version
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	return "unknown"
}

// fetchLatestVersion fetches the latest version from GitHub releases
func fetchLatestVersion() (string, error) {
	return fetchLatestVersionWithRetry(3)
}

// fetchLatestVersionWithRetry fetches the latest version with retry logic
func fetchLatestVersionWithRetry(maxRetries int) (string, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff: 1s, 2s, 4s, etc.
			backoff := time.Duration(1<<uint(i)) * time.Second
			fmt.Printf("  Retrying in %v...\n", backoff)
			time.Sleep(backoff)
		}

		version, err := doFetchLatestVersion()
		if err == nil {
			return version, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// doFetchLatestVersion performs the actual HTTP request to fetch the latest version
func doFetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	// Configure HTTP client with explicit TLS settings
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// buildNewVersion builds the latest version to a temporary location
func buildNewVersion() (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "n0man-build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	repoPath := filepath.Join(tempDir, "n0man")
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", repoOwner, repoName)

	ctx, cancel := context.WithTimeout(context.Background(), getUpdateTimeout())
	defer cancel()

	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, repoPath)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w\n%s", err, output)
	}

	// Build binary
	outputPath := filepath.Join(tempDir, "n0man")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, "./cmd/n0man")
	buildCmd.Dir = repoPath

	if output, err := buildCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build failed: %w\n%s", err, output)
	}

	return outputPath, nil
}

// verifyBinary verifies that a binary is valid and functional
func verifyBinary(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s", path)
	}

	// Check file size
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	minBinarySize, maxBinarySize := getBinarySizeThresholds()

	if info.Size() < minBinarySize {
		return fmt.Errorf("binary too small: %d bytes (minimum %d)", info.Size(), minBinarySize)
	}

	if info.Size() > maxBinarySize {
		return fmt.Errorf("binary too large: %d bytes (maximum %d)", info.Size(), maxBinarySize)
	}

	// Try to execute with --help to verify it's a valid Go binary
	cmd := exec.Command(path, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("binary is not functional: %w", err)
	}

	return nil
}

// atomicReplace atomically replaces the old binary with the new one
func atomicReplace(oldPath, newPath string) error {
	backupPath := oldPath + ".backup"

	// Step 1: Create backup of current binary
	if err := os.Rename(oldPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Step 2: Move new binary to target location
	if err := os.Rename(newPath, oldPath); err != nil {
		// Attempt to restore backup
		if restoreErr := os.Rename(backupPath, oldPath); restoreErr != nil {
			return fmt.Errorf("failed to replace binary and failed to restore backup: %w (restore error: %w)", err, restoreErr)
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Step 3: Set executable permissions
	if err := os.Chmod(oldPath, 0755); err != nil {
		// Attempt to restore backup
		if restoreErr := os.Rename(backupPath, oldPath); restoreErr != nil {
			return fmt.Errorf("failed to set permissions and failed to restore backup: %w (restore error: %w)", err, restoreErr)
		}
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Step 4: Remove backup after successful replacement
	if err := os.Remove(backupPath); err != nil {
		fmt.Printf("Warning: failed to remove backup file: %v\n", err)
	}

	return nil
}

// acquireUpdateLock acquires a lock file to prevent concurrent updates
func acquireUpdateLock() (releaseLock func(), err error) {
	lockFilePath := getLockFilePath()
	for {
		// Try to create exclusive lock file
		lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			if os.IsExist(err) {
				// Lock file exists, check if it's stale
				info, statErr := os.Stat(lockFilePath)
				if statErr == nil {
					age := time.Since(info.ModTime())
					if age > lockTimeout {
						// Lock is stale, remove it and retry
						os.Remove(lockFilePath)
						// Add small delay to prevent race condition
						time.Sleep(100 * time.Millisecond)
						continue // Retry instead of recursion
					}
				}
				return nil, fmt.Errorf("another update is in progress")
			}
			return nil, fmt.Errorf("failed to acquire lock: %w", err)
		}

		// Write PID and timestamp to lock file
		pid := os.Getpid()
		timestamp := time.Now().Format(time.RFC3339)
		lockContent := fmt.Sprintf("%d\n%s\n", pid, timestamp)

		if _, err := lockFile.WriteString(lockContent); err != nil {
			lockFile.Close()
			os.Remove(lockFilePath)
			return nil, fmt.Errorf("failed to write lock file: %w", err)
		}

		lockFile.Close()

		// Return cleanup function
		releaseLock = func() {
			if err := os.Remove(lockFilePath); err != nil {
				fmt.Printf("Warning: failed to remove lock file: %v\n", err)
			}
		}

		return releaseLock, nil
	}
}

// getLockFilePath returns the appropriate lock file path
// Checks XDG_RUNTIME_DIR first, falls back to os.TempDir()
func getLockFilePath() string {
	if xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntimeDir != "" {
		return filepath.Join(xdgRuntimeDir, "n0man-update.lock")
	}
	return filepath.Join(os.TempDir(), "n0man-update.lock")
}

// getUpdateTimeout returns the update timeout duration
// Reads from N0MAN_UPDATE_TIMEOUT env var, defaults to 5m
func getUpdateTimeout() time.Duration {
	if timeoutStr := os.Getenv("N0MAN_UPDATE_TIMEOUT"); timeoutStr != "" {
		if duration, err := time.ParseDuration(timeoutStr); err == nil {
			return duration
		}
	}
	return 5 * time.Minute
}

// getBinarySizeThresholds returns the min and max binary size thresholds
// Reads from N0MAN_MIN_BINARY_SIZE and N0MAN_MAX_BINARY_SIZE env vars
func getBinarySizeThresholds() (min, max int64) {
	min = 500_000     // 500KB default
	max = 200_000_000 // 200MB default

	if minStr := os.Getenv("N0MAN_MIN_BINARY_SIZE"); minStr != "" {
		if minVal, err := parseSize(minStr); err == nil {
			min = minVal
		}
	}

	if maxStr := os.Getenv("N0MAN_MAX_BINARY_SIZE"); maxStr != "" {
		if maxVal, err := parseSize(maxStr); err == nil {
			max = maxVal
		}
	}

	return min, max
}

// parseSize parses a size string (e.g., "1000000" or "1M") to bytes
func parseSize(s string) (int64, error) {
	// Try to parse as integer first
	if size, err := strconv.ParseInt(s, 10, 64); err == nil {
		return size, nil
	}
	return 0, fmt.Errorf("invalid size: %s", s)
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
}
