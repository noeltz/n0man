# n0man Documentation

Welcome to the n0man documentation wiki. This documentation covers all aspects of n0man, a safe and simple dotfiles manager with backup and security features.

## Quick Start

New to n0man? Start here:

1. **[Getting Started](guides/getting-started.md)** - Installation and first steps
2. **[Core Concepts](guides/core-concepts.md)** - How n0man works
3. **[Command Reference](commands/)** - All available commands

## Documentation Structure

### 📚 Guides

Conceptual guides and how-to documentation:

| Guide | Description |
|-------|-------------|
| [Getting Started](guides/getting-started.md) | Installation and quick start guide |
| [Core Concepts](guides/core-concepts.md) | How n0man works, architecture overview |
| [Security Guide](guides/security.md) | Understanding security scanning and protection |
| [Backup Guide](guides/backup.md) | Backup and rollback strategies |
| [Configuration Guide](guides/configuration.md) | Complete configuration reference |

### 📖 Command Reference

Detailed documentation for each command:

| Command | Description |
|---------|-------------|
| [`init`](commands/init.md) | Initialize n0man (local or clone remote) |
| [`add`](commands/add.md) | Add a file/directory and replace with symlink |
| [`rm`](commands/rm.md) | Stop tracking and restore original file |
| [`sync`](commands/sync.md) | Bidirectional sync with pre-flight checks |
| [`status`](commands/status.md) | Inspect divergence between system, config, and repo |
| [`list`](commands/list.md) | List all tracked dotfiles |
| [`backup`](commands/backup.md) | Manage dotfile backups interactively |
| [`doctor`](commands/doctor.md) | Run comprehensive health checks |
| [`security scan`](commands/security.md) | Scan dotfiles for sensitive information |
| [`self-update`](commands/self-update.md) | Update n0man to latest version |
| [`version`](commands/version.md) | Print version information |
| [`completion`](commands/completion.md) | Generate shell autocompletion scripts |

### 📋 References

Technical references and specifications:

| Reference | Description |
|-----------|-------------|
| [Configuration Reference](references/configuration.md) | Complete `n0man.toml` reference |
| [Exit Codes](references/exit-codes.md) | All exit codes and their meanings |
| [File Permissions](references/file-permissions.md) | Security and file permissions |
| [Environment Variables](references/environment-variables.md) | Environment variable reference |

## Common Tasks

### First Time Setup

```bash
# Initialize n0man (local-only mode)
n0man init

# Add your first dotfile
n0man add ~/.vimrc

# Sync changes
n0man sync
```

### Daily Usage

```bash
# Check status before syncing
n0man status

# Sync with pre-flight checks
n0man sync

# Run health checks
n0man doctor
```

### Backup and Recovery

```bash
# Create manual backup
n0man backup create

# Restore from latest backup
n0man backup rollback

# Interactive backup management
n0man backup
```

### Security

```bash
# Scan for secrets
n0man security scan

# Add file without security scan (if you know it's safe)
n0man add ~/.config/app --no-security
```

## Troubleshooting

Common issues and solutions:

| Issue | Solution |
|-------|----------|
| "Author identity unknown" | Run `n0man init` for auto-configuration or configure Git manually |
| "Path must be within home directory" | Files must be in your home directory for security |
| "Git URL contains invalid character" | Remove shell metacharacters from Git URL |
| TUI errors in CI/CD | Use `--conflict-strategy` flag for non-interactive mode |
| Broken symlinks | Run `n0man doctor --fix` to repair |

See [Troubleshooting Guide](guides/troubleshooting.md) for more help.

## Getting Help

- **Documentation**: You're reading it!
- **Issues**: Report bugs on GitHub
- **Discussions**: Ask questions in GitHub Discussions

## Documentation Conventions

This documentation uses the following conventions:

### Command Syntax

```bash
command [subcommand] [flags] <argument>
```

- `[square brackets]` - Optional
- `<angle brackets>` - Required argument
- `...` - Repeatable

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `130` | Interrupted (Ctrl+C) |

### File Permissions

| Permission | Octal | Description |
|------------|-------|-------------|
| `rw-------` | `0600` | User read/write only (files) |
| `rwx------` | `0700` | User read/write/execute (directories) |

## Version Information

**Documentation Version:** 1.0.0  
**Last Updated:** March 16, 2026  
**n0man Version:** 1.0.0

---

**n0man** - A safe and simple dotfiles manager with bidirectional sync, backup, and security features.
