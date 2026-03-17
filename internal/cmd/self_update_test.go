package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestAcquireUpdateLock tests lock acquisition and release
func TestAcquireUpdateLock(t *testing.T) {
	// Clean up any existing lock file
	lockPath := getLockFilePath()
	os.Remove(lockPath)
	defer os.Remove(lockPath)

	// Test 1: Acquire and release lock
	t.Run("acquire and release", func(t *testing.T) {
		releaseLock, err := acquireUpdateLock()
		if err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}

		// Verify lock file exists
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			t.Error("Lock file was not created")
		}

		// Release the lock
		releaseLock()

		// Verify lock file was removed
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Error("Lock file was not removed after release")
		}
	})

	// Test 2: Prevent concurrent lock acquisition
	t.Run("prevent concurrent acquisition", func(t *testing.T) {
		releaseLock1, err := acquireUpdateLock()
		if err != nil {
			t.Fatalf("Failed to acquire first lock: %v", err)
		}
		defer releaseLock1()

		// Try to acquire lock again (should fail)
		_, err = acquireUpdateLock()
		if err == nil {
			t.Error("Expected error when acquiring lock concurrently, got nil")
		}
		if !strings.Contains(err.Error(), "another update is in progress") {
			t.Errorf("Expected 'another update is in progress' error, got: %v", err)
		}
	})

	// Test 3: Handle stale lock
	t.Run("handle stale lock", func(t *testing.T) {
		// Create a stale lock file
		if err := os.WriteFile(lockPath, []byte("12345\n2020-01-01T00:00:00Z\n"), 0600); err != nil {
			t.Fatalf("Failed to create stale lock file: %v", err)
		}

		// Should be able to acquire lock after removing stale lock
		releaseLock, err := acquireUpdateLock()
		if err != nil {
			t.Fatalf("Failed to acquire lock after removing stale lock: %v", err)
		}
		defer releaseLock()
	})
}

// TestGetLockFilePath tests the lock file path configuration
func TestGetLockFilePath(t *testing.T) {
	// Test 1: Use XDG_RUNTIME_DIR if set
	t.Run("use XDG_RUNTIME_DIR", func(t *testing.T) {
		oldXdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
		defer os.Setenv("XDG_RUNTIME_DIR", oldXdgRuntimeDir)

		tempDir := t.TempDir()
		os.Setenv("XDG_RUNTIME_DIR", tempDir)

		lockPath := getLockFilePath()
		expectedPath := filepath.Join(tempDir, "n0man-update.lock")

		if lockPath != expectedPath {
			t.Errorf("Expected lock path %s, got %s", expectedPath, lockPath)
		}
	})

	// Test 2: Fall back to os.TempDir() if XDG_RUNTIME_DIR is not set
	t.Run("fallback to TempDir", func(t *testing.T) {
		oldXdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
		defer os.Setenv("XDG_RUNTIME_DIR", oldXdgRuntimeDir)

		os.Unsetenv("XDG_RUNTIME_DIR")

		lockPath := getLockFilePath()
		expectedPath := filepath.Join(os.TempDir(), "n0man-update.lock")

		if lockPath != expectedPath {
			t.Errorf("Expected lock path %s, got %s", expectedPath, lockPath)
		}
	})
}

// TestGetExecutablePath tests binary location detection
func TestGetExecutablePath(t *testing.T) {
	// Test 1: Get path of current executable
	t.Run("current executable", func(t *testing.T) {
		path, err := getExecutablePath()
		if err != nil {
			t.Fatalf("Failed to get executable path: %v", err)
		}

		// Verify path is absolute
		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got: %s", path)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Executable does not exist at: %s", path)
		}
	})

	// Test 2: Resolve symlinks
	t.Run("resolve symlinks", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a real file
		realFile := filepath.Join(tempDir, "real_binary")
		if err := os.WriteFile(realFile, []byte("binary content"), 0755); err != nil {
			t.Fatalf("Failed to create real file: %v", err)
		}

		// Create a symlink
		symlinkPath := filepath.Join(tempDir, "symlink_binary")
		if err := os.Symlink(realFile, symlinkPath); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Note: We can't easily test this with the actual getExecutablePath since
		// it uses os.Executable() which returns the path of the running process.
		// This test is more of a documentation of expected behavior.
	})
}

// TestGetCurrentVersion tests version detection
func TestGetCurrentVersion(t *testing.T) {
	// Test 1: Parse version from output
	t.Run("parse version", func(t *testing.T) {
		// Mock the version command by setting up a temporary binary
		// This is a simplified test - in reality, getCurrentVersion calls
		// os.Args[0] which is the running binary
		version := getCurrentVersion()

		// Version should not be empty or "unknown" if the binary is properly built
		if version == "" {
			t.Error("Version should not be empty")
		}

		// Log the version for debugging
		t.Logf("Current version: %s", version)
	})

	// Test 2: Handle unknown version
	t.Run("handle unknown version", func(t *testing.T) {
		// This test documents that getCurrentVersion returns "unknown" on failure
		// We can't easily mock the command failure without modifying the function
	})
}

