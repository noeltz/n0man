# `n0man sync`

Bidirectional sync with pre-flight checks, commit local changes, pull remote changes, and push.

## Usage

```bash
n0man sync [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--no-security` | Skip security scanning before commit |
| `--conflict-strategy` | Non-interactive conflict resolution (`keep-local`, `keep-remote`, `abort`) |

## Description

The `sync` command performs a complete bidirectional synchronization between your local machine and the remote Git repository. It handles the full lifecycle:

1. **Pre-Flight Checks**: Validates environment before any operations
2. **Pre-Sync Backup**: Creates a snapshot backup before any destructive operations
3. **Local Commit**: Commits any changes in the store to the local Git repository
4. **Pull**: Pulls remote changes using rebase
5. **Push**: Pushes local changes to the remote repository

This ensures that both local and remote changes are preserved and merged intelligently.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully synced |
| 1 | Configuration not found or store path not configured |
| 1 | Backup creation failed |
| 1 | Security scan found issues |
| 1 | Git operations failed (commit, pull, push) |
| 1 | Conflict resolution aborted |
| 1 | Pre-flight checks failed |

## Examples

### Successful Sync with Pre-Flight Checks

```bash
$ n0man sync
n0man sync
  → Running pre-flight checks...
  ✓ Configuration loaded
  ✓ Store directory exists
  ✓ Git repository exists
  ✓ All symlinks valid
  → Creating pre-sync backup snapshot...
  ✓ Backup created: 20240312_143052
  → Checking for local changes...
  → Running security scan...
  ✓ Security scan passed
  → Committing local changes...
  ✓ Local changes committed
  → Synchronizing with remote...
  ✓ Remote synchronization complete

  ✓ Sync completed successfully!
```

### Sync with Pre-Flight Issues (Auto-Fix)

```bash
$ n0man sync
n0man sync
  → Running pre-flight checks...
  ✓ Configuration loaded
  ✓ Store directory exists
  ✓ Git repository exists
  ⚠ 1 broken symlink(s)

  ? Fix 1 issue(s) before continuing? (Y/n)
  > Yes

  → Fixing issues...
  ✓ ✓ Fixed: Symlinks
  ✓ Fixed 1 issue(s)!

  → Creating pre-sync backup snapshot...
  ...
```

### Non-Interactive Conflict Resolution (CI/CD)

```bash
$ n0man sync --conflict-strategy=keep-remote
n0man sync
  → Running pre-flight checks...
  ✓ Configuration loaded
  ✓ Store directory exists
  ✓ Git repository exists
  ✓ All symlinks valid
  → Creating pre-sync backup snapshot...
  ✓ Backup created: 20240312_143052
  → Checking for local changes...
  ✓ Local changes committed
  → Synchronizing with remote...
  → Conflict strategy: Keeping remote changes...
  ✓ Remote synchronization complete

  ✓ Sync completed successfully!
```

### Security Scan Failure

```bash
$ n0man sync
n0man sync
  → Running pre-flight checks...
  ✓ Configuration loaded
  ✓ Store directory exists
  ✓ Git repository exists
  ✓ All symlinks valid
  → Creating pre-sync backup snapshot...
  ✓ Backup created: 20240312_143052
  → Checking for local changes...
  → Running security scan...
🚨 Found 3 potential security issue(s):

  1. [HIGH] .local/share/n0man/store/.env
     → api_key pattern match
  2. [HIGH] .local/share/n0man/store/config.yaml
     → password pattern match
  3. [CRITICAL] .local/share/n0man/store/private.pem
     → private_key pattern match

Error: security scan found 3 issue(s). Fix them or re-run with --no-security
```

### Conflict Resolution (Interactive)

