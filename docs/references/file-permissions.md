# File Permissions Reference

n0man uses secure file permissions by default to protect your dotfiles and configuration.

## Default Permissions

### Files

| File Type | Permission | Octal | Description |
|-----------|------------|-------|-------------|
| Configuration (`n0man.toml`) | `rw-------` | `0600` | User read/write only |
| Store files (dotfiles) | `rw-------` | `0600` | User read/write only |
| Backup files | `rw-------` | `0600` | User read/write only |
| Git config files | `rw-------` | `0600` | User read/write only |

### Directories

| Directory Type | Permission | Octal | Description |
|----------------|------------|-------|-------------|
| Store directory | `rwx------` | `0700` | User read/write/execute only |
| Backup directories | `rwx------` | `0700` | User read/write/execute only |
| Config directory | `rwx------` | `0700` | User read/write/execute only |
| Git directory | `rwx------` | `0700` | User read/write/execute only |

## Permission Rationale

### Why 0600 for Files?

- **Privacy**: Dotfiles may contain sensitive information
- **Security**: Prevents other users from reading your configuration
- **Best Practice**: Follows principle of least privilege

### Why 0700 for Directories?

- **Privacy**: Prevents other users from listing directory contents
- **Security**: Prevents unauthorized access to dotfiles
- **Isolation**: Each user's dotfiles are isolated

## Permission Verification

### Check Configuration Permissions

```bash
ls -la ~/.config/n0man/n0man.toml
# Expected: -rw------- 1 user user ... n0man.toml
```

### Check Store Permissions

```bash
ls -lad ~/.local/share/n0man/store
# Expected: drwx------ ... store

ls -la ~/.local/share/n0man/store/
# Expected: -rw------- for files, drwx------ for directories
```

### Check Backup Permissions

```bash
ls -lad ~/.local/share/n0man/store/.backups
# Expected: drwx------ ... .backups

ls -la ~/.local/share/n0man/store/.backups/
# Expected: drwx------ for each backup directory
```

## Permission Changes

### Manual Permission Changes

If permissions are incorrect, you can fix them:

```bash
# Fix configuration permissions
chmod 0600 ~/.config/n0man/n0man.toml

# Fix store directory permissions
chmod 0700 ~/.local/share/n0man/store

# Fix store file permissions
chmod 0600 ~/.local/share/n0man/store/*

# Fix backup permissions
chmod 0700 ~/.local/share/n0man/store/.backups
chmod 0700 ~/.local/share/n0man/store/.backups/*
```

### Automatic Permission Setting

n0man automatically sets correct permissions when:
- Creating configuration files
- Creating store directories
- Creating backup snapshots
- Copying files

## Security Implications

### Risks of Incorrect Permissions

| Permission | Risk | Impact |
|------------|------|--------|
| World-readable config | Other users can read your dotfile mappings | Medium |
| World-readable store | Other users can read your dotfiles | High |
| World-writable store | Other users can modify your dotfiles | Critical |
| World-readable backups | Other users can access backup history | Medium |

### Multi-User Systems

On multi-user systems, correct permissions are critical:
- Prevents other users from reading sensitive configuration
- Prevents accidental or malicious modification
- Maintains privacy of your dotfiles

## Environment Variables

### N0MAN_ALLOW_OUTSIDE_HOME

When set, allows managing files outside home directory:

```bash
export N0MAN_ALLOW_OUTSIDE_HOME=true
```

**Warning:** This bypasses security restrictions. Use only in trusted environments.

**Permission Note:** Even with this set, file permissions remain secure (0600/0700).

## Troubleshooting

### "Permission denied" Errors

**Cause:** Incorrect file or directory permissions.

**Solution:**
```bash
# Check current permissions
ls -la ~/.local/share/n0man/store

# Fix permissions
chmod 0700 ~/.local/share/n0man/store
chmod 0600 ~/.local/share/n0man/store/*
```

### "Cannot read configuration" Errors

**Cause:** Configuration file permissions too restrictive.

**Solution:**
```bash
# Ensure user can read
chmod u+r ~/.config/n0man/n0man.toml

# Should be 0600
chmod 0600 ~/.config/n0man/n0man.toml
```

### Backup Creation Fails

**Cause:** Backup directory permissions incorrect.

**Solution:**
```bash
chmod 0700 ~/.local/share/n0man/store/.backups
```

## Best Practices

1. **Never use 0777** - Never make n0man directories world-writable
2. **Stick to defaults** - n0man's default permissions are secure
3. **Regular audits** - Periodically check permissions:
   ```bash
   find ~/.local/share/n0man -perm /o+rwx  # Find world-accessible files
   ```
4. **Use n0man commands** - Let n0man manage permissions automatically
5. **Backup before changes** - Before manually changing permissions

## Comparison with Other Tools

| Tool | Config Permissions | Store Permissions |
|------|-------------------|-------------------|
| n0man | 0600 | 0700 |
| GNU Stow | Default umask | Default umask |
| Chezmoi | 0600 | 0700 |
| Home Manager | Default umask | Default umask |

n0man follows security best practices with restrictive default permissions.
