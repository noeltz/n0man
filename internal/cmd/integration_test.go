package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/system"
)

func TestAddListRmFlow(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Setup mock config and store data
	configHome := filepath.Join(tempDir, "config_home")
	cfgDir := filepath.Join(configHome, "n0man")
	storeDir := filepath.Join(tempDir, "store")
	os.MkdirAll(cfgDir, 0755)
	os.MkdirAll(storeDir, 0755)

	cfgPath := filepath.Join(cfgDir, "n0man.toml")
	cfg := config.DefaultConfig()
	cfg.LocalPath = storeDir
	_ = cfg.Save(cfgPath)

	// Mock XDG paths to point to temp dir
	os.Setenv("XDG_CONFIG_HOME", configHome)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_DATA_HOME", storeDir)
	defer os.Unsetenv("XDG_DATA_HOME")
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Set testing flag so we don't accidentally trash real dotfiles
	// (Go's testing framework doesn't have a reliable `os.UserConfigDir` mock out of box
	// without overriding env variables for Windows/macOS/Linux)

	// Create a dummy file to add
	dummyFilePath := filepath.Join(tempDir, "dummy.txt")
	os.WriteFile(dummyFilePath, []byte("dummy file content"), 0644)

	// 2. Add
	rootCmd.SetArgs([]string{"add", dummyFilePath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Add command failed: %v", err)
	}

	// Verify symlink and move occurred
	isLink, err := system.IsSymlink(dummyFilePath)
	if err != nil || !isLink {
		t.Errorf("Expected dummyFilePath to be a symlink: %v", err)
	}

	storeFile := filepath.Join(storeDir, "dummy.txt")
	b, err := os.ReadFile(storeFile)
	if err != nil || string(b) != "dummy file content" {
		t.Errorf("Expected store file to contain content: %v", err)
	}

	// 3. List
	rootCmd.SetArgs([]string{"list"})

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	_ = w.Close()
	os.Stdout = oldStdout

	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(r)

	if !strings.Contains(stdout.String(), "dummy.txt") {
		t.Errorf("List output missing 'dummy.txt'. Got: %v", stdout.String())
	}

	// 4. Rm
	rootCmd.SetArgs([]string{"rm", "dummy.txt"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Rm command failed: %v", err)
	}

	// Verify restored file
	isLink, err = system.IsSymlink(dummyFilePath)
	if err == nil && isLink {
		t.Errorf("Expected dummyFilePath to be a regular file, not a symlink")
	}

	b, err = os.ReadFile(dummyFilePath)
	if err != nil || string(b) != "dummy file content" {
		t.Errorf("Expected restored file to contain content: %v", err)
	}

	// Verify removed from config
	cfgPtr, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if _, ok := cfgPtr.GetPaths()["dummy.txt"]; ok {
		t.Errorf("Expected dummy.txt to be removed from config")
	}
}
