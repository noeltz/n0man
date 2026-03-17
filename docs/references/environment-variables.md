# Environment Variables Reference

This document lists all environment variables recognized by n0man.

## Security Variables

### N0MAN_ALLOW_OUTSIDE_HOME

**Purpose:** Allow managing files outside the home directory.

**Values:**
- `true` - Allow paths outside home directory
- (unset) - Restrict to home directory only (default)

**Default:** Unset (restricted)

**Example:**
```bash
export N0MAN_ALLOW_OUTSIDE_HOME=true
n0man add /etc/myapp/config  # Now allowed
```

**Warning:** This bypasses important security restrictions. Only use in trusted environments.

**Use Cases:**
- Testing in non-standard environments
- Managing system-wide configuration (advanced users)
- Containerized environments with different home directory structure

**Security Note:** When set, a warning is printed to stderr:
```
WARNING: N0MAN_ALLOW_OUTSIDE_HOME is set - path restrictions disabled
```

## Git Variables

n0man respects standard Git environment variables:

### GIT_AUTHOR_NAME

**Purpose:** Override Git author name for commits.

**Default:** Git's configured user.name or "n0man" for local-only mode.

### GIT_AUTHOR_EMAIL

**Purpose:** Override Git author email for commits.

**Default:** Git's configured user.email or "n0man@localhost" for local-only mode.

### GIT_SSH_COMMAND

**Purpose:** Specify SSH command for Git operations.

**Example:**
```bash
export GIT_SSH_COMMAND="ssh -i ~/.ssh/id_ed25519"
n0man sync
```

## Path Variables

### XDG_CONFIG_HOME

**Purpose:** Override default configuration directory.

**Default:** `~/.config`

**Example:**
```bash
export XDG_CONFIG_HOME=~/.my-config
n0man init  # Config created at ~/.my-config/n0man/n0man.toml
```

### HOME

**Purpose:** Override home directory (affects all path resolution).

**Default:** User's actual home directory

**Example:**
```bash
HOME=/tmp/test-home n0man init
```

**Use Cases:**
- Testing in isolated environments
- Running n0man in containers
- Testing without affecting actual home directory

## System Variables

### TMPDIR

**Purpose:** Override temporary directory for temporary files.

**Default:** `/tmp` or system default

**Example:**
```bash
export TMPDIR=/var/tmp
n0man sync
```

## Debug Variables

### N0MAN_DEBUG

**Purpose:** Enable debug logging (future feature).

**Values:**
- `true` - Enable debug output
- (unset) - Normal output (default)

**Note:** This feature is planned for future versions.

## Shell Variables

### SHELL

**Purpose:** Determines default shell for bootstrap scripts.

**Default:** User's login shell

**Example:**
```bash
SHELL=/bin/zsh n0man init
```

## Usage in Scripts

### CI/CD Environment

```yaml
# GitHub Actions
- name: Setup n0man
  env:
    N0MAN_ALLOW_OUTSIDE_HOME: "false"  # Explicit security
    GIT_AUTHOR_NAME: "GitHub Actions"
    GIT_AUTHOR_EMAIL: "actions@github.com"
  run: |
    n0man init
    n0man sync --conflict-strategy=keep-remote
```

### Testing Environment

```bash
#!/bin/bash
# test-setup.sh

export HOME=/tmp/n0man-test-home
export XDG_CONFIG_HOME=$HOME/.config
export N0MAN_ALLOW_OUTSIDE_HOME=false

# Clean test environment
rm -rf $HOME
mkdir -p $HOME

# Run tests
n0man init
n0man add ~/.testrc
n0man sync
```

### Container Environment

```dockerfile
ENV HOME=/home/n0man
ENV XDG_CONFIG_HOME=/home/n0man/.config
ENV N0MAN_ALLOW_OUTSIDE_HOME=false

RUN n0man init
```

## Variable Precedence

When multiple sources define the same variable:

1. **Command-line** (highest priority)
   ```bash
   HOME=/tmp n0man init
   ```

2. **Exported environment**
   ```bash
   export HOME=/tmp
   n0man init
   ```

3. **Shell variables**
   ```bash
   HOME=/tmp
   n0man init
   ```

4. **Default values** (lowest priority)

## Troubleshooting

### "Path must be within home directory"

**Problem:** Trying to add a file outside home directory.

**Solution:** Either:
1. Move file to home directory (recommended)
2. Use `N0MAN_ALLOW_OUTSIDE_HOME=true` (not recommended for production)

```bash
# Option 1: Move to home
cp /etc/myapp/config ~/.config/myapp-config
n0man add ~/.config/myapp-config

# Option 2: Allow outside home (use with caution)
export N0MAN_ALLOW_OUTSIDE_HOME=true
n0man add /etc/myapp/config
```

### "Configuration not found"

**Problem:** n0man can't find configuration file.

**Solution:** Check `XDG_CONFIG_HOME`:
```bash
echo $XDG_CONFIG_HOME
# Should be ~/.config or where you expect

# Find configuration
find ~ -name "n0man.toml" 2>/dev/null
```

### Git Author Unknown

**Problem:** Git commits fail due to missing author information.

**Solution:** Set Git environment variables:
```bash
export GIT_AUTHOR_NAME="Your Name"
export GIT_AUTHOR_EMAIL="you@example.com"
```

Or configure Git globally:
```bash
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
```

## Best Practices

1. **Don't set N0MAN_ALLOW_OUTSIDE_HOME in production** - Security risk
2. **Use XDG_CONFIG_HOME for custom config locations** - Standard approach
3. **Set Git variables in CI/CD** - Ensures consistent commits
4. **Document environment in scripts** - Makes scripts portable
5. **Test with different HOME values** - Ensures portability

## See Also

- [Configuration Reference](configuration.md) - n0man.toml reference
- [File Permissions Reference](file-permissions.md) - Security and permissions
- [Security Guide](../guides/security.md) - Security features overview