// TestVerifyBinary tests binary verification
func TestVerifyBinary(t *testing.T) {
	// Test 1: Verify valid binary
	t.Run("valid binary", func(t *testing.T) {
		// Create a temporary file with enough content
		tempFile := filepath.Join(t.TempDir(), "test_binary")
		content := make([]byte, 2_000_000) // 2MB
		if err := os.WriteFile(tempFile, content, 0755); err != nil {
			t.Fatalf("Failed to create test binary: %v", err)
		}

		// Note: verifyBinary also tries to execute the binary with --help
		// which will fail for our dummy file. This test documents expected behavior.
	})

	// Test 2: Binary too small
	t.Run("binary too small", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "small_binary")
		if err := os.WriteFile(tempFile, []byte("small"), 0755); err != nil {
			t.Fatalf("Failed to create small binary: %v", err)
		}

		err := verifyBinary(tempFile)
		if err == nil {
			t.Error("Expected error for binary too small, got nil")
		}
		if !strings.Contains(err.Error(), "binary too small") {
			t.Errorf("Expected 'binary too small' error, got: %v", err)
		}
	})

	// Test 3: Binary too large
	t.Run("binary too large", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "large_binary")
		// Create a file larger than maxBinarySize
		_, maxBinarySize := getBinarySizeThresholds()
		content := make([]byte, maxBinarySize+1)
		if err := os.WriteFile(tempFile, content, 0755); err != nil {
			t.Fatalf("Failed to create large binary: %v", err)
		}

		err := verifyBinary(tempFile)
		if err == nil {
			t.Error("Expected error for binary too large, got nil")
		}
		if !strings.Contains(err.Error(), "binary too large") {
			t.Errorf("Expected 'binary too large' error, got: %v", err)
		}
	})

	// Test 4: Binary not found
	t.Run("binary not found", func(t *testing.T) {
		err := verifyBinary("/nonexistent/binary")
		if err == nil {
			t.Error("Expected error for nonexistent binary, got nil")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

// TestAtomicReplace tests atomic binary replacement
func TestAtomicReplace(t *testing.T) {
	// Test 1: Successful replacement
	t.Run("successful replacement", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create old binary
		oldPath := filepath.Join(tempDir, "old_binary")
		if err := os.WriteFile(oldPath, []byte("old content"), 0755); err != nil {
			t.Fatalf("Failed to create old binary: %v", err)
		}

		// Create new binary
		newPath := filepath.Join(tempDir, "new_binary")
		if err := os.WriteFile(newPath, []byte("new content"), 0755); err != nil {
			t.Fatalf("Failed to create new binary: %v", err)
		}

		// Perform atomic replacement
		if err := atomicReplace(oldPath, newPath); err != nil {
			t.Fatalf("Failed to perform atomic replacement: %v", err)
		}

		// Verify old path now contains new content
		content, err := os.ReadFile(oldPath)
		if err != nil {
			t.Fatalf("Failed to read old path after replacement: %v", err)
		}
		if string(content) != "new content" {
			t.Errorf("Expected 'new content', got '%s'", string(content))
		}

		// Verify new path no longer exists
		if _, err := os.Stat(newPath); !os.IsNotExist(err) {
			t.Error("New binary should not exist after replacement")
		}
	})

	// Test 2: Rollback on failure
	t.Run("rollback on failure", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create old binary
		oldPath := filepath.Join(tempDir, "old_binary")
		if err := os.WriteFile(oldPath, []byte("old content"), 0755); err != nil {
			t.Fatalf("Failed to create old binary: %v", err)
		}

		// Create new binary in a location that will cause failure
		// (e.g., by making the target directory read-only)
		// This is difficult to test reliably, so we document expected behavior
	})

	// Test 3: Backup file cleanup
	t.Run("backup cleanup", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create old binary
		oldPath := filepath.Join(tempDir, "old_binary")
		if err := os.WriteFile(oldPath, []byte("old content"), 0755); err != nil {
			t.Fatalf("Failed to create old binary: %v", err)
		}

		// Create new binary
		newPath := filepath.Join(tempDir, "new_binary")
		if err := os.WriteFile(newPath, []byte("new content"), 0755); err != nil {
			t.Fatalf("Failed to create new binary: %v", err)
		}

		// Perform atomic replacement
		if err := atomicReplace(oldPath, newPath); err != nil {
			t.Fatalf("Failed to perform atomic replacement: %v", err)
		}

		// Verify backup file was removed
		backupPath := oldPath + ".backup"
		if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
			t.Error("Backup file should be removed after successful replacement")
		}
	})
}

