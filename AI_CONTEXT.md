# AI/Agent Context Guide for n0man

This document provides context and patterns for AI agents working with the n0man codebase.

## Quick Reference

| Topic | Location | Key Pattern |
|-------|----------|-------------|
| Security Scanning | `internal/security/scanner.go` | Facade pattern with pipeline detection |
| Path Validation | `internal/system/fs.go` | Defense in depth validation |
| Git Operations | `internal/git/git.go` | Command abstraction with sanitization |
| Configuration | `internal/config/config.go` | TOML with type-safe accessors |
| CLI Commands | `internal/cmd/*.go` | Cobra with pre-flight checks |
| Backup System | `internal/backup/backup.go` | Context-aware snapshots |

---

## Code Comment Conventions

n0man uses structured comments to help AI agents understand code intent:

### Comment Types

```go
// PURPOSE: Explains WHY this code exists
// Example:
// PURPOSE: Prevent accidental commit of sensitive data

// PATTERN: Identifies design patterns in use
// Example:
// PATTERN: Facade pattern - provides simple interface to complex subsystem

// CONSTRAINT: Documents limitations or requirements
// Example:
// CONSTRAINT: Only set N0MAN_ALLOW_OUTSIDE_HOME in test environments

// SEE: References related code
// Example:
// SEE: Similar implementation in services/auth.ts

// TEST CASE: Documents expected test scenarios
// Example:
// TEST CASE: Empty file → Passed=true, Findings=0

// ARCHITECTURE: Explains component relationships
// Example:
// ARCHITECTURE:
//   Scanner (orchestrator)
//   ├── PatternMatcher (file path patterns)
//   └── Detector (content analysis)
```

### Function Documentation Template

```go
// FunctionName does X for purpose Y.
// 
// PURPOSE: Why this function exists
// 
// DETECTION PIPELINE / ALGORITHM / STEPS:
//   1. First step with explanation
//   2. Second step with explanation
//   3. Third step with explanation
// 
// PARAMETERS:
//   - param1: Description including units/constraints
//   - param2: Description including valid range
// 
// RETURNS:
//   - Type: What it represents, when it's non-zero/non-nil
//   - error: When it's non-nil, common error cases
// 
// CONSTRAINTS:
//   - Performance characteristics
//   - Memory usage patterns
//   - Thread safety notes
// 
// USAGE EXAMPLE:
//   result, err := FunctionName(param1, param2)
//   if err != nil {
//       // Handle error
//   }
// 
// TEST CASES:
//   - Normal case: input X → output Y
//   - Edge case: empty input → default value
//   - Error case: invalid input → specific error
func FunctionName(param1 Type, param2 Type) (Type, error) {
```

---

## Module Boundaries

### `internal/security/` - Secret Detection

**Responsibility:** Detect and prevent commit of sensitive data.

**Key Types:**
- `Scanner` - Main orchestrator (facade pattern)
- `Detector` - Content analysis with regex and entropy
- `PatternMatcher` - File path pattern matching
- `Finding` - Individual secret detection result

**Integration Points:**
```go
// Used in: internal/cmd/sync.go, internal/cmd/add.go
scanner := security.NewScanner(&cfg.Security)
report, err := scanner.ScanPath(cfg.LocalPath)
if report.TotalFindings > 0 {
    // Block operation or prompt user
}
```

**AI Optimization Notes:**
- All functions have PURPOSE comments
- Detection pipeline documented step-by-step
- Test vectors in `scanner_test.go`

---

### `internal/system/` - File Operations

**Responsibility:** Secure file system operations with path validation.

**Key Functions:**
- `IsPathSafe(path)` - Validate path is within home directory
- `ValidateSymlinkTarget(linkPath)` - Validate symlink targets
- `MovePath(src, dst)` - Secure file/directory move
- `CopyFile(src, dst)` - Secure file copy with permissions

**Security Model:**
```
User Input → IsPathSafe() → File Operation
              ↓
         Reject if:
         - Empty path
         - Null bytes
         - Path traversal (..)
         - Outside home directory
```

**AI Optimization Notes:**
- Attack vectors documented in comments
- Security constraints explicit
- Usage examples provided

---

### `internal/git/` - Git Operations

**Responsibility:** Git command execution with input sanitization.

**Key Types:**
- `Client` interface - Git operations abstraction
- `OSClient` struct - OS command implementation
- `GitError` struct - Sanitized error type

**Security Features:**
```go
// Git URL validation
validateGitURL(url) {
    // Reject shell metacharacters: ; | & $ ` \ !
    // Validate format: https://, git@, ssh://
    // Parse SSH URLs for username injection
}

