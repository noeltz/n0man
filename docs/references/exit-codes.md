# Exit Codes Reference

This document lists all exit codes used by n0man commands.

## Standard Exit Codes

| Code | Meaning | Used By |
|------|---------|---------|
| `0` | Success | All commands |
| `1` | General error | All commands |
| `130` | Interrupted by user (SIGINT/Ctrl+C) | All commands |

## Command-Specific Exit Codes

### `n0man init`

| Code | Meaning |
|------|---------|
| `0` | Successfully initialized |
| `1` | Failed to clone or initialize |
| `1` | Failed to save configuration |
| `1` | Invalid Git URL (contains shell metacharacters) |

### `n0man add`

| Code | Meaning |
|------|---------|
| `0` | Successfully added |
| `1` | Path does not exist |
| `1` | Invalid path (outside home directory) |
| `1` | Security scan failed |
| `1` | Aborted due to security findings |

### `n0man rm`

| Code | Meaning |
|------|---------|
| `0` | Successfully removed |
| `1` | Dotfile not tracked |

### `n0man sync`

| Code | Meaning |
|------|---------|
| `0` | Successfully synced |
| `1` | Configuration not found |
| `1` | Store path not configured |
| `1` | Backup creation failed |
| `1` | Security scan failed |
| `1` | Git operations failed |
| `1` | Conflict resolution aborted |
| `1` | Pre-flight checks failed |

### `n0man status`

| Code | Meaning |
|------|---------|
| `0` | Status displayed (may include warnings) |
| `1` | Configuration not found |

### `n0man list`

| Code | Meaning |
|------|---------|
| `0` | List displayed |
| `1` | Configuration not found |

### `n0man backup`

| Command | Exit Codes |
|---------|------------|
| `backup` (interactive) | `0`: Success, `1`: Error or user cancel |
| `backup create` | `0`: Success, `1`: Error |
| `backup rollback` | `0`: Success, `1`: No backups available or error |

### `n0man doctor`

| Code | Meaning |
|------|---------|
| `0` | All checks passed or issues fixed |
| `0` | Issues found (displayed to user) |
| `1` | Failed to load configuration |

### `n0man security scan`

| Code | Meaning |
|------|---------|
| `0` | No security issues found |
| `1` | Security scan failed (error) |
| `1` | Security issues found (secrets detected) |

### `n0man self-update`

| Code | Meaning |
|------|---------|
| `0` | Successfully updated |
| `1` | Update failed |
| `1` | Go not installed |
| `1` | Timeout (5 minutes) |

### `n0man version`

| Code | Meaning |
|------|---------|
| `0` | Version displayed |

### `n0man completion`

| Code | Meaning |
|------|---------|
| `0` | Completion script generated |

## Signal Handling

| Signal | Exit Code | Behavior |
|--------|-----------|----------|
| SIGINT (Ctrl+C) | `130` | Graceful shutdown with cleanup |
| SIGTERM | `130` | Graceful shutdown with cleanup |

## Usage in Scripts

### Check for Success

```bash
if n0man sync; then
    echo "Sync successful"
else
    echo "Sync failed with exit code $?"
fi
```

### Handle Specific Errors

```bash
n0man security scan
case $? in
    0)
        echo "No security issues"
        ;;
    1)
        echo "Security issues found - review and fix"
        exit 1
        ;;
esac
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Sync dotfiles
  run: n0man sync --conflict-strategy=keep-remote
  continue-on-error: false
  
- name: Security scan
  run: n0man security scan || exit 1
```

## Error Code Categories

### Configuration Errors (1xx)

| Code | Meaning |
|------|---------|
| `1` | Configuration not found |
| `1` | Invalid configuration format |
| `1` | Store path not configured |

### Security Errors (2xx)

| Code | Meaning |
|------|---------|
| `1` | Security scan failed |
| `1` | Secrets detected |
| `1` | Invalid path (security restriction) |
| `1` | Invalid Git URL (security restriction) |

### Git Errors (3xx)

| Code | Meaning |
|------|---------|
| `1` | Git operation failed |
| `1` | Conflict resolution failed |
| `1` | Remote repository error |

### File System Errors (4xx)

| Code | Meaning |
|------|---------|
| `1` | File not found |
| `1` | Permission denied |
| `1` | Path validation failed |

## Troubleshooting Exit Codes

### Exit Code 130 (Interrupted)

**Cause:** User pressed Ctrl+C or sent SIGINT.

**Solution:** Command was interrupted. Check for partial operations:
```bash
n0man doctor  # Check for issues
n0man status  # Check sync status
```

### Exit Code 1 (General Error)

**Cause:** Various errors.

**Solution:** Check error message for details:
```bash
n0man sync 2>&1 | tail -20  # View error output
```

### Exit Code 1 (Security Scan)

**Cause:** Secrets detected in dotfiles.

**Solution:** Review and fix security issues:
```bash
n0man security scan  # See what was detected
# Fix or remove secrets
n0man sync --no-security  # Only if you're sure it's safe
```