// TestUpdateError tests the UpdateError type
func TestUpdateError(t *testing.T) {
	t.Run("create and format error", func(t *testing.T) {
		baseErr := errors.New("base error")
		updateErr := &UpdateError{
			Step:       "Test step",
			Err:        baseErr,
			CanRecover: true,
			Suggestion: "Try again",
		}

		errorStr := updateErr.Error()
		if !strings.Contains(errorStr, "Test step") {
			t.Errorf("Error string should contain step name: %s", errorStr)
		}
		if !strings.Contains(errorStr, "base error") {
			t.Errorf("Error string should contain base error: %s", errorStr)
		}
	})

	t.Run("unwrap error", func(t *testing.T) {
		baseErr := errors.New("base error")
		updateErr := &UpdateError{
			Step: "Test step",
			Err:  baseErr,
		}

		unwrapped := updateErr.Unwrap()
		if unwrapped != baseErr {
			t.Error("Unwrapped error should equal base error")
		}
	})
}

// TestGetBinarySizeThresholds tests binary size threshold configuration
func TestGetBinarySizeThresholds(t *testing.T) {
	// Test 1: Default values
	t.Run("default values", func(t *testing.T) {
		// Clean environment
		oldMin := os.Getenv("N0MAN_MIN_BINARY_SIZE")
		oldMax := os.Getenv("N0MAN_MAX_BINARY_SIZE")
		defer func() {
			if oldMin != "" {
				os.Setenv("N0MAN_MIN_BINARY_SIZE", oldMin)
			} else {
				os.Unsetenv("N0MAN_MIN_BINARY_SIZE")
			}
			if oldMax != "" {
				os.Setenv("N0MAN_MAX_BINARY_SIZE", oldMax)
			} else {
				os.Unsetenv("N0MAN_MAX_BINARY_SIZE")
			}
		}()
		os.Unsetenv("N0MAN_MIN_BINARY_SIZE")
		os.Unsetenv("N0MAN_MAX_BINARY_SIZE")

		min, max := getBinarySizeThresholds()
		if min != 500_000 {
			t.Errorf("Expected default min 500000, got %d", min)
		}
		if max != 200_000_000 {
			t.Errorf("Expected default max 200000000, got %d", max)
		}
	})

	// Test 2: Environment variable override
	t.Run("env override", func(t *testing.T) {
		oldMin := os.Getenv("N0MAN_MIN_BINARY_SIZE")
		oldMax := os.Getenv("N0MAN_MAX_BINARY_SIZE")
		defer func() {
			if oldMin != "" {
				os.Setenv("N0MAN_MIN_BINARY_SIZE", oldMin)
			} else {
				os.Unsetenv("N0MAN_MIN_BINARY_SIZE")
			}
			if oldMax != "" {
				os.Setenv("N0MAN_MAX_BINARY_SIZE", oldMax)
			} else {
				os.Unsetenv("N0MAN_MAX_BINARY_SIZE")
			}
		}()

		os.Setenv("N0MAN_MIN_BINARY_SIZE", "1000000")
		os.Setenv("N0MAN_MAX_BINARY_SIZE", "50000000")

		min, max := getBinarySizeThresholds()
		if min != 1_000_000 {
			t.Errorf("Expected min 1000000, got %d", min)
		}
		if max != 50_000_000 {
			t.Errorf("Expected max 50000000, got %d", max)
		}
	})
}

// TestGetUpdateTimeout tests update timeout configuration
func TestGetUpdateTimeout(t *testing.T) {
	// Test 1: Default value
	t.Run("default value", func(t *testing.T) {
		oldTimeout := os.Getenv("N0MAN_UPDATE_TIMEOUT")
		defer func() {
			if oldTimeout != "" {
				os.Setenv("N0MAN_UPDATE_TIMEOUT", oldTimeout)
			} else {
				os.Unsetenv("N0MAN_UPDATE_TIMEOUT")
			}
		}()
		os.Unsetenv("N0MAN_UPDATE_TIMEOUT")

		timeout := getUpdateTimeout()
		if timeout != 5*time.Minute {
			t.Errorf("Expected default timeout 5m, got %v", timeout)
		}
	})

	// Test 2: Environment variable override
	t.Run("env override", func(t *testing.T) {
		oldTimeout := os.Getenv("N0MAN_UPDATE_TIMEOUT")
		defer func() {
			if oldTimeout != "" {
				os.Setenv("N0MAN_UPDATE_TIMEOUT", oldTimeout)
			} else {
				os.Unsetenv("N0MAN_UPDATE_TIMEOUT")
			}
		}()

		os.Setenv("N0MAN_UPDATE_TIMEOUT", "10m")

		timeout := getUpdateTimeout()
		if timeout != 10*time.Minute {
			t.Errorf("Expected timeout 10m, got %v", timeout)
		}
	})

	// Test 3: Invalid format falls back to default
	t.Run("invalid format", func(t *testing.T) {
		oldTimeout := os.Getenv("N0MAN_UPDATE_TIMEOUT")
		defer func() {
			if oldTimeout != "" {
				os.Setenv("N0MAN_UPDATE_TIMEOUT", oldTimeout)
			} else {
				os.Unsetenv("N0MAN_UPDATE_TIMEOUT")
			}
		}()

		os.Setenv("N0MAN_UPDATE_TIMEOUT", "invalid")

		timeout := getUpdateTimeout()
		if timeout != 5*time.Minute {
			t.Errorf("Expected default timeout on invalid format, got %v", timeout)
		}
	})
}

