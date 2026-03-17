# `n0man init`

Initialize n0man and optionally clone a remote dotfiles repository.

## Usage

```bash
n0man init [remote_url]
```

## Arguments

- `remote_url` (optional) - Git repository URL to clone (e.g., `git@github.com:user/dotfiles.git`)

## Description

The `init` command sets up n0man for first-time use. It can either:

1. **Local-Only Mode**: Create a new local Git repository with automatic Git user configuration
2. **Remote Mode**: Clone an existing remote repository and configure it for sync

When run without arguments, `init` creates a local repository at `~/.local/share/n0man/store/`, initializes it as a Git repository, and automatically configures Git user.name and user.email for local-only mode.

When run with a `remote_url`, it clones the repository and configures it as the remote for sync operations.

The command creates a default configuration file at `~/.config/n0man/n0man.toml` with sensible defaults.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully initialized |
| 1 | Failed to clone or initialize |
| 1 | Failed to save configuration |
| 1 | Invalid Git URL (contains shell metacharacters) |

## Examples

### Local-Only Initialization (Git-Free Onboarding)

```bash
$ n0man init
n0man init
  Initializing local repository at ~/.local/share/n0man/store
  Configuring Git for local use...
✅ n0man initialized successfully

  Local-only mode active. Your dotfiles are tracked but not backed up remotely.
  Add a remote later with: n0man init git@github.com:user/dotfiles.git
```

**Note**: Git user.name and user.email are automatically configured for local-only mode.

### Initialize with Remote Repository

```bash
$ n0man init git@github.com:user/dotfiles.git
n0man init
  Cloning git@github.com:user/dotfiles.git → ~/.local/share/n0man/store
✅ n0man initialized successfully
```

**Note**: For remote mode, ensure Git user.name and user.email are configured globally.

### Invalid Git URL (Security Protection)

```bash
$ n0man init 'git@github.com:user;rm -rf /'
n0man init
Error: invalid Git URL: Git URL contains invalid character: ;
```

## What It Does

1. Creates or loads configuration at `~/.config/n0man/n0man.toml`
2. Sets default store path to `~/.local/share/n0man/store`
3. If `remote_url` provided:
   - Validates Git URL (rejects shell metacharacters)
   - Clones the repository to the store path
   - Sets `remote_url` in configuration
4. If no `remote_url` provided (local-only mode):
   - Initializes a new Git repository at the store path
   - Automatically configures Git user.name="n0man" and user.email="n0man@localhost"
5. Saves configuration with secure permissions (0600)

## After Initialization

After running `init`, you can:

- Add dotfiles: `n0man add ~/.vimrc`
- List tracked files: `n0man list`
- Sync to remote: `n0man sync` (if remote configured)
- Run health checks: `n0man doctor`

## Notes

- The store directory is created with secure permissions (0700)
- Git must be installed and available in PATH
- If a configuration already exists, it will be loaded and updated
- For SSH remote URLs, ensure your SSH keys are configured
- Local-only mode auto-configures Git for users without Git setup
- Git URL validation prevents command injection attacks
