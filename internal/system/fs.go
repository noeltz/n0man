// Package system provides secure file system operations for n0man.
//
// PURPOSE: Abstract file operations with built-in security validation.
//
// SECURITY FEATURES:
// 1. Path validation - All paths validated against home directory
// 2. Secure permissions - Files created with 0600, directories with 0700
// 3. Symlink validation - Symlink targets validated before creation
//
// MODULE BOUNDARIES:
//   - This package handles low-level file operations
//   - Callers are responsible for high-level business logic
//   - All exported functions perform security validation
//
// USAGE PATTERN:
//
//	// Always validate paths before operations
//	if err := system.IsPathSafe(targetPath); err != nil {
//	    return fmt.Errorf("invalid path: %w", err)
//	}
//
//	// Then perform operation
//	err := system.CopyFile(src, dst)
package system

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// isPathWithinHome checks if a path is within the user's home directory.
//
// PURPOSE: Security boundary enforcement - prevent access to system files.
//
// SECURITY MODEL:
//   - All file operations restricted to user's home directory
//   - Prevents accidental exposure of /etc/passwd, /etc/shadow, etc.
//   - Defense in depth against path traversal attacks
//
// ALGORITHM:
//  1. Get home directory (e.g., /home/user)
//  2. Clean both paths (resolve .., remove duplicates)
//  3. Calculate relative path from home to target
//  4. If relative path starts with "..", target is outside home
//
// ENVIRONMENT OVERRIDE:
//
//	N0MAN_ALLOW_OUTSIDE_HOME=true bypasses this check
//	WARNING: Only use for testing in trusted environments
//	SIDE EFFECT: Prints warning to stderr
//
// RETURNS:
//   - true, nil: Path is within home directory
//   - false, nil: Path is outside home directory
//   - false, err: Error determining home directory
//
// TEST CASES:
//   - "/home/user/file" → true (inside home)
//   - "/etc/passwd" → false (outside home)
//   - "/home/user/../other" → false (escapes home)
func isPathWithinHome(path string) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("cannot determine home directory: %w", err)
	}

	// SECURITY OVERRIDE: Allow paths outside home for testing
	// CONSTRAINT: Only set N0MAN_ALLOW_OUTSIDE_HOME in test environments
	// SIDE EFFECT: Prints warning to stderr (visible to user)
	if allowOutside := os.Getenv("N0MAN_ALLOW_OUTSIDE_HOME"); allowOutside == "true" {
		// Log warning to stderr (only once per process)
		fmt.Fprintln(os.Stderr, "WARNING: N0MAN_ALLOW_OUTSIDE_HOME is set - path restrictions disabled")
		return true, nil
	}

	// Clean both paths for comparison
	cleanPath := filepath.Clean(path)
	cleanHome := filepath.Clean(homeDir)

	// Check if path is within home directory
	rel, err := filepath.Rel(cleanHome, cleanPath)
	if err != nil {
		return false, err
	}

	// If the relative path starts with "..", it's outside home
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}

	return true, nil
}

// IsPathSafe validates that a path is safe to use.
//
// PURPOSE: Comprehensive path validation for security.
//
// VALIDATION CHECKS (in order):
//  1. Not empty
//  2. No null bytes (prevents truncation attacks)
//  3. No path traversal (.. components)
//  4. Within home directory
//
// ATTACK VECTORS PREVENTED:
//   - Path traversal: /home/user/../../../etc/passwd
//   - Null byte injection: /home/user/file.txt\x00/etc/passwd
//   - Empty path: "" (could default to root)
//
// PARAMETERS:
//   - path: Path to validate (absolute or relative)
//
// RETURNS:
//   - nil: Path is safe to use
//   - error: Path is unsafe (do not proceed with operation)
//
// USAGE:
//
//	if err := system.IsPathSafe(userPath); err != nil {
//	    return fmt.Errorf("invalid path: %w", err)
//	}
//	// Safe to proceed with file operation
func IsPathSafe(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// SECURITY: Reject null bytes to prevent truncation attacks
	// ATTACK: "file.txt\x00/etc/passwd" could truncate at null byte
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path cannot contain null bytes")
	}

	// Clean the path to resolve any .. components
	cleanPath := filepath.Clean(path)

	// SECURITY: Reject paths that escape via parent directory
	// ATTACK: "../../../etc/passwd" tries to escape home directory
	if strings.HasPrefix(cleanPath, "..") {
		return fmt.Errorf("path cannot contain parent directory references")
	}

	// Check if path is within home directory
	withinHome, err := isPathWithinHome(cleanPath)
	if err != nil {
		return err
	}
	if !withinHome {
		return fmt.Errorf("path must be within home directory (%s)", path)
	}

	return nil
}

