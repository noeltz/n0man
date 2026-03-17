# n0man

A safe and simple dotfiles manager that provides real bidirectional synchronization, backup, and security for your dotfiles.

**Acknowledgments:** n0man was inspired by the innovative work on lnk, gart, and paw. Thank you to the creators of these projects for pioneering modern dotfiles management approaches.

## Features

- **Centralized Configuration**: Single TOML configuration file (`n0man.toml`) for all dotfiles
- **Bidirectional Synchronization**: Seamlessly sync changes between local machines and remote repositories
- **Security Scanning**: Built-in secret detection to prevent committing sensitive data
- **Backup & Rollback**: Automatic snapshot backups before sync operations with easy rollback
- **Symlink Management**: Automatic file-to-symlink conversion for clean dotfile organization
- **Host-Specific Overrides**: Customize dotfile paths for different machines
- **Health Checks**: Comprehensive diagnostics via the `doctor` command
- **Secure by Default**: User-only file permissions (0600/0700), path validation, command injection protection

## Installation

### From Source

```bash
go install github.com/noeltz/n0man/cmd/n0man@latest
```

### Manual Build

```bash
git clone https://github.com/noeltz/n0man.git
cd n0man
go build -o n0man ./cmd/n0man
sudo mv n0man /usr/local/bin/
```

## Quick Start

### 1. Initialize n0man

Initialize a local-only repository:
```bash
n0man init
```

Or clone an existing remote repository:
```bash
n0man init git@github.com:user/dotfiles.git
```

### 2. Add Your First Dotfile

```bash
n0man add ~/.vimrc
n0man add ~/.config/nvim
n0man add ~/.gitconfig
```

### 3. Sync Changes

```bash
n0man sync
```

This will:
- Create a pre-sync backup snapshot
- Commit local changes
- Pull remote changes (if remote configured)
- Push to remote (if configured)

## Configuration

The main configuration file is located at `~/.config/n0man/n0man.toml`. Example:

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

[dotfiles.ignores]
nvim = ["*.swap", "backup/"]

[overrides]
"laptop" = { "ssh" = "~/.ssh/config_laptop" }
```

See [Configuration Documentation](docs/configuration.md) for detailed options.

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize n0man (local or clone remote) with automatic Git configuration |
| `add` | Add a file/directory and replace with symlink |
| `rm` | Stop tracking and restore original file |
| `sync` | Bidirectional sync with pre-flight checks and automatic backups |
| `status` | Inspect divergence between system, config, and repo |
| `list` | List all tracked dotfiles |
| `backup` | Manage dotfile backups interactively |
| `backup create` | Create manual backup snapshot |
| `backup rollback` | Restore from latest backup (includes store recovery) |
| `doctor` | Run comprehensive health checks with interactive fixes |
| `security scan` | Scan dotfiles for sensitive information |
| `self-update` | Update n0man to latest version |
| `version` | Print version information |
| `completion` | Generate shell autocompletion scripts |

### Global Flags

| Flag | Description |
|------|-------------|
| `-h, --help` | Help for any command |
| `-v, --version` | Print version information |

### Sync Flags

| Flag | Description |
|------|-------------|
| `--no-security` | Skip security scanning before commit |
| `--conflict-strategy` | Non-interactive conflict resolution (`keep-local`, `keep-remote`, `abort`) |

### Doctor Flags

| Flag | Description |
|------|-------------|
| `-f, --fix` | Automatically fix issues without prompting |

See [Commands Documentation](docs/commands/) for detailed usage.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error or command failure |
| 1 | Security scan found issues (for `security scan`) |
| 130 | Interrupted by user (SIGINT/Ctrl+C) |

## Security

n0man includes comprehensive security features to protect your dotfiles and system:

### Secret Detection
- **Pattern-Based Detection**: Identifies high-risk file patterns (`.env`, `.pem`, SSH keys, etc.)
- **Content Scanning**: Detects API keys, passwords, tokens using regex patterns
- **Entropy Analysis**: Detects high-entropy strings that may indicate secrets
- **Interactive Mode**: Prompts before blocking operations when issues found

### System Security
- **Command Injection Protection**: Git URLs and file paths are validated to prevent injection attacks
- **Path Traversal Protection**: All file operations restricted to home directory
- **Secure File Permissions**: Configuration files (0600) and directories (0700) are user-only
- **Stderr Sanitization**: Sensitive data redacted from error messages
- **Graceful Shutdown**: Handles SIGINT/SIGTERM for clean interruption

### Security Options
- `--no-security` flag: Skip security scanning when you know files are safe
- Configurable sensitivity levels: `low`, `medium`, `high`, `paranoid`
- Custom allowlists for known-safe files and patterns

See [Security Documentation](docs/security.md) for details on configuring security scans.

## Backup & Rollback

n0man automatically creates encrypted snapshots before `sync` operations. You can also create manual backups:

```bash
n0man backup create
```

Restore from backups interactively:
```bash
n0man backup
```

Rollback to the latest backup directly:
```bash
n0man backup rollback
```

### Backup Features
- **Automatic**: Created before every sync operation
- **Timestamped**: Each backup named with timestamp (e.g., `20260312_162413`)
- **Secure**: Backup directories use 0700 permissions (user-only)
- **Cleanup**: Old backups automatically removed based on `housekeeping_max_backups` setting
- **Cancelable**: Long-running backups can be interrupted with Ctrl+C

See [Backup Documentation](docs/backup.md) for details.

## Host-Specific Overrides

Override dotfile paths for specific hosts:

```toml
[overrides]
"work-laptop" = { "ssh" = "~/.ssh/config_work" }
"home-server" = { "ssh" = "~/.ssh/config_personal" }
```

The override is determined by your system's hostname (`hostname` command).

## Directory Structure

```
~/.local/share/n0man/store/       # Main store directory (0700)
  .backups/                        # Backup snapshots (0700)
  .git/                            # Git repository
  .gitignore                       # Generated gitignore
  vimrc                            # Actual dotfile content (0600)
  nvim/                            # Actual dotfile content (0700)
  ...

~/.config/n0man/
  n0man.toml                       # Configuration file (0600)
```

### File Permissions
- **Configuration files**: 0600 (user read/write only)
- **Directories**: 0700 (user read/write/execute only)
- **Backup snapshots**: 0700 (user-only access)

This ensures your dotfiles and configuration are protected from other users on the system.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Troubleshooting

### Git Configuration Required
If you see "Author identity unknown" errors during sync, configure Git:
```bash
git config --global user.email "you@example.com"
git config --global user.name "Your Name"
```

### TTY Errors in CI/CD
The interactive conflict resolution TUI requires a terminal. In non-interactive environments:
- Use `--no-security` flag to skip interactive prompts
- Resolve conflicts manually in the store directory

### Path Outside Home Directory
n0man only manages files within your home directory for security. To manage files elsewhere:
- Create symlinks from your home directory to the target location
- Or set `N0MAN_ALLOW_OUTSIDE_HOME=true` environment variable (not recommended for production)

### Command Injection Protection
If you see "Git URL contains invalid character" errors:
- Remove shell metacharacters (`;`, `|`, `&`, `$`, etc.) from the URL
- Use standard Git URL formats: `https://`, `git@`, `ssh://`

## License

MIT License - see [LICENSE](LICENSE) for details.
