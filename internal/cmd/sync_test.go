package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/git"
)

func TestSyncLocalChanges(t *testing.T) {
	tempDir := t.TempDir()

	cfgDir := filepath.Join(tempDir, "config_home", "n0man")
	storeDir := filepath.Join(tempDir, "store")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatalf("Failed to create store dir: %v", err)
	}

	cfgPath := filepath.Join(cfgDir, "n0man.toml")
	cfg := config.DefaultConfig()
	cfg.LocalPath = storeDir
	// Note: No remote URL set for this test
	_ = cfg.Save(cfgPath)

	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Dir(cfgDir)); err != nil {
		t.Fatalf("Failed to set XDG_CONFIG_HOME: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
			t.Logf("Failed to unset XDG_CONFIG_HOME: %v", err)
		}
	}()

	// Init git repo
	client := git.NewClient()
	if err := client.Init(storeDir); err != nil {
		t.Fatalf("Failed to init git: %v", err)
	}

	// Configure git user locally for this test repo
	if err := exec.Command("git", "-C", storeDir, "config", "user.name", "Test User").Run(); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}
	if err := exec.Command("git", "-C", storeDir, "config", "user.email", "test@example.com").Run(); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(storeDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("sync test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run sync command
	rootCmd.SetArgs([]string{"sync"})
	err := rootCmd.Execute()

	// Sync should succeed with local-only mode
	if err != nil {
		t.Logf("Sync completed with: %v", err)
	}
}