// ValidateSymlinkTarget validates that a symlink target is safe.
//
// PURPOSE: Prevent symlinks from pointing to sensitive locations.
//
// SECURITY CONCERN:
//
//	Without validation, an attacker could create symlinks like:
//	~/.config/myfile → /etc/shadow
//	Then n0man would backup/track /etc/shadow content
//
// VALIDATION:
//  1. Read symlink target
//  2. Resolve to absolute path if relative
//  3. Clean path (remove .., normalize)
//  4. Verify target is within home directory
//
// PARAMETERS:
//   - linkPath: Path to the symlink (not the target)
//
// RETURNS:
//   - nil: Symlink target is safe
//   - error: Symlink target is unsafe or symlink is broken
func ValidateSymlinkTarget(linkPath string) error {
	// Get the symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("cannot read symlink: %w", err)
	}

	// Resolve to absolute path if relative
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}

	// Clean the resolved path
	cleanTarget := filepath.Clean(target)

	// Check if target is within home directory
	withinHome, err := isPathWithinHome(cleanTarget)
	if err != nil {
		return err
	}
	if !withinHome {
		return fmt.Errorf("symlink target must be within home directory: %s", target)
	}

	return nil
}

// MovePath safely moves a file or directory from src to dst.
func MovePath(src, dst string) error {
	// Validate source and destination paths
	if err := IsPathSafe(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := IsPathSafe(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// First check if dst already exists
	if _, err := os.Lstat(dst); err == nil {
		return fmt.Errorf("destination already exists: %s", dst)
	}

	// Ensure parent dir of dst exists with secure permissions
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return fmt.Errorf("failed to create parent directory for destination: %w", err)
	}

	// Try rename first
	if err := os.Rename(src, dst); err != nil {
		// Fallback to copy+delete
		// Check if source is a directory
		srcInfo, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("failed to stat source: %w", err)
		}

		var copyErr error
		if srcInfo.IsDir() {
			copyErr = CopyDir(src, dst)
		} else {
			copyErr = CopyFile(src, dst)
		}

		if copyErr != nil {
			return fmt.Errorf("failed to move (rename failed: %v, copy failed: %v)", err, copyErr)
		}
		if removeErr := os.RemoveAll(src); removeErr != nil {
			return fmt.Errorf("copied but failed to remove original: %v", removeErr)
		}
	}
	return nil
}

// CopyFile copies a file from src to dst with secure permissions.
func CopyFile(src, dst string) error {
	// Validate paths
	if err := IsPathSafe(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := IsPathSafe(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Create with secure permissions (user-only read/write)
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Preserve source file permissions but keep user-only restriction
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	// Use source permissions but mask to user-only (0600 base)
	secureMode := info.Mode()&0600 | 0600
	return os.Chmod(dst, secureMode)
}

// CopyDir recursively copies a directory from src to dst with secure permissions.
func CopyDir(src, dst string) error {
	// Validate paths
	if err := IsPathSafe(src); err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := IsPathSafe(dst); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	if _, err := os.Stat(src); err != nil {
		return err
	}

	// Create directory with secure permissions (user-only rwx)
	if err := os.MkdirAll(dst, 0700); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateSymlink safely creates a symlink at linkPath pointing to targetPath.
func CreateSymlink(targetPath, linkPath string) error {
	// Validate both paths
	if err := IsPathSafe(targetPath); err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}
	if err := IsPathSafe(linkPath); err != nil {
		return fmt.Errorf("invalid link path: %w", err)
	}

	// Ensure parent dir of link exists with secure permissions
	if err := os.MkdirAll(filepath.Dir(linkPath), 0700); err != nil {
		return fmt.Errorf("failed to create parent directory for symlink: %w", err)
	}

	// If a symlink already exists at linkPath, fail unless we implement forced overwrite
	if _, err := os.Lstat(linkPath); err == nil {
		return fmt.Errorf("symlink or file already exists at link path: %s", linkPath)
	}

	return os.Symlink(targetPath, linkPath)
}

// IsSymlink returns true if the given path is a symlink
func IsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return info.Mode()&fs.ModeSymlink != 0, nil
}
