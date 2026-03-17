# n0man Architecture

## Overview

n0man is a CLI tool for managing dotfiles with bidirectional synchronization, backup, and security scanning. It's built with Go using the Cobra framework for CLI commands and Bubble Tea for TUI components.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         n0man CLI                                │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ cmd/ (Cobra Commands)                                       ││
│  │  ├── init    - Initialize repository                        ││
│  │  ├── add     - Add dotfile to tracking                     ││
│  │  ├── rm      - Remove dotfile from tracking                ││
│  │  ├── sync    - Bidirectional sync                          ││
│  │  ├── status  - Show divergence status                       ││
│  │  ├── list    - List tracked dotfiles                       ││
│  │  ├── backup  - Manage backups                              ││
│  │  ├── doctor  - Health checks                                ││
│  │  └── security - Security scanning                          ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    internal/ Packages                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   config    │  │     git      │  │      security       │  │
│  │              │  │              │  │                      │  │
│  │ TOML parsing │  │ Git CLI exec │  │ Pattern matching    │  │
│  │ Path config │  │ Clone/Init   │  │ Content detection   │  │
│  │ Overrides   │  │ Add/Commit   │  │ Entropy analysis     │  │
│  └──────────────┘  │ Push/Pull    │  └──────────────────────┘  │
│                     │ HasChanges   │                             │
│                     └──────────────┘                             │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   system     │  │    backup    │  │         ui           │  │
│  │              │  │              │  │                      │  │
│  │ File ops     │  │ Snapshots    │  │ Bubble Tea TUI       │  │
│  │ Path safety  │  │ Rollback     │  │ Conflict resolution  │  │
│  │ Symlinks     │  │ Cleanup      │  │ Prompts/Styles       │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Package Responsibilities

### cmd/

The command layer using Cobra framework. Each command is a separate file that handles:
- Flag parsing and validation
- Calling appropriate internal packages
- Output formatting via ui package

### config/

- `Config` struct representing `n0man.toml`
- Load/Save TOML configuration
- GetPaths, GetTargetPath with host-specific overrides
- Default configuration values

### git/

- Wrapper around git CLI commands
- `Client` interface for testability
- URL validation to prevent command injection
- Sensitive data redaction in error messages

### security/

- **Scanner**: Orchestrates scanning operations
- **PatternMatcher**: File path pattern detection (.env, .pem, etc.)
- **Detector**: Content analysis with regex patterns
- **EntropyAnalyzer**: Shannon entropy for random token detection

### system/

- `IsPathSafe()`: Validates paths within home directory
- `ValidateSymlinkTarget()`: Prevents symlink attacks
- `MovePath()`, `CopyFile()`, `CopyDir()`: Secure file operations
- `CreateSymlink()`: Symlink creation with validation

### backup/

- `CreateSnapshot()`: Create timestamped backups
- `RestoreBackup()`: Restore from backup
- `CleanOldBackups()`: Maintain max backup count

### ui/

- **styles.go**: lipgloss styling for consistent output
- **prompt.go**: Bubble Tea model for interactive prompts
- **conflict/ui.go**: Conflict resolution TUI

## Data Flow

### Adding a Dotfile

```
user: n0man add ~/.vimrc
        │
        ▼
cmd/add.go:Validate path with system.IsPathSafe()
        │
        ▼
security.Scanner:ScanFile() for secrets
        │
        ▼
system.MovePath(): Move ~/.vimrc → store/vimrc
        │
        ▼
system.CreateSymlink(): Create ~/.vimrc → store/vimrc
        │
        ▼
config.SetPath(): Update n0man.toml
        │
        ▼
config.Save(): Persist configuration
```

### Syncing Changes

```
user: n0man sync
        │
        ▼
preflight.RunPreflightChecks(): Validate environment
        │
        ├── backup.CreateSnapshot(): Create pre-sync backup
        │
        ├── git.HasChanges(): Check for local changes
        │
        └── security.Scanner: Scan for secrets
        │
        ▼
git.Commit(): Commit local changes (if any)
        │
        ▼
git.Pull(): Pull remote changes
        │
        ▼
Handle conflicts (if any) via ui/conflict TUI
        │
        ▼
git.Push(): Push to remote
```

## Security Model

### Path Safety

All file paths are validated against the user's home directory:
1. No empty paths
2. No null bytes
3. No `..` traversal components
4. Must resolve within `$HOME`

### Symlink Protection

Before following any symlink:
1. Read symlink target
2. Resolve to absolute path
3. Validate target is within home directory

### Command Injection Prevention

Git URLs validated for shell metacharacters:
- Blocked: `; | & $ ` \ ! { } < > ( ) \n \r
- Allowed prefixes: `https://`, `http://`, `git@`, `ssh://`, `file://`

### Secret Detection

Multi-layered scanning:
1. **Pattern matching**: High-risk file names (.env, .pem, id_rsa)
2. **Content scanning**: Regex for API keys, passwords, tokens
3. **Entropy analysis**: High-entropy strings indicating random secrets

## Configuration

### n0man.toml Structure

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
sensitivity = "medium"
fail_on_secrets = true
interactive = true

[dotfiles]
vim = "~/.vimrc"
nvim = "~/.config/nvim"

[overrides]
"hostname" = { "ssh" = "~/.ssh/config_work" }
```

## File Permissions

- Configuration files: `0600` (user read/write only)
- Store directories: `0700` (user rwx only)
- Backup snapshots: `0700` (user-only access)

## Extension Points

### Adding New Commands

1. Create `internal/cmd/newcmd.go`
2. Define cobra.Command with Use, Short, RunE
3. Add to rootCmd in `init()`
4. Use existing packages for functionality

### Adding Security Patterns

1. Edit `internal/security/patterns.go`
2. Add regex pattern to `highRiskPatterns` or `mediumRiskPatterns`
3. Add test cases in `scanner_test.go`

### Adding Backup Storage Backends

1. Extend `internal/backup/backup.go`
2. Implement `BackupStorage` interface
3. Add configuration options in `config/`
