package git

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GitError represents a Git command error with sanitized output
type GitError struct {
	Args   []string
	Stderr string
	Err    error
}

// Error implements the error interface
func (e *GitError) Error() string {
	return fmt.Sprintf("git %v failed: %v (stderr: %s)", e.Args, e.Err, e.Stderr)
}

// Unwrap returns the underlying error for errors.Is/As
func (e *GitError) Unwrap() error {
	return e.Err
}

// Client defines the interface for interacting with Git
type Client interface {
	Clone(url, dest string) error
	Init(path string) error
	Add(path string, files ...string) error
	Commit(path, message string) error
	Push(path string) error
	Pull(path string) error
	HasChanges(path string) (bool, error)
	AbortRebase(path string) error
	ContinueRebase(path string) error
	IsRebasing(path string) bool
	ResolveConflict(path string, useLocal bool) error
}

// OSClient implements the Client interface by executing git commands via the OS
type OSClient struct{}

// NewClient returns a new OSClient as a Client interface
func NewClient() Client {
	return &OSClient{}
}

// validateGitURL validates that a Git URL is safe to use
// It rejects URLs with shell metacharacters that could lead to command injection
func validateGitURL(gitURL string) error {
	if gitURL == "" {
		return fmt.Errorf("git URL cannot be empty")
	}

	// Reject URLs with shell metacharacters that could be exploited
	shellMetachars := []string{";", "|", "&", "$", "`", "\\", "!", "{", "}", "<", ">", "(", ")", "\n", "\r"}
	for _, char := range shellMetachars {
		if strings.Contains(gitURL, char) {
			return fmt.Errorf("git URL contains invalid character: %s", char)
		}
	}

	// Validate URL format - must be one of: https://, http://, git@, ssh://, or local path
	if !strings.HasPrefix(gitURL, "https://") &&
		!strings.HasPrefix(gitURL, "http://") &&
		!strings.HasPrefix(gitURL, "git@") &&
		!strings.HasPrefix(gitURL, "ssh://") &&
		!strings.HasPrefix(gitURL, "file://") {
		// Could be a local path, which is acceptable
		// Validate it's a clean path
		cleanPath := filepath.Clean(gitURL)
		if strings.HasPrefix(cleanPath, "..") {
			return fmt.Errorf("git URL path cannot contain parent directory references")
		}
	}

	// For SSH URLs, validate the format
	if strings.HasPrefix(gitURL, "ssh://") {
		parsed, err := url.Parse(gitURL)
		if err != nil {
			return fmt.Errorf("invalid SSH URL format: %w", err)
		}
		// Check for command injection in user info
		if parsed.User != nil {
			username := parsed.User.Username()
			if strings.ContainsAny(username, ";|&$`\\!") {
				return fmt.Errorf("invalid characters in SSH username")
			}
		}
	}

	return nil
}

// validatePath validates that a file system path is safe to use
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean the path to resolve any .. components
	cleanPath := filepath.Clean(path)

	// Reject paths with null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path cannot contain null bytes")
	}

	// Reject paths with shell metacharacters
	shellMetachars := []string{";", "|", "&", "$", "`", "!", "(", ")", "<", ">", "\n", "\r"}
	for _, char := range shellMetachars {
		if strings.Contains(cleanPath, char) {
			return fmt.Errorf("path contains invalid character: %s", char)
		}
	}

	return nil
}

// sanitizeSensitiveData redacts sensitive information from strings
func sanitizeSensitiveData(s string) string {
	// Redact potential secrets and credentials
	sensitivePatterns := []string{
		`(?i)(password|secret|key|token|pass|pwd)[=:]\s*\S+`,
		`(?i)(password|secret|key|token|pass|pwd)\s*[=:]\s*['"]?[^'"\s]+`,
		`ghp_[a-zA-Z0-9]{36}`,
		`github_pat_[a-zA-Z0-9]{82}`,
		`sk-[a-zA-Z0-9]{20,}`,
		`sk-ant-[a-zA-Z0-9_-]{95}`,
		`AKIA[0-9A-Z]{16}`,
		`-----BEGIN\s+(?:RSA\s+|DSA\s+|EC\s+|OPENSSH\s+)?PRIVATE\s+KEY-----`,
	}

	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(pattern)
		s = re.ReplaceAllString(s, "[REDACTED]")
	}

	return s
}

// redactArgs redacts sensitive information from command arguments
func redactArgs(args []string) []string {
	redacted := make([]string, len(args))
	copy(redacted, args)

	for i, arg := range redacted {
		// Redact URLs that might contain credentials
		if strings.Contains(arg, "@") && (strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "ssh://")) {
			redacted[i] = "[URL REDACTED]"
		}
		// Redact potential tokens
		if strings.HasPrefix(arg, "ghp_") || strings.HasPrefix(arg, "sk-") {
			redacted[i] = "[REDACTED]"
		}
	}

	return redacted
}

func (c *OSClient) runCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Sanitize stderr to prevent leaking sensitive information
		stderrContent := sanitizeSensitiveData(stderr.String())
		redactedArgs := redactArgs(args)
		return &GitError{
			Args:   redactedArgs,
			Stderr: stderrContent,
			Err:    err,
		}
	}
	return nil
}

func (c *OSClient) Clone(gitURL, dest string) error {
	// Validate Git URL to prevent command injection
	if err := validateGitURL(gitURL); err != nil {
		return fmt.Errorf("invalid Git URL: %w", err)
	}

	// Validate destination path
	if err := validatePath(dest); err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	return c.runCmd("", "clone", gitURL, dest)
}

func (c *OSClient) Init(path string) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	return c.runCmd(path, "init")
}

func (c *OSClient) Add(path string, files ...string) error {
	args := append([]string{"add"}, files...)
	return c.runCmd(path, args...)
}

func (c *OSClient) Commit(path, message string) error {
	return c.runCmd(path, "commit", "-m", message)
}

func (c *OSClient) Push(path string) error {
	return c.runCmd(path, "push", "origin", "HEAD")
}

func (c *OSClient) Pull(path string) error {
	return c.runCmd(path, "pull", "--rebase", "origin", "HEAD")
}

func (c *OSClient) HasChanges(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, &GitError{
			Args:   []string{"status", "--porcelain"},
			Stderr: sanitizeSensitiveData(stderr.String()),
			Err:    err,
		}
	}

	return stdout.Len() > 0, nil
}

func (c *OSClient) AbortRebase(path string) error {
	return c.runCmd(path, "rebase", "--abort")
}

func (c *OSClient) ContinueRebase(path string) error {
	return c.runCmd(path, "rebase", "--continue")
}

func (c *OSClient) IsRebasing(path string) bool {
	// Check for rebase indicators
	rebaseMerge := filepath.Join(path, ".git", "rebase-merge")
	rebaseApply := filepath.Join(path, ".git", "rebase-apply")

	_, err1 := os.Stat(rebaseMerge)
	_, err2 := os.Stat(rebaseApply)

	return err1 == nil || err2 == nil
}

func (c *OSClient) ResolveConflict(path string, useLocal bool) error {
	// During rebase:
	// HEAD (ours) = The changes we are rebasing ONTO (Remote/Upstream)
	// MERGE_HEAD (theirs) = The changes we are rebasing (Local)

	// Wait, to keep LOCAL changes during REBASE, we want THEIRs.
	// To keep REMOTE changes during REBASE, we want OURS.

	gitStrategy := "--ours"
	if useLocal {
		gitStrategy = "--theirs"
	}

	err := c.runCmd(path, "checkout", gitStrategy, ".")
	if err != nil {
		return err
	}
	return c.runCmd(path, "add", ".")
}
