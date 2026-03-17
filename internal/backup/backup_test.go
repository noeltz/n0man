package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/system"
)

func TestCreateSnapshot(t *testing.T) {
	// Allow paths outside home for testing
	t.Setenv("N0MAN_ALLOW_OUTSIDE_HOME", "true")

	tempDir := t.TempDir()

	// Mock config
	cfg := config.DefaultConfig()
	cfg.LocalPath = filepath.Join(tempDir, "store")
	os.MkdirAll(cfg.LocalPath, 0755)

	// Create dummy file
	dummyFile := filepath.Join(tempDir, "test.txt")
	_ = os.WriteFile(dummyFile, []byte("original content"), 0644)

	cfg.SetPath("test.txt", dummyFile)

	// 1. Create backup
	ts, err := CreateSnapshot(&cfg)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	backupDir := filepath.Join(cfg.LocalPath, ".backups", ts)
	backupFile := filepath.Join(backupDir, "test.txt")

	b, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backed up file: %v", err)
	}
	if string(b) != "original content" {
		t.Errorf("Backup contains wrong content: %v", string(b))
	}

	// 2. Modify original and mock symlink (how n0man works)
	// move original to store, and symlink it
	storeFile := filepath.Join(cfg.LocalPath, "test.txt")
	_ = system.MovePath(dummyFile, storeFile)
	_ = system.CreateSymlink(storeFile, dummyFile)

	_ = os.WriteFile(storeFile, []byte("changed content"), 0644)

	// 3. Create second backup (should backup resolved symlink or the store file)
	ts2, err := CreateSnapshot(&cfg)
	if err != nil {
		t.Fatalf("Second snapshot failed: %v", err)
	}

	backupFile2 := filepath.Join(cfg.LocalPath, ".backups", ts2, "test.txt")
	b2, err := os.ReadFile(backupFile2)
	if err != nil {
		t.Fatalf("Failed to read second backup file: %v", err)
	}

	if string(b2) != "changed content" {
		t.Errorf("Second backup expected 'changed content', got '%s'", string(b2))
	}
}

func TestCleanOldBackups(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.LocalPath = filepath.Join(tempDir, "store")
	cfg.Settings.HousekeepingMaxBackups = 2

	backupsDir := filepath.Join(cfg.LocalPath, ".backups")
	os.MkdirAll(backupsDir, 0755)

	// Create 4 dummy backups
	dirs := []string{"20260101_000000", "20260102_000000", "20260103_000000", "20260104_000000"}
	for _, d := range dirs {
		_ = os.MkdirAll(filepath.Join(backupsDir, d), 0755)
		// sleep a tiny bit just in case, though name sorting applies
		time.Sleep(1 * time.Millisecond)
	}

	err := CleanOldBackups(&cfg)
	if err != nil {
		t.Fatalf("CleanOldBackups failed: %v", err)
	}

	entries, _ := os.ReadDir(backupsDir)
	if len(entries) != 2 {
		t.Errorf("Expected 2 backups, found %d", len(entries))
	}

	// Should keep the latest two
	names := []string{entries[0].Name(), entries[1].Name()}
	if strings.Join(names, ",") != "20260103_000000,20260104_000000" {
		t.Errorf("Expected to keep the latest 2 backups. Real result: %v", names)
	}
}

func TestListBackups(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.LocalPath = filepath.Join(tempDir, "store")
	backupsDir := filepath.Join(cfg.LocalPath, ".backups")
	os.MkdirAll(backupsDir, 0755)

	// Create 3 dummy backups
	dirs := []string{"20260101_000000", "20260102_000000", "20260103_000000"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(backupsDir, d), 0755)
	}

	backups, err := ListBackups(&cfg)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}

	// Should be sorted latest first
	if backups[0] != "20260103_000000" {
		t.Errorf("Expected latest backup first, got %s", backups[0])
	}
}

func TestRestoreBackup(t *testing.T) {
	// Allow paths outside home for testing
	t.Setenv("N0MAN_ALLOW_OUTSIDE_HOME", "true")

	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.LocalPath = filepath.Join(tempDir, "store")
	_ = os.MkdirAll(cfg.LocalPath, 0755)

	// Create a backup with a file
	backupTime := "20260101_120000"
	backupDir := filepath.Join(cfg.LocalPath, ".backups", backupTime)
	_ = os.MkdirAll(backupDir, 0755)

	testFile := "config.txt"
	backupFile := filepath.Join(backupDir, testFile)
	_ = os.WriteFile(backupFile, []byte("backup content"), 0644)

	// Restore
	err := RestoreBackup(&cfg, backupTime)
	if err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	// Check file restored to store
	restoredFile := filepath.Join(cfg.LocalPath, testFile)
	b, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(b) != "backup content" {
		t.Errorf("Restored content mismatch: got %s", string(b))
	}
}

func TestCheckIfPathInStore(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.LocalPath = filepath.Join(tempDir, "store")
	_ = os.MkdirAll(cfg.LocalPath, 0755)

	// Create some files in store
	os.WriteFile(filepath.Join(cfg.LocalPath, "file1.txt"), []byte("data1"), 0644)
	os.MkdirAll(filepath.Join(cfg.LocalPath, "subdir"), 0755)
	os.WriteFile(filepath.Join(cfg.LocalPath, "subdir", "file2.txt"), []byte("data2"), 0644)

	// Test existing file
	if !CheckIfPathInStore(&cfg, filepath.Join(cfg.LocalPath, "file1.txt")) {
		t.Error("Expected CheckIfPathInStore to return true for existing file")
	}

	// Test existing file in subdirectory
	if !CheckIfPathInStore(&cfg, filepath.Join(cfg.LocalPath, "subdir", "file2.txt")) {
		t.Error("Expected CheckIfPathInStore to return true for existing subdirectory file")
	}

	// Test non-existent file
	if CheckIfPathInStore(&cfg, filepath.Join(cfg.LocalPath, "nonexistent.txt")) {
		t.Error("Expected CheckIfPathInStore to return false for non-existent file")
	}

	// Test path outside store
	if CheckIfPathInStore(&cfg, filepath.Join(tempDir, "outside.txt")) {
		t.Error("Expected CheckIfPathInStore to return false for path outside store")
	}
}
