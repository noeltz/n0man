# Troubleshooting Guide

This guide helps you diagnose and fix common n0man issues.

## Quick Diagnosis

Run these commands to diagnose issues:

```bash
# Check n0man is installed
n0man version

# Check configuration exists
ls -la ~/.config/n0man/n0man.toml

# Check store directory
ls -la ~/.local/share/n0man/store

# Run health checks
n0man doctor

# Check status
n0man status
```

## Installation Issues

### "command not found: n0man"

**Problem:** n0man is not in your PATH.

**Solutions:**

1. **Verify installation:**
   ```bash
   # Check if installed
   go list -f '{{.Target}}' github.com/noeltz/n0man/cmd/n0man
   
   # Find binary
   find ~ -name "n0man" -type f 2>/dev/null
   ```

2. **Add to PATH:**
   ```bash
   # Add Go bin to PATH
   echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Reinstall:**
   ```bash
   go install github.com/noeltz/n0man/cmd/n0man@latest
   ```

### Build Fails

**Problem:** `go build` fails with errors.

**Solutions:**

1. **Check Go version:**
   ```bash
   go version
   # Should be 1.26.1 or later
   ```

2. **Update Go:**
   ```bash
   # Download latest from https://go.dev/dl/
   ```

3. **Clean and rebuild:**
   ```bash
   go clean
   go build -o n0man ./cmd/n0man
   ```

## Configuration Issues

### "n0man.toml not found"

**Problem:** Configuration file is missing.

**Solutions:**

1. **Check location:**
   ```bash
   ls -la ~/.config/n0man/n0man.toml
   ```

2. **Reinitialize:**
   ```bash
   n0man init
   ```

3. **Restore from backup:**
   ```bash
   ls -la ~/.config/n0man/n0man.toml.backup-*
   cp ~/.config/n0man/n0man.toml.backup-<latest> ~/.config/n0man/n0man.toml
   ```

### "Invalid configuration format"

**Problem:** TOML syntax error in configuration.

**Solutions:**

1. **Check syntax:**
   ```bash
   cat ~/.config/n0man/n0man.toml
   ```

2. **Restore from backup:**
   ```bash
   cp ~/.config/n0man/n0man.toml.backup-* ~/.config/n0man/n0man.toml
   ```

3. **Recreate configuration:**
   ```bash
   mv ~/.config/n0man/n0man.toml ~/.config/n0man/n0man.toml.broken
   n0man init
   ```

## Git Issues

### "Author identity unknown"

**Problem:** Git user.name or user.email not configured.

**Solutions:**

1. **For local-only mode:** n0man auto-configures Git. Reinitialize:
   ```bash
   rm -rf ~/.local/share/n0man/store
   n0man init
   ```

2. **For remote mode:** Configure Git globally:
   ```bash
   git config --global user.name "Your Name"
   git config --global user.email "you@example.com"
   ```

3. **Check configuration:**
   ```bash
   git config user.name
   git config user.email
   ```

### "Git URL contains invalid character"

**Problem:** Git URL contains shell metacharacters.

**Solutions:**

1. **Use valid URL formats:**
   ```bash
   # Valid
   n0man init git@github.com:user/dotfiles.git
   n0man init https://github.com/user/dotfiles.git
   n0man init ssh://git@github.com/user/dotfiles.git
   
   # Invalid (contains ;)
   n0man init 'git@github.com:user;rm -rf /'
   ```

2. **Remove special characters:**
   ```bash
   # Check URL
   echo $YOUR_GIT_URL
   
   # Remove special chars
   n0man init "git@github.com:user/dotfiles.git"
   ```

### "Failed to pull" / "Failed to push"

**Problem:** Git remote operations failed.

**Solutions:**

1. **Check SSH keys:**
   ```bash
   # Test SSH connection
   ssh -T git@github.com
   ```

2. **Check remote URL:**
   ```bash
   cd ~/.local/share/n0man/store
   git remote -v
   ```

3. **Fix remote:**
   ```bash
   git remote set-url origin git@github.com:user/dotfiles.git
   ```

4. **Test connection:**
   ```bash
   git fetch origin
   ```

## File System Issues

### "Path must be within home directory"

**Problem:** Trying to manage files outside home directory.

**Solutions:**

1. **Move file to home directory (recommended):**
   ```bash
   cp /etc/myapp/config ~/.config/myapp-config
   n0man add ~/.config/myapp-config
   ```

2. **Create symlink in home:**
   ```bash
   ln -s /etc/myapp/config ~/.config/myapp-config
   n0man add ~/.config/myapp-config
   ```

3. **Allow outside home (not recommended):**
   ```bash
   export N0MAN_ALLOW_OUTSIDE_HOME=true
   n0man add /etc/myapp/config
   ```

### "Destination already exists"

**Problem:** File already exists in store.

**Solutions:**

1. **Check if already tracked:**
   ```bash
   n0man list | grep <filename>
   ```

2. **Remove and re-add:**
   ```bash
   n0man rm <name>
   n0man add <path>
   ```

3. **Use different name:**
   ```bash
   n0man add <path> custom-name
   ```

### "Permission denied"

**Problem:** File or directory permissions incorrect.

**Solutions:**

1. **Check permissions:**
   ```bash
   ls -la ~/.local/share/n0man/store
   ls -la ~/.config/n0man/n0man.toml
   ```

2. **Fix permissions:**
   ```bash
   chmod 0700 ~/.local/share/n0man/store
   chmod 0600 ~/.local/share/n0man/store/*
   chmod 0600 ~/.config/n0man/n0man.toml
   ```

## Symlink Issues

### Broken Symlinks

**Problem:** Symlinks point to non-existent files.

**Solutions:**

1. **Detect broken symlinks:**
   ```bash
   n0man doctor
   ```

2. **Fix automatically:**
   ```bash
   n0man doctor --fix
   ```

3. **Manual fix:**
   ```bash
   # Find broken symlinks
   find ~ -type l ! -exec test -e {} \; -print
   
   # Remove broken symlink
   rm ~/.bashrc
   
   # Recreate
   ln -s ~/.local/share/n0man/store/.bashrc ~/.bashrc
   ```

### "Symlink already exists"

**Problem:** File exists where symlink should be.

**Solutions:**

1. **Backup and remove:**
   ```bash
   mv ~/.bashrc ~/.bashrc.backup
   n0man add ~/.bashrc
   ```

2. **Force re-add:**
   ```bash
   n0man rm <name>
   n0man add <path>
   ```

## Security Issues

### "Security scan failed"

**Problem:** Secrets detected in dotfiles.

**Solutions:**

1. **Review findings:**
   ```bash
   n0man security scan
   ```

2. **Remove or fix secrets:**
   ```bash
   # Edit file to remove secret
   vim ~/.config/app/config
   
   # Or move secret to separate file
   mv ~/.config/app/config ~/.config/app/config.secrets
   echo "include config.secrets" >> ~/.config/app/config
   ```

3. **Add to allowlist (if false positive):**
   ```toml
   # ~/.config/n0man/n0man.toml
   [security.allowlist]
   files = ["safe_config.yaml"]
   ```

4. **Skip scan (only if sure it's safe):**
   ```bash
   n0man add <path> --no-security
   ```

### False Positives

**Problem:** Safe files flagged as secrets.

**Solutions:**

1. **Add to allowlist:**
   ```toml
   # ~/.config/n0man/n0man.toml
   [security.allowlist]
   patterns = ["*test*", "*example*"]
   files = ["demo_config.yaml"]
   ```

2. **Lower sensitivity:**
   ```toml
   [security]
   sensitivity = "low"
   ```

3. **Disable content scanning:**
   ```toml
   [security]
   scan_content = false
   ```

## Backup Issues

### "No backups available"

**Problem:** No backups exist to restore.

**Solutions:**

1. **Check backup directory:**
   ```bash
   ls -la ~/.local/share/n0man/store/.backups/
   ```

2. **If store is missing, reinitialize:**
   ```bash
   n0man init
   ```

3. **If remote configured, re-clone:**
   ```bash
   rm -rf ~/.local/share/n0man/store
   n0man init git@github.com:user/dotfiles.git
   ```

### "Backup creation failed"

**Problem:** Cannot create backup.

**Solutions:**

1. **Check disk space:**
   ```bash
   df -h ~/.local/share/n0man
   ```

2. **Check permissions:**
   ```bash
   ls -la ~/.local/share/n0man/store/.backups
   chmod 0700 ~/.local/share/n0man/store/.backups
   ```

3. **Clean old backups:**
   ```bash
   # Manually remove old backups
   ls -lt ~/.local/share/n0man/store/.backups/
   rm -rf ~/.local/share/n0man/store/.backups/<old-backup>
   ```

## Sync Issues

### Pre-Flight Checks Fail

**Problem:** Pre-flight checks detect issues.

**Solutions:**

1. **Run doctor:**
   ```bash
   n0man doctor --fix
   ```

2. **Check specific issues:**
   ```bash
   n0man status
   n0man list
   ```

3. **Manual fix:**
   ```bash
   # Fix broken symlinks
   n0man doctor --fix
   
   # Reinitialize if store missing
   n0man init
   ```

### Conflict During Sync

**Problem:** Merge conflicts during pull.

**Solutions:**

1. **Interactive resolution:**
   ```bash
   n0man sync
   # Follow prompts to resolve
   ```

2. **Non-interactive (CI/CD):**
   ```bash
   n0man sync --conflict-strategy=keep-local
   n0man sync --conflict-strategy=keep-remote
   n0man sync --conflict-strategy=abort
   ```

3. **Manual resolution:**
   ```bash
   cd ~/.local/share/n0man/store
   git status
   # Edit conflicted files
   git add .
   git rebase --continue
   ```

## Performance Issues

### Slow Operations

**Problem:** n0man commands are slow.

**Solutions:**

1. **Check store size:**
   ```bash
   du -sh ~/.local/share/n0man/store
   ```

2. **Clean old backups:**
   ```bash
   ls -lt ~/.local/share/n0man/store/.backups/
   # Remove old backups
   ```

3. **Reduce backup retention:**
   ```toml
   # ~/.config/n0man/n0man.toml
   [settings]
   housekeeping_max_backups = 3
   ```

### Large File Issues

**Problem:** Operations fail with large files.

**Solutions:**

1. **Check file size:**
   ```bash
   ls -lh ~/.config/<large-file>
   ```

2. **Exclude from tracking:**
   ```bash
   # Don't track large files
   # Use separate backup solution
   ```

3. **Increase file size limit:**
   ```toml
   # ~/.config/n0man/n0man.toml
   [security.content_scan]
   max_file_size = 104857600  # 100MB
   ```

## Getting More Help

### Enable Debug Output

```bash
# Future feature - check for updates
export N0MAN_DEBUG=true
n0man sync
```

### Collect Diagnostic Information

```bash
# Create diagnostic report
{
    echo "=== n0man version ==="
    n0man version
    echo ""
    echo "=== n0man doctor ==="
    n0man doctor
    echo ""
    echo "=== n0man status ==="
    n0man status
    echo ""
    echo "=== Configuration ==="
    cat ~/.config/n0man/n0man.toml
    echo ""
    echo "=== Store contents ==="
    ls -la ~/.local/share/n0man/store/
} > n0man-diagnostic.txt
```

### Report a Bug

1. **Check existing issues:** GitHub Issues
2. **Collect diagnostics:** See above
3. **Include:**
   - n0man version
   - OS and version
   - Steps to reproduce
   - Expected vs actual behavior
   - Diagnostic output

## See Also

- [Security Guide](../guides/security.md) - Security features
- [Backup Guide](../guides/backup.md) - Backup and recovery
- [Configuration Reference](../references/configuration.md) - Config options