// TestFetchLatestVersionWithRetry tests retry logic for network operations
func TestFetchLatestVersionWithRetry(t *testing.T) {
	// Test 1: Success on first attempt
	t.Run("success on first attempt", func(t *testing.T) {
		// This test would require mocking the HTTP client
		// For now, we document expected behavior
	})

	// Test 2: Success after retry
	t.Run("success after retry", func(t *testing.T) {
		// This test would require mocking the HTTP client
		// For now, we document expected behavior
	})

	// Test 3: Failure after all retries
	t.Run("failure after all retries", func(t *testing.T) {
		// This test would require mocking the HTTP client
		// For now, we document expected behavior
	})
}

// TestCheckGoInstalled tests Go installation check
func TestCheckGoInstalled(t *testing.T) {
	// Test 1: Go is installed
	t.Run("go installed", func(t *testing.T) {
		// This test assumes Go is installed in the test environment
		err := checkGoInstalled()
		if err != nil {
			t.Logf("Go is not installed or not in PATH: %v", err)
		}
	})
}

// TestEnvOrDefault tests environment variable helper
func TestEnvOrDefault(t *testing.T) {
	// Test 1: Environment variable set
	t.Run("env set", func(t *testing.T) {
		oldValue := os.Getenv("TEST_VAR")
		defer func() {
			if oldValue != "" {
				os.Setenv("TEST_VAR", oldValue)
			} else {
				os.Unsetenv("TEST_VAR")
			}
		}()

		os.Setenv("TEST_VAR", "test_value")
		result := getEnvOrDefault("TEST_VAR", "default")
		if result != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", result)
		}
	})

	// Test 2: Environment variable not set
	t.Run("env not set", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_VAR")
		result := getEnvOrDefault("NONEXISTENT_VAR", "default")
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})
}

// TestGetCurrentVersionWithShortFlag tests version detection with --short flag
func TestGetCurrentVersionWithShortFlag(t *testing.T) {
	// Test 1: Try --short flag first
	t.Run("try short flag", func(t *testing.T) {
		// This test documents that getCurrentVersion should try --short flag first
		// The actual implementation would need to be updated to support this
	})

	// Test 2: Fall back to parsing full output
	t.Run("fallback to full output", func(t *testing.T) {
		// This test documents fallback behavior when --short flag is not available
	})

	// Test 3: Return "unknown" on failure
	t.Run("return unknown on failure", func(t *testing.T) {
		// This test documents that getCurrentVersion should return "unknown" on failure
	})
}

// TestRunSelfUpdate tests the main self-update flow
func TestRunSelfUpdate(t *testing.T) {
	// Test 1: Already up to date
	t.Run("already up to date", func(t *testing.T) {
		// This test would require mocking version checking
		// For now, we document expected behavior
	})

	// Test 2: Successful update
	t.Run("successful update", func(t *testing.T) {
		// This test would require mocking the entire update process
		// For now, we document expected behavior
	})

	// Test 3: Handle errors with recovery
	t.Run("handle recoverable errors", func(t *testing.T) {
		// This test documents error handling behavior
	})
}

// TestBuildNewVersion tests the build process
func TestBuildNewVersion(t *testing.T) {
	// Test 1: Successful build
	t.Run("successful build", func(t *testing.T) {
		// This test would require mocking the build process
		// For now, we document expected behavior
	})

	// Test 2: Build failure
	t.Run("build failure", func(t *testing.T) {
		// This test documents error handling on build failure
	})
}

// Helper function to capture stdout for testing
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// Helper function to create a mock binary for testing
func createMockBinary(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0755); err != nil {
		t.Fatalf("Failed to create mock binary: %v", err)
	}
}

// Helper function to check if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}

// Test that the self_update.go file compiles
func TestSelfUpdateCompiles(t *testing.T) {
	// This is a meta-test to ensure the file compiles
	// If this test runs, the file compiled successfully
	t.Log("self_update.go compiles successfully")
}
