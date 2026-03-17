# Getting Started with n0man

This guide walks you through installing and setting up n0man for the first time.

## Prerequisites

- **Go 1.26.1** or later (for building from source)
- **Git** (required for version control)
- **Linux/macOS** (Windows support via WSL)

## Installation

### Option 1: Install from Source (Recommended)

```bash
go install github.com/noeltz/n0man/cmd/n0man@latest
```

### Option 2: Manual Build

```bash
# Clone the repository
git clone https://github.com/noeltz/n0man.git
cd n0man

# Build the binary
go build -o n0man ./cmd/n0man

# Install system-wide (optional)
sudo mv n0man /usr/local/bin/
```

### Option 3: Download Pre-built Binary

Download the latest release from GitHub and add to your PATH.

## Verify Installation

```bash
n0man version
```

Expected output:
```
n0man version 1.0.0
  Go version: go1.26.1
  OS/Arch: linux/amd64
```

## First-Time Setup

### Step 1: Initialize n0man

Initialize n0man in local-only mode (no remote repository):

```bash
n0man init
```

**What this does:**
- Creates configuration at `~/.config/n0man/n0man.toml`
- Creates store directory at `~/.local/share/n0man/store`
- Initializes Git repository
- Auto-configures Git user for local-only mode

**Expected output:**
```
╭──────────────╮
│  n0man init  │
╰──────────────╯

  → Initializing local repository at ~/.local/share/n0man/store
  → Configuring Git for local use...
  ✓ n0man initialized successfully

  Local-only mode active. Your dotfiles are tracked but not backed up remotely.
  Add a remote later with: n0man init git@github.com:user/dotfiles.git
```

### Step 2: Add Your First Dotfile

```bash
n0man add ~/.vimrc
```

**What this does:**
- Moves `~/.vimrc` to the store
- Creates a symlink from `~/.vimrc` to the store
- Updates configuration

**Expected output:**
```
  → Moving ~/.vimrc → ~/.local/share/n0man/store/.vimrc
  → Linking ~/.local/share/n0man/store/.vimrc → ~/.vimrc
  ✓ Added '.vimrc' (~/.vimrc)
```

### Step 3: Verify

```bash
# List tracked files
n0man list

# Check status
n0man status

# Run health checks
n0man doctor
```

## Basic Workflow

### Daily Usage

```bash
# 1. Make changes to your dotfiles
vim ~/.vimrc

# 2. Check status
n0man status

# 3. Sync changes
n0man sync
```

### Adding More Dotfiles

```bash
n0man add ~/.config/nvim
n0man add ~/.gitconfig
n0man add ~/.zshrc
```

### Syncing with Pre-Flight Checks

```bash
n0man sync
```

**Pre-flight checks verify:**
- ✅ Configuration loaded
- ✅ Store directory exists
- ✅ Git repository exists
- ✅ All symlinks valid

## Next Steps

### Configure Remote Repository (Optional)

To backup your dotfiles to a remote repository:

```bash
# Initialize with remote
n0man init git@github.com:user/dotfiles.git

# Or add remote to existing setup
cd ~/.local/share/n0man/store
git remote add origin git@github.com:user/dotfiles.git
```

### Learn More

- **[Core Concepts](core-concepts.md)** - Understand how n0man works
- **[Commands](../commands/)** - Complete command reference
- **[Security Guide](security.md)** - Security features and configuration
- **[Backup Guide](backup.md)** - Backup and recovery strategies

## Troubleshooting

### "Author identity unknown"

**Problem:** Git commit fails due to missing user configuration.

**Solution:** n0man auto-configures Git for local-only mode. For remote mode, configure Git:

```bash
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
```

### "Path must be within home directory"

**Problem:** Trying to add a file outside your home directory.

**Solution:** n0man only manages files within your home directory for security. Create a symlink:

```bash
ln -s /etc/myapp/config ~/.config/myapp-config
n0man add ~/.config/myapp-config
```

### "Git URL contains invalid character"

**Problem:** Git URL contains shell metacharacters.

**Solution:** Use standard Git URL formats:

```bash
# Valid formats
n0man init git@github.com:user/dotfiles.git
n0man init https://github.com/user/dotfiles.git
n0man init ssh://git@github.com/user/dotfiles.git
```

## Quick Reference

| Task | Command |
|------|---------|
| Initialize | `n0man init` |
| Add file | `n0man add ~/.file` |
| List files | `n0man list` |
| Sync | `n0man sync` |
| Status | `n0man status` |
| Health check | `n0man doctor` |
| Create backup | `n0man backup create` |
| Restore backup | `n0man backup rollback` |

---

**Next:** [Core Concepts](core-concepts.md) - Learn how n0man works