// Error sanitization
runCmd() {
    stderr := sanitizeSensitiveData(stderr)  // Redact secrets
    args := redactArgs(args)                  // Redact URLs with credentials
    return &GitError{Args: redacted, Stderr: sanitized}
}
```

**AI Optimization Notes:**
- Custom error type for consistent handling
- Sanitization functions documented
- Command injection prevention explained

---

### `internal/backup/` - Backup System

**Responsibility:** Snapshot creation and restoration.

**Key Functions:**
- `CreateSnapshot(cfg)` - Create timestamped backup
- `CreateSnapshotWithContext(ctx, cfg)` - Cancellable backup
- `RestoreBackup(cfg, timestamp)` - Restore from backup
- `ListBackups(cfg)` - List available backups

**Context Usage Pattern:**
```go
func CreateSnapshotWithContext(ctx context.Context, cfg *config.Config) (string, error) {
    for name := range cfg.GetPaths() {
        select {
        case <-ctx.Done():
            os.RemoveAll(backupDir)  // Cleanup partial backup
            return "", ctx.Err()
        default:
            // Continue processing
        }
    }
}
```

**AI Optimization Notes:**
- Context cancellation pattern documented
- Cleanup on cancellation explicit
- Usage examples in comments

---

### `internal/cmd/` - CLI Commands

**Responsibility:** User-facing command implementations.

**Pre-Flight Check Pattern:**
```go
// Run before all operations
preflightResults := RunPreflightChecks(cfg, homeDir)
PrintPreflightResults(preflightResults)

// Check for unfixable issues
criticalFailed := false
for _, r := range preflightResults {
    if !r.Passed && !r.CanFix {
        criticalFailed = true
        break
    }
}
if criticalFailed {
    return fmt.Errorf("pre-flight checks failed")
}

// Offer to fix fixable issues
if hasIssues {
    if err := HandlePreflightFailure(preflightResults, cfgPath, cfg, homeDir); err != nil {
        return err
    }
}
```

**AI Optimization Notes:**
- Consistent command structure
- Error messages include suggestions
- Interactive flows documented

---

## Error Handling Patterns

### Structured Error Types

```go
// GitError - Sanitized Git command errors
type GitError struct {
    Args   []string  // Redacted arguments
    Stderr string    // Sanitized stderr
    Err    error     // Underlying error
}

func (e *GitError) Error() string {
    return fmt.Sprintf("git %v failed: %v (stderr: %s)", e.Args, e.Err, e.Stderr)
}

func (e *GitError) Unwrap() error {
    return e.Err  // Enables errors.Is/As
}
```

### Error Message Guidelines

**Good Error Messages:**
```go
// Include context
return fmt.Errorf("failed to clone: %w", err)

// Include suggestions
return fmt.Errorf("dotfile '%s' is not tracked\n\nDid you mean '%s'?\n\nTracked dotfiles:\n  %s", 
    name, suggestion, strings.Join(names, "\n  "))

