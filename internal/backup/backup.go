package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/system"
)

func CreateSnapshot(cfg *config.Config) (string, error) {
	return CreateSnapshotWithContext(context.Background(), cfg)
}

// CreateSnapshotWithContext creates a backup snapshot with context for cancellation
func CreateSnapshotWithContext(ctx context.Context, cfg *config.Config) (string, error) {
	if len(cfg.GetPaths()) == 0 {
		return "", nil // Nothing to backup
	}

	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(cfg.LocalPath, ".backups", timestamp)

	// Create backup directory with secure permissions
	err := os.MkdirAll(backupDir, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to create backup dir: %w", err)
	}

	homeDir, _ := os.UserHomeDir()

	// Copy actual files mapped in config
	for name := range cfg.GetPaths() {
		// Check context for cancellation
		select {
		case <-ctx.Done():
			// Clean up partial backup
			_ = os.RemoveAll(backupDir)
			return "", ctx.Err()
		default:
		}

		targetPath := cfg.GetTargetPath(name)
		if targetPath == "" {
			continue
		}

		realTarget := targetPath
		if strings.HasPrefix(targetPath, "~") {
			realTarget = strings.Replace(targetPath, "~", homeDir, 1)
		}

		// Read link to ensure we backup what it currently points to, or if it's the actual file
		// Note: The intention of backup is to store the actual file content before n0man or a pull might overwrite it

		info, err := os.Lstat(realTarget)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Nothing to backup if it doesn't exist
			}
			return "", fmt.Errorf("failed to stat %s: %w", targetPath, err)
		}

		srcPath := realTarget
		if info.Mode()&os.ModeSymlink != 0 {
			resolvedPath, err := os.Readlink(realTarget)
			if err == nil {
				srcPath = resolvedPath

				if !filepath.IsAbs(resolvedPath) {
					srcPath = filepath.Join(filepath.Dir(realTarget), resolvedPath)
				}
			}
		}

		// Re-stat the resolved path to correctly detect dir vs file
		resolvedInfo, err := os.Stat(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("failed to stat resolved path %s: %w", srcPath, err)
		}

		destPath := filepath.Join(backupDir, name)

		if resolvedInfo.IsDir() {
			err = copyDir(srcPath, destPath)
		} else {
			err = system.CopyFile(srcPath, destPath)
		}

		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to backup %s: %w", name, err)
		}
	}

	return timestamp, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return system.CopyFile(path, destPath)
	})
}

// CleanOldBackups keeps only the max amount of backups specified in config
func CleanOldBackups(cfg *config.Config) error {
	maxBackups := cfg.Settings.HousekeepingMaxBackups
	if maxBackups <= 0 {
		return nil
	}

	backupBaseDir := filepath.Join(cfg.LocalPath, ".backups")
	entries, err := os.ReadDir(backupBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var directories []string
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	// Sort chronologically (oldest first)
	sort.Strings(directories)

	if len(directories) > maxBackups {
		// Remove oldest
		toRemove := len(directories) - maxBackups
		for i := 0; i < toRemove; i++ {
			dirToRemove := filepath.Join(backupBaseDir, directories[i])
			if err := os.RemoveAll(dirToRemove); err != nil {
				// Log error but continue trying to remove others
				return fmt.Errorf("failed to remove old backup %s: %w", dirToRemove, err)
			}
		}
	}

	return nil
}

// ListBackups returns a slice of backup timestamps, sorted Latest first
func ListBackups(cfg *config.Config) ([]string, error) {
	backupBaseDir := filepath.Join(cfg.LocalPath, ".backups")
	entries, err := os.ReadDir(backupBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	// Sort chronologically (latest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	return backups, nil
}

// RestoreBackup restores a specific backup timestamp to the store
func RestoreBackup(cfg *config.Config, timestamp string) error {
	backupDir := filepath.Join(cfg.LocalPath, ".backups", timestamp)
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup %s not found", timestamp)
	}

	return filepath.WalkDir(backupDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == backupDir {
			return nil
		}

		relPath, err := filepath.Rel(backupDir, path)
		if err != nil {
			return err
		}

		storePath := filepath.Join(cfg.LocalPath, relPath)

		if d.IsDir() {
			// Use secure permissions (user-only)
			return os.MkdirAll(storePath, 0700)
		}

		return system.CopyFile(path, storePath)
	})
}

// CheckIfPathInStore checks if the given path is within the store directory AND exists
func CheckIfPathInStore(cfg *config.Config, localPath string) bool {
	storePath := cfg.LocalPath
	if storePath == "" {
		return false
	}

	absStore, err := filepath.Abs(storePath)
	if err != nil {
		return false
	}
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return false
	}

	// Check if absPath is within absStore
	rel, err := filepath.Rel(absStore, absPath)
	if err != nil {
		return false
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return false
	}

	// Also check that the file/directory actually exists
	_, err = os.Stat(absPath)
	return err == nil
}
