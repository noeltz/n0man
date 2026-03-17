# Core Concepts

This guide explains how n0man works and its core architectural concepts.

## Overview

n0man is a dotfiles manager that provides:

- **Bidirectional synchronization** between local machines and remote repositories
- **Automatic backups** before sync operations
- **Security scanning** to prevent committing sensitive data
- **Symlink management** for clean dotfile organization

**Acknowledgments:** n0man was inspired by the innovative work on lnk, gart, and paw. Thank you to the creators of these projects for pioneering modern dotfiles management approaches.

## Architecture

### Directory Structure

```
~/.local/share/n0man/store/       # Main store directory (0700)
  .backups/                        # Backup snapshots (0700)
  .git/                            # Git repository
  .gitignore                       # Generated gitignore
  .vimrc                           # Actual dotfile content (0600)
  nvim/                            # Actual dotfile content (0700)
  ...

~/.config/n0man/
  n0man.toml                       # Configuration file (0600)

~/.                              # Your home directory
  .vimrc → ~/.local/share/n0man/store/.vimrc   # Symlink
  .config/nvim → ~/.local/share/n0man/store/nvim  # Symlink
```

### Key Components

#### 1. Store Directory

The central repository for all tracked dotfiles:
- Located at `~/.local/share/n0man/store` (configurable)
- Git repository for version control
- Secure permissions (0700 for directories, 0600 for files)
- Contains actual dotfile content (not symlinks)

#### 2. Configuration File

TOML configuration at `~/.config/n0man/n0man.toml`:
- Tracks dotfile mappings
- Configures security settings
- Stores remote repository URL
- Defines backup retention

#### 3. Symlinks

Symbolic links in your home directory:
- Point to files in the store
- Allow applications to find dotfiles normally
- Automatically created/updated by n0man

#### 4. Backups

Timestamped snapshots in `.backups/`:
- Created automatically before sync
- Can be created manually
- Used for rollback and recovery

## How n0man Works

### Adding a Dotfile

```bash
n0man add ~/.vimrc
```

**Process:**
1. **Validate path** - Ensure file is within home directory
2. **Security scan** - Check for secrets (unless `--no-security`)
3. **Move file** - Move `~/.vimrc` to `~/.local/share/n0man/store/.vimrc`
4. **Create symlink** - Link `~/.vimrc` → `~/.local/share/n0man/store/.vimrc`
5. **Update config** - Add mapping to `n0man.toml`

### Syncing Changes

```bash
n0man sync
```

**Process:**
1. **Pre-flight checks** - Validate environment
2. **Create backup** - Snapshot current state
3. **Check local changes** - Git status
4. **Security scan** - Scan for secrets
5. **Commit local** - Git commit if changes exist
6. **Pull remote** - Git pull with rebase
7. **Push remote** - Git push (if remote configured)

### Restoring from Backup

```bash
n0man backup rollback
```

**Process:**
1. **Find latest backup** - List `.backups/` directory
2. **Copy files** - Restore from backup to store
3. **Recreate symlinks** - Ensure symlinks point to restored files

## Security Model

### Path Validation

All file operations are restricted to your home directory:
- Prevents accessing sensitive system files
- Validates both source and destination paths
- Can be bypassed with `N0MAN_ALLOW_OUTSIDE_HOME=true` (not recommended)

### Command Injection Protection

Git URLs are validated before use:
- Rejects shell metacharacters (`;`, `|`, `&`, `$`, etc.)
- Validates URL format (https://, git@, ssh://)
- Prevents command injection attacks

### Secret Detection

Multiple detection methods:
- **Pattern-based** - High-risk file patterns (`.env`, `.pem`, SSH keys)
- **Content scanning** - Regex patterns for API keys, passwords, tokens
- **Entropy analysis** - High-entropy string detection

### File Permissions

Secure by default:
- Configuration files: `0600` (user read/write only)
- Directories: `0700` (user read/write/execute only)
- Backup snapshots: `0700` (user-only access)

## Backup System

### Automatic Backups

Created before every `sync` operation:
- Timestamped directory in `.backups/`
- Contains all tracked dotfiles
- Used for rollback if sync fails

### Manual Backups

Created on-demand:
```bash
n0man backup create
```

### Backup Retention

Automatic cleanup based on `housekeeping_max_backups`:
- Default: Keep 5 most recent backups
- Older backups automatically deleted after sync
- Set to `0` to keep all backups

### Store Recovery

If store directory is missing:
```bash
n0man backup rollback
```

Offers recovery options:
1. **Restore from backup** - Recreate from snapshot
2. **Re-clone from remote** - Fresh clone from remote
3. **Reinitialize** - Fresh Git repository

## Configuration System

### Configuration Loading

1. Load from `~/.config/n0man/n0man.toml`
2. Apply defaults for missing values
3. Apply host-specific overrides

### Host-Specific Overrides

Different machines can have different dotfile paths:

```toml
[overrides]
"work-laptop" = { "ssh" = "~/.ssh/config_work" }
"home-server" = { "ssh" = "~/.ssh/config_personal" }
```

Override is determined by hostname (`hostname` command).

### Ignored Patterns

Per-dotfile ignore patterns:

```toml
[dotfiles.ignores]
nvim = ["*.swap", "backup/"]
```

Added to `.gitignore` automatically.

## Error Handling

### Pre-Flight Checks

Run before `sync` to catch issues early:
- Configuration loaded
- Store directory exists
- Git repository exists
- All symlinks valid

Auto-fix available with interactive prompts.

### Graceful Shutdown

Handles SIGINT/SIGTERM:
- Cancels in-progress operations
- Cleans up partial work
- Exits with code 130

### Error Messages

User-friendly with suggestions:
```
Error: dotfile 'bashrc' is not tracked

Tracked dotfiles:
  .bashrc
  .vimrc

Did you mean '.bashrc'?

Or use 'n0man add' to track a new file.
```

## Comparison with Other Tools

| Feature | n0man | GNU Stow | Chezmoi |
|---------|-------|----------|---------|
| Bidirectional sync | ✅ | ❌ | ✅ |
| Automatic backups | ✅ | ❌ | ❌ |
| Security scanning | ✅ | ❌ | ❌ |
| Interactive fixes | ✅ | ❌ | ❌ |
| Pre-flight checks | ✅ | ❌ | ❌ |
| Git-free onboarding | ✅ | N/A | ❌ |

## Next Steps

- **[Getting Started](getting-started.md)** - Install and setup
- **[Security Guide](security.md)** - Deep dive into security features
- **[Backup Guide](backup.md)** - Backup strategies and recovery
- **[Configuration Guide](../references/configuration.md)** - Complete config reference