// Include constraints
return fmt.Errorf("path must be within home directory (%s)", path)
```

**AI Parsing Hints:**
- Error messages start with operation context
- Suggestions follow newline separators
- Lists use consistent formatting

---

## Testing Patterns

### Table-Driven Tests

```go
func TestScanner(t *testing.T) {
    tests := []struct {
        name           string
        fileName       string
        fileContent    string
        expectFindings int
        expectRisk     RiskLevel
        expectPass     bool
    }{
        {
            name:           "AWS key in env file",
            fileName:       "config.env",
            fileContent:    "AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI7MDENG/bPxRfiCYEXAMPLEKEY",
            expectFindings: 1,
            expectRisk:     RiskLevelCritical,
            expectPass:     false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Case Documentation

```go
// TEST CASES for ScanFile:
//   - Empty file → Passed=true, Findings=0
//   - File with AWS key → Passed=false, Risk=Critical
//   - Binary file with ScanBinaryFiles=false → Skipped, Passed=true
//   - File >1MB → Uses streaming, same detection accuracy
```

---

## Configuration Patterns

### TOML Configuration Structure

```toml
version = 1
remote_url = "git@github.com:user/dotfiles.git"
local_path = "~/.local/share/n0man/store"

[settings]
housekeeping_max_backups = 5
run_bootstrap_after_init = true

[security]
enabled = true
scan_content = true
exclude_patterns = true
sensitivity = "medium"  # low, medium, high, paranoid
fail_on_secrets = true
interactive = true

[security.content_scan]
entropy_threshold = 4.5
min_secret_length = 20
max_file_size = 10485760  # 10MB
scan_binary_files = false
context_window = 50

[security.allowlist]
patterns = ["*test*", "*example*"]
files = ["demo_config.yaml"]

[dotfiles]
vim = "~/.vimrc"
nvim = "~/.config/nvim"

[dotfiles.ignores]
nvim = ["*.swap", "backup/"]

[overrides]
"work-laptop" = { "ssh" = "~/.ssh/config_work" }
```

### Configuration Access Pattern

```go
// Type-safe accessors
func (c *Config) GetPaths() map[string]string {
    // Converts map[string]any to map[string]string
    // Handles TOML unmarshaling quirks
}

func (c *Config) GetTargetPath(name string) string {
    // Applies host-specific overrides
    hostname, _ := os.Hostname()
    if override, ok := c.Overrides[hostname][name]; ok {
        return override
    }
    return c.GetPaths()[name]
}
```

---

## Security Model

### Defense in Depth

```
Layer 1: Input Validation
  └─ IsPathSafe() - Validate all paths
  
Layer 2: Command Sanitization
  └─ validateGitURL() - Reject shell metacharacters
  
Layer 3: Output Sanitization
  └─ sanitizeSensitiveData() - Redact secrets from errors
  
Layer 4: File Permissions
  └─ 0600 for files, 0700 for directories
```

### Security Constraints

```go
// CONSTRAINT: Path validation cannot be bypassed in production
// Only N0MAN_ALLOW_OUTSIDE_HOME=true allows outside paths
// This is intentional - security over convenience

// CONSTRAINT: Git URLs are validated before any command execution
// This prevents command injection via malicious URLs
// Valid formats: https://, git@, ssh://, file://

// CONSTRAINT: All file operations use secure permissions
// Files: 0600 (user read/write)
// Dirs:  0700 (user read/write/execute)
```

---

## Performance Characteristics

### Memory Usage

| Operation | Memory Pattern | Optimization |
|-----------|----------------|--------------|
| `ScanFile` | O(file_size) | Full file in memory |
| `scanFileStreaming` | O(batch_size) | 100-line batches |
| `CreateSnapshot` | O(1) per file | Streams file content |
| `RestoreBackup` | O(1) per file | Streams file content |

### Thresholds

```go
// Large file threshold for streaming
const LargeFileThreshold = 1024 * 1024  // 1MB

// Max file size for security scanning
MaxFileSize = 10 * 1024 * 1024  // 10MB (configurable)

// Batch size for streaming
BatchLines = 100  // Lines per batch
```

---

## Common AI Tasks

### Adding a New Secret Pattern

1. **Add pattern to `internal/security/detector.go`:**
```go
// In compilePatterns()
SecretTypeNewPattern: `(pattern-regex-here)`,
```

2. **Add test case to `internal/security/scanner_test.go`:**
```go
{
    name:           "New pattern detection",
    fileName:       "test.txt",
    fileContent:    "test-value",
    expectFindings: 1,
    expectRisk:     RiskLevelHigh,
    expectPass:     false,
},
```

3. **Update documentation in `docs/guides/security.md`**

### Adding a New Command

1. **Create `internal/cmd/newcmd.go`:**
```go
var newCmd = &cobra.Command{
    Use:   "newcmd",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Pre-flight checks
        preflightResults := RunPreflightChecks(cfg, homeDir)
        // ... implementation
    },
}

func init() {
    rootCmd.AddCommand(newCmd)
}
```

2. **Add documentation to `docs/commands/newcmd.md`**

3. **Add tests to `internal/cmd/newcmd_test.go`**

---

## Debugging Tips

### Enable Verbose Output

```bash
# Run with verbose error messages
n0man sync 2>&1 | tee debug.log

# Check pre-flight issues
n0man doctor --fix
```

### Common Issues

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| "Path must be within home directory" | File outside `~` | Move file to home or create symlink |
| "Git URL contains invalid character" | Shell metachar in URL | Remove `;`, `|`, `&`, etc. |
| "Security scan failed" | Secret detected | Review with `n0man security scan` |
| "Pre-flight checks failed" | Broken symlinks | Run `n0man doctor --fix` |

---

## See Also

- [Architecture Overview](docs/guides/core-concepts.md)
- [Security Guide](docs/guides/security.md)
- [Configuration Reference](docs/references/configuration.md)
- [Troubleshooting Guide](docs/guides/troubleshooting.md)
