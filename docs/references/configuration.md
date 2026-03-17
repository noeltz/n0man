# Configuration

n0man is configured through a single TOML file located at `~/.config/n0man/n0man.toml`. This file controls all aspects of dotfiles management including paths, security settings, and behavior.

## File Location

The configuration file is stored at:

```
~/.config/n0man/n0man.toml
```

This follows the XDG Base Directory specification for configuration files.

## Default Configuration

When you run `n0man init`, a default configuration is created:

```toml
version = 1

[settings]
housekeeping_max_backups = 5
run_bootstrap_after_init = true

[security]
enabled = true
scan_content = true
exclude_patterns = true
sensitivity = "medium"
fail_on_secrets = true
interactive = true

[security.content_scan]
entropy_threshold = 4.5
min_secret_length = 20
max_file_size = 10485760
scan_binary_files = false
context_window = 50

[dotfiles]
```

## Configuration Options

### Top-Level Options

#### `version` (integer, required)

Configuration format version. Currently always `1`.

```toml
version = 1
```

#### `remote_url` (string, optional)

Git repository URL for remote synchronization.

```toml
remote_url = "git@github.com:user/dotfiles.git"
```

- **SSH URL**: `git@github.com:user/dotfiles.git`
- **HTTPS URL**: `https://github.com/user/dotfiles.git`
- **Omitted**: Local-only mode (no sync to remote)

#### `local_path` (string, optional)

Path to the n0man store directory.

```toml
local_path = "~/.local/share/n0man/store"
```

- Default: `~/.local/share/n0man/store`
- Can be absolute path or with `~` expansion
- Must exist or be creatable

### Settings Section

#### `[settings]`

General n0man behavior settings.

```toml
[settings]
housekeeping_max_backups = 5
run_bootstrap_after_init = true
```

#### `housekeeping_max_backups` (integer, default: 5)

Maximum number of backup snapshots to retain.

```toml
housekeeping_max_backups = 5
```

- **0**: Never clean up backups (unlimited)
- **5**: Keep only the 5 most recent backups
- **10**: Keep only the 10 most recent backups

Older backups are automatically deleted after `sync` operations.

#### `run_bootstrap_after_init` (boolean, default: true)

Run bootstrap script after initialization.

```toml
run_bootstrap_after_init = true
```

If `true`, n0man looks for and runs `bootstrap.sh` in the store after `init` or `clone`.

### Security Section

#### `[security]`

Security scanning configuration.

```toml
[security]
enabled = true
scan_content = true
exclude_patterns = true
sensitivity = "medium"
fail_on_secrets = true
interactive = true
```

#### `enabled` (boolean, default: true)

Enable or disable all security scanning.

```toml
enabled = true
```

When `false`, all security scanning is skipped for `add` and `sync` operations.

#### `scan_content` (boolean, default: true)

Scan file content for secrets (not just file patterns).

```toml
scan_content = true
```

- `true`: Analyzes file content with regex and entropy detection
- `false`: Only checks file name patterns (`.env`, `.pem`, etc.)

#### `exclude_patterns` (boolean, default: true)

Use pattern-based exclusion for high-risk files.

```toml
exclude_patterns = true
```

When `true`, files matching high-risk patterns (`.env`, SSH keys, etc.) are flagged.

#### `sensitivity` (string, default: "medium")

Overall sensitivity level for secret detection.

```toml
sensitivity = "medium"
```

Values:
- `"low"`: Fewer false positives, may miss some secrets
- `"medium"`: Balanced (recommended)
- `"high"`: More sensitive, more false positives
- `"paranoid"`: Maximum sensitivity, many false positives

Affects entropy threshold and pattern matching.

#### `fail_on_secrets` (boolean, default: true)

Block operations when secrets are detected.

```toml
fail_on_secrets = true
```

- `true`: Blocks `add` and `sync` if secrets found
- `false`: Warns but allows operation (not recommended)

#### `interactive` (boolean, default: true)

Prompt user when secrets are detected.

```toml
interactive = true
```

When `true` and `fail_on_secrets = true`, prompts:
```
? Potential secrets found. Do you want to continue? (y/N)
```

### Security Content Scan Section

#### `[security.content_scan]`

Advanced content scanning parameters.

```toml
[security.content_scan]
entropy_threshold = 4.5
min_secret_length = 20
max_file_size = 10485760
scan_binary_files = false
context_window = 50
```

#### `entropy_threshold` (float, default: 4.5)

Shannon entropy threshold for secret detection.

```toml
entropy_threshold = 4.5
```

Values range from 0.0 to 8.0:
- **3.5**: Low threshold (more detection, more false positives)
- **4.5**: Medium threshold (balanced)
- **5.5**: High threshold (fewer false positives, may miss secrets)

Lower = more sensitive.

#### `min_secret_length` (integer, default: 20)

Minimum string length for entropy analysis.

```toml
min_secret_length = 20
```

Strings shorter than this are not analyzed for entropy.

#### `max_file_size` (integer, default: 10485760)

Maximum file size to scan in bytes (default: 10MB).

```toml
max_file_size = 10485760
```

Larger files are skipped to avoid performance issues.

#### `scan_binary_files` (boolean, default: false)

Scan binary files for secrets.

```toml
scan_binary_files = false
```

- `false`: Skip binary files
- `true`: Attempt to scan binary files (not recommended)