```bash
$ n0man sync
n0man sync
  → Running pre-flight checks...
  ✓ Configuration loaded
  ✓ Store directory exists
  ✓ Git repository exists
  ✓ All symlinks valid
  → Creating pre-sync backup snapshot...
  ✓ Backup created: 20240312_143052
  → Checking for local changes...
  ✓ Local changes committed
  → Synchronizing with remote...
CONFLICT (content): Merge conflict in vimrc

? How would you like to resolve the conflict?
  Keep local changes
  Keep remote changes
  Abort and resolve manually

> Keep local changes

  → Resolving: Keeping local changes...
  ✓ Remote synchronization complete

  ✓ Sync completed successfully!
```

## What It Does

### 1. Pre-Sync Backup

Creates a timestamped backup snapshot in `.backups/`:

```
~/.local/share/n0man/store/.backups/20240312_143052/
  vimrc
  nvim/
  ...
```

### 2. Local Changes Check

Runs `git status --porcelain` to detect uncommitted changes.

### 3. Security Scan

If changes exist and security is enabled:
- Scans all files in the store for secrets
- If findings detected:
  - If `fail_on_secrets=true`: Fails with error
  - If `interactive=true`: Prompts to continue
- If no findings: Proceeds

### 4. Commit Local Changes

Updates `.gitignore` with ignore patterns, then:

```bash
git add .
git commit -m "chore: auto-sync dotfiles update"
```

### 5. Pull Remote Changes

```bash
git pull --rebase origin HEAD
```

If conflicts occur during rebase, prompts for resolution:
- **Keep local**: Use `git checkout --theirs .` (during rebase)
- **Keep remote**: Use `git checkout --ours .` (during rebase)
- **Abort manually**: Run `git rebase --abort`

### 6. Push to Remote

```bash
git push origin HEAD
```

## Backup Management

Every `sync` operation creates a new backup snapshot. After a successful sync, old backups are automatically cleaned up to retain only the N most recent backups, where N is `housekeeping_max_backups` in configuration (default: 5).

**Backup Lifecycle**:
1. Before every `sync`: New backup created with timestamp
2. After successful `sync`: Old backups deleted if total exceeds `housekeeping_max_backups`
3. Oldest backups are removed first to maintain the limit

Manual backup management:

```bash
n0man backup create    # Create manual backup
n0man backup            # Interactive restore
n0man backup rollback   # Restore latest
```

## Security Scanning

The sync command includes an automatic security scan. To skip:

```bash
n0man sync --no-security
```

**Warning**: Only skip scanning if you're confident no secrets have been added.

## Conflict Resolution Strategy

n0man uses Git rebase for pulling changes. During a rebase:

- `HEAD` (ours) = The changes we're rebasing **onto** (Remote/Upstream)
- `MERGE_HEAD` (theirs) = The changes we're rebasing (Local)

To keep local changes: Use `--theirs` during rebase
To keep remote changes: Use `--ours` during rebase

This may seem counterintuitive but is correct for rebase operations.

## Error Recovery

### Backup Restore

If sync fails or corrupts your files, restore from backup:

```bash
n0man backup rollback
```

Or interactively:

```bash
n0man backup
[+] Create New Backup
[↺] Restore from List
[x] Exit

? What would you like to do?
> [↺] Restore from List
```

### Manual Git Recovery

If Git operations fail, you can manually intervene:

```bash
cd ~/.local/share/n0man/store
git status
git pull --rebase origin HEAD
# Resolve conflicts manually
git add .
git rebase --continue
git push origin HEAD
```

## Local-Only Mode

If no `remote_url` is configured, `sync` still works:

- Creates backups
- Commits local changes
- Skips pull/push operations

This is useful for local version control without a remote repository.

## Notes

- Always creates a backup before sync (even if no changes)
- Security scan is enabled by default
- Uses rebase for cleaner history
- Conflict resolution is interactive
- Backup cleanup is automatic
- Git must be installed and configured
- SSH keys must be set up for SSH remotes

## Best Practices

1. **Run `status` first**: Check for issues before syncing
2. **Review changes**: Use `git diff` in the store to review changes
3. **Test locally**: Make changes, run `sync` without remote to verify
4. **Use branches**: For experimental changes, create a Git branch
5. **Regular syncs**: Sync frequently to avoid large conflicts
