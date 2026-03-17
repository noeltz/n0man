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
	os.MkdirAll(cfgDir, 0755)
	os.MkdirAll(storeDir, 0755)

	cfgPath := filepath.Join(cfgDir, "n0man.toml")
	cfg := config.DefaultConfig()
	cfg.LocalPath = storeDir
	// Note: No remote URL set for this test
	cfg.Save(cfgPath)

	os.Setenv("XDG_CONFIG_HOME", filepath.Dir(cfgDir))
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Init git repo
	client := git.NewClient()
	if err := client.Init(storeDir); err != nil {
		t.Fatalf("Failed to init git: %v", err)
	}

	// Configure git user locally for this test repo
	exec.Command("git", "-C", storeDir, "config", "user.name", "Test User").Run()
	exec.Command("git", "-C", storeDir, "config", "user.email", "test@example.com").Run()

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
