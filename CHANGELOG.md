# Changelog

All notable changes to n0man will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-16

### Added

#### Core Features
- **Bidirectional synchronization** - Seamless sync between local machines and remote Git repositories
- **Backup and rollback system** - Automatic snapshot backups before sync with easy restoration
- **Host-specific overrides** - Customize dotfile paths for different machines
- **Interactive conflict resolution** - TUI-based conflict resolution during git rebase

#### Security
- **Pattern-based detection** - Identifies high-risk file patterns (.env, .pem, SSH keys, etc.)
- **Content scanning** - Detects API keys, passwords, tokens using regex patterns
- **Entropy analysis** - Shannon entropy detection for random-looking secrets
- **Path validation** - All operations restricted to home directory
- **Command injection prevention** - Git URLs validated for shell metacharacters
- **Secure file permissions** - Configuration files (0600) and directories (0700)

#### Commands
- `n0man init` - Initialize repository (local or clone remote)
- `n0man add` - Add file/directory and replace with symlink
- `n0man rm` - Stop tracking and restore original file
- `n0man sync` - Bidirectional sync with pre-flight checks
- `n0man status` - Inspect divergence between system, config, and repo
- `n0man list` - List all tracked dotfiles
- `n0man backup` - Manage dotfile backups interactively
- `n0man backup create` - Create manual backup snapshot
- `n0man backup rollback` - Restore from latest backup
- `n0man doctor` - Run comprehensive health checks with interactive fixes
- `n0man security scan` - Scan dotfiles for sensitive information
- `n0man self-update` - Update n0man to latest version
- `n0man version` - Print version information
- `n0man completion` - Generate shell autocompletion scripts

#### System
- **Signal handling** - Graceful shutdown on SIGINT/SIGTERM
- **Context cancellation** - Long-running operations can be cancelled
- **TTY detection** - Interactive mode adapts to terminal availability
- **Error sanitization** - Sensitive data redacted from error messages

### Configuration

The following configuration options are available in `n0man.toml`:

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

### Dependencies

- **Cobra** (v1.10.2) - CLI framework
- **Bubble Tea** (v1.3.10) - TUI framework
- **Lipgloss** (v1.1.0) - Terminal styling
- **go-toml** (v2.2.4) - TOML parsing

---

## [Unreleased]

### Planned Features

- [ ] Enhanced backup encryption
- [ ] Plugin system for custom commands
- [ ] Cloud storage integration for backups
- [ ] Web UI for remote management

### Known Issues

- [ ] Large repositories may experience slow sync times
- [ ] Some edge cases in entropy detection may produce false positives

---

## Upgrade Guide

### Upgrading from 0.x to 1.0.0

1. **Backup your data** - Run `n0man backup create` before upgrading
2. **Update the binary** - Run `n0man self-update` or reinstall
3. **Verify configuration** - Check `~/.config/n0man/n0man.toml` is intact
4. **Test sync** - Run `n0man sync` to verify functionality

### Configuration Migration

The configuration format changed in 1.0.0. A new `version` field is required:

```toml
version = 1  # Required in 1.0.0+
```

---

## Reporting Problems

If you encounter issues:

1. Check the [troubleshooting guide](docs/guides/troubleshooting.md)
2. Run `n0man doctor` to diagnose problems
3. Search existing [GitHub Issues](https://github.com/noeltz/n0man/issues)
4. Open a new issue with reproduction steps
