package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "n0man.toml")

	// Create a default config
	cfg := DefaultConfig()
	cfg.RemoteURL = "git@github.com:user/dotfiles.git"
	cfg.LocalPath = "/home/user/.local/share/n0man/store"
	cfg.SetPath("nvim", "~/.config/nvim")
	cfg.SetIgnores("nvim", []string{"*.swap", "backup/"})
	cfg.Overrides["work"] = map[string]string{
		"ssh": "~/.ssh/config_work",
	}

	// Save it
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load it
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify
	if loadedCfg.RemoteURL != cfg.RemoteURL || loadedCfg.LocalPath != cfg.LocalPath {
		t.Errorf("Top-level fields mismatch")
	}

	if !reflect.DeepEqual(cfg.GetPaths(), loadedCfg.GetPaths()) {
		t.Errorf("Paths mismatch")
	}

	if !reflect.DeepEqual(cfg.GetIgnores(), loadedCfg.GetIgnores()) {
		t.Errorf("Ignores mismatch")
	}

	if !reflect.DeepEqual(cfg.Overrides, loadedCfg.Overrides) {
		t.Errorf("Overrides mismatch")
	}
}

func TestConfigLoadNotExist(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.toml")

	_, err := Load(configPath)
	if err != ErrConfigNotFound {
		t.Errorf("Expected ErrConfigNotFound, got: %v", err)
	}
}

func TestGetTargetPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetPath("bashrc", "~/.bashrc")
	cfg.SetPath("zshrc", "~/.zshrc")

	hostname, _ := os.Hostname()
	cfg.Overrides[hostname] = map[string]string{
		"bashrc": "~/.bashrc_overridden",
	}

	// Should be overridden
	if got := cfg.GetTargetPath("bashrc"); got != "~/.bashrc_overridden" {
		t.Errorf("Expected ~/.bashrc_overridden, got %s", got)
	}

	// Should be default
	if got := cfg.GetTargetPath("zshrc"); got != "~/.zshrc" {
		t.Errorf("Expected ~/.zshrc, got %s", got)
	}

	// Should be empty for untracked
	if got := cfg.GetTargetPath("nonexistent"); got != "" {
		t.Errorf("Expected empty string, got %s", got)
	}
}

func TestConfigDelete(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetPath("vim", "~/.vimrc")
	cfg.SetPath("nvim", "~/.config/nvim")
	cfg.SetIgnores("nvim", []string{"*.swap"})

	// Delete one path
	cfg.Delete("vim")

	paths := cfg.GetPaths()
	if _, ok := paths["vim"]; ok {
		t.Errorf("vim should be deleted")
	}

	// Verify other path still exists
	if _, ok := paths["nvim"]; !ok {
		t.Errorf("nvim should still exist")
	}

	// Verify ignores were cleaned up
	ignores := cfg.GetIgnores()
	if _, ok := ignores["nvim"]; !ok {
		t.Errorf("nvim ignores should be cleaned up after delete")
	}
}

func TestConfigDeleteWithNoIgnores(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetPath("test", "~/.testfile")

	// Should not panic
	cfg.Delete("test")

	paths := cfg.GetPaths()
	if _, ok := paths["test"]; ok {
		t.Errorf("test should be deleted")
	}
}

func TestConfigDeleteEmpty(t *testing.T) {
	cfg := DefaultConfig()

	// Should not panic
	cfg.Delete("nonexistent")
}

func TestGetIgnoresEmpty(t *testing.T) {
	cfg := DefaultConfig()
	ignores := cfg.GetIgnores()
	if len(ignores) != 0 {
		t.Errorf("Expected empty ignores, got %v", ignores)
	}
}

func TestSetIgnores(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetPath("test", "~/.testfile")

	cfg.SetIgnores("test", []string{"*.log", "*.tmp"})

	ignores := cfg.GetIgnores()
	if len(ignores["test"]) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(ignores["test"]))
	}
}

func TestGetPathsEmpty(t *testing.T) {
	cfg := DefaultConfig()
	paths := cfg.GetPaths()
	if paths == nil {
		t.Errorf("Expected empty map, got nil")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath failed: %v", err)
	}
	if path == "" {
		t.Errorf("Expected non-empty path")
	}
}

func TestDefaultStorePath(t *testing.T) {
	path, err := DefaultStorePath()
	if err != nil {
		t.Fatalf("DefaultStorePath failed: %v", err)
	}
	if path == "" {
		t.Errorf("Expected non-empty path")
	}
}