#### `context_window` (integer, default: 50)

Lines of context for analysis.

```toml
context_window = 50
```

Number of lines before and after a match to analyze context.

### Security Pattern Config Section

#### `[security.pattern_config]`

Custom file patterns to flag as high-risk.

```toml
[security.pattern_config]
custom = ["*.mysecret", "keys/*", "api/*.yaml"]
```

Additional patterns beyond built-in defaults.

### Security Allowlist Section

#### `[security.allowlist]`

Files and patterns that are always allowed (never flagged).

```toml
[security.allowlist]
patterns = ["*test*", "*example*", "*demo*", "*fake*"]
files = ["safe_config.yaml", "sample_keys.txt"]
```

Useful for test files, examples, or known-safe configurations.

### Dotfiles Section

#### `[dotfiles]`

Maps dotfile names to their target paths.

```toml
[dotfiles]
vim = "~/.vimrc"
nvim = "~/.config/nvim"
zsh = "~/.zshrc"
```

Each entry:
- **Key**: Name used by n0man (used with `rm`, `list`)
- **Value**: Target path where symlink is created (supports `~`)

#### `[dotfiles.ignores]`

Ignore patterns for specific dotfiles.

```toml
[dotfiles.ignores]
nvim = ["*.swap", "backup/*", "*.log"]
zsh = ["*.zwc", ".zcompdump*"]
```

These patterns are added to `.gitignore` automatically.

Format:
- **Key**: Dotfile name (must match `[dotfiles]` key)
- **Value**: Array of glob patterns

### Overrides Section

#### `[overrides]`

Host-specific dotfile path overrides.

```toml
[overrides]
"work-laptop" = { "ssh" = "~/.ssh/config_work" }
"home-server" = { "ssh" = "~/.ssh/config_personal" }
"desktop" = { "vim" = "~/.config/vim/custom.vimrc" }
```

Keys are hostnames (from `hostname` command).

Overrides are used instead of default path when:
- Machine hostname matches
- Override key exists for the dotfile

## Examples

### Minimal Configuration

```toml
version = 1

[settings]
housekeeping_max_backups = 5

[security]
enabled = true
fail_on_secrets = true
```

### Full Configuration

```toml
version = 1
remote_url = "git@github.com:user/dotfiles.git"
local_path = "~/.local/share/n0man/store"

[settings]
housekeeping_max_backups = 10
run_bootstrap_after_init = false

[security]
enabled = true
scan_content = true
exclude_patterns = true
sensitivity = "medium"
fail_on_secrets = true
interactive = true

[security.content_scan]
entropy_threshold = 4.5
min_secret_length = 20
max_file_size = 10485760
scan_binary_files = false
context_window = 50

[security.pattern_config]
custom = ["*.internal_keys"]

[security.allowlist]
patterns = ["*test*", "*example*"]
files = ["demo_config.yaml"]

[dotfiles]
vim = "~/.vimrc"
nvim = "~/.config/nvim"
zsh = "~/.zshrc"
git = "~/.gitconfig"

[dotfiles.ignores]
nvim = ["*.swap", "backup/*"]
zsh = ["*.zwc"]

[overrides]
"workstation" = { "ssh" = "~/.ssh/config_work" }
"personal-laptop" = { "ssh" = "~/.ssh/config_personal" }
```

### Local-Only Configuration

```toml
version = 1
local_path = "~/dotfiles-store"

[settings]
housekeeping_max_backups = 3

[security]
enabled = true
fail_on_secrets = true

[dotfiles]
vim = "~/.vimrc"
```

### Development Configuration (Lenient Security)

```toml
version = 1

[security]
enabled = true
sensitivity = "low"
fail_on_secrets = false
interactive = false

[security.content_scan]
entropy_threshold = 5.5
min_secret_length = 30

[security.allowlist]
patterns = ["*test*", "*dev*", "*local*"]
files = ["env.dev", "env.test"]
```

## Configuration Management

### View Configuration

```bash
cat ~/.config/n0man/n0man.toml
```

### Edit Configuration

```bash
# Use your favorite editor
vim ~/.config/n0man/n0man.toml
nano ~/.config/n0man/n0man.toml
```

### Validate Configuration

```bash
# Most commands will fail if configuration is invalid
n0man list
```

### Backup Configuration

```bash
cp ~/.config/n0man/n0man.toml ~/.config/n0man/n0man.toml.backup
```

## Troubleshooting

### Configuration Not Found

```bash
# Initialize n0man
n0man init
```

### Invalid TOML

```bash
# Check TOML syntax
cat ~/.config/n0man/n0man.toml | toml-lint
```

### Settings Not Applied

```bash
# Verify file permissions
ls -la ~/.config/n0man/n0man.toml
# Should be readable: -rw-r--r--
```

### Security Not Working

```bash
# Verify security section exists
cat ~/.config/n0man/n0man.toml | grep -A 10 "\[security\]"
```

## Notes

- Configuration changes take effect immediately (no restart needed)
- Invalid TOML will cause commands to fail
- Paths are stored with `~` for portability
- Overrides are determined by `hostname` command
- Ignores are automatically added to `.gitignore`
- Security settings affect both `add` and `sync` operations

## Related Documentation

- [Security Scanning](security.md): Detailed security configuration
- [Commands](commands/): How commands use configuration
- [Backups](backup.md): Backup management
