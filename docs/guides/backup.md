# Backups and Rollback

n0man provides a robust backup and rollback system to protect your dotfiles from accidental changes, sync errors, or corruption. Backups are created automatically before sync operations, and can also be created manually.

## Overview

### Backup Types

1. **Automatic Backups**: Created before every `sync` operation
2. **Manual Backups**: Created on-demand using `n0man backup create`

### Backup Lifecycle

```
Add dotfile → Sync → Automatic Backup → Sync Operations → Cleanup old backups
```

## Backup Storage

### Location

Backups are stored in the store directory:

```
~/.local/share/n0man/store/
  .backups/
    20240312_150430/          # Backup timestamp
      vimrc                   # Backed up files
      nvim/
        init.lua
      ...
    20240312_143052/
      ...
  .gitignore                 # Excludes .backups/
  vimrc                      # Current files
  nvim/
```

### Timestamp Format

Backups use the format: `YYYYMMDD_HHMMSS`

Example: `20240312_150430` = March 12, 2024 at 15:04:30

### What's Backed Up

Only files tracked in n0man configuration are backed up:

- The actual file/directory content (following symlinks)
- File permissions and metadata
- Directory structure

**NOT backed up**:
- Git repository (`.git/`)
- Configuration file (`n0man.toml`)
- Untracked files in store

## Automatic Backups

### When Created

Before every `sync` operation, a backup is created:

```bash
$ n0man sync
n0man sync
  Creating pre-sync backup snapshot...
✅ Backup created: 20240312_150430
  Checking for local changes...
  ...
```

### Behavior

- Created even if there are no local changes
- Always creates a new timestamped backup
- Backup is complete before any Git operations
- Old backups are cleaned up after successful sync

### Cleanup

After successful sync, old backups are automatically cleaned up:

```toml
[settings]
housekeeping_max_backups = 5
```

Only the N most recent backups are retained. Older backups are deleted.

## Manual Backups

### Create Manual Backup

Create a backup at any time:

```bash
$ n0man backup create
  Creating manual backup snapshot...
✅ Backup created: 20240312_151023
```

**Use cases**:
- Before making major changes
- Before adding risky files
- As a safety checkpoint

### No Files Tracked

If no dotfiles are tracked, backup creation is skipped:

```bash
$ n0man backup create
  Creating manual backup snapshot...
ℹ️  No dotfiles tracked, nothing to backup.
```

## Restore and Rollback

### Interactive Restore

Browse and restore from available backups:

```bash
$ n0man backup
n0man Backup Manager
? What would you like to do?
  [+] Create New Backup
  [↺] Restore from List
  [x] Exit

> [↺] Restore from List

Restorable Backups
? Choose a backup to restore:
  20240312_150430
  20240312_143052
  20240311_092015

> 20240312_150430

  Restoring from backup 20240312_150430
✅ Backup restore complete
```

### Rollback to Latest

Restore the most recent backup directly:

```bash
$ n0man backup rollback
  Restoring from latest backup 20240312_151023
✅ Rollback complete
```

**Use cases**:
- Quick undo after a bad sync
- Recover from sync errors
- Revert experimental changes

### Store Recovery

If the store directory is missing or corrupted, `n0man backup rollback` automatically detects this and offers recovery options:

```bash
$ n0man backup rollback
⚠️  Store directory is missing!
? Choose recovery method:
  > Restore from backup (3 snapshots available)
    Re-clone from remote (git@github.com:user/dotfiles.git)
    Reinitialize store (fresh start)
    Cancel

> Restore from backup

Restorable Backups
? Choose a backup to restore:
  20240312_151023
  20240312_143052
  20240311_092015

> 20240312_151023

  → Restoring from backup 20240312_151023
  → Recreating symlinks...
  ✓ Symlinks recreated
✅ Store restored from backup
```

**Recovery Options**:
- **Restore from backup**: Recreates store from a backup snapshot and recreates symlinks
- **Re-clone from remote**: Clones the remote repository fresh (if configured)
- **Reinitialize store**: Creates a fresh Git repository (loses history)
- **Cancel**: Aborts recovery

**Note**: After store recovery, all symlinks are automatically recreated to point to the restored store files.

## What Restore Does

When you restore a backup:

1. **Copies files from backup**: Files from backup directory are copied to store
2. **Overwrites existing files**: Current store files are replaced with backup versions
3. **Preserves symlinks**: Symlinks in your home directory are NOT modified
4. **Does not modify Git**: Git history is unchanged
5. **Does not modify configuration**: `n0man.toml` is unchanged

### What Changes

- **Store files**: `~/.local/share/n0man/store/<name>` are replaced
- **Symlinks**: `~/.config/app` still points to store (no change)
- **Configuration**: `~/.config/n0man/n0man.toml` unchanged
- **Git history**: Commits, branches, etc. unchanged

### What Doesn't Change

- Git repository state (`.git/`)
- Configuration file
- Symlink targets
- Untracked files in store

## Use Cases

### Before Risky Operations

```bash
# Create manual backup
$ n0man backup create
✅ Backup created: 20240312_151023

# Perform risky operation
$ n0man add ~/.config/production-app --no-security
```

### After Sync Issues

```bash
# Sync goes wrong
$ n0man sync
Error: sync failed

# Rollback immediately
$ n0man backup rollback
✅ Rollback complete
```

### Point-in-Time Recovery

```bash
# List backups
$ n0man backup

# Restore from specific time
$ n0man backup
# Choose: [↺] Restore from List
# Select: 20240311_092015
```

### Experimental Changes

```bash
# Create checkpoint
$ n0man backup create

# Make experimental changes
vim ~/.config/nvim/init.lua

# If it breaks, restore
$ n0man backup rollback
```

### Migration to New Machine

```bash
# On old machine
$ n0man backup create

# Copy backup directory to new machine
scp -r ~/.local/share/n0man/store/.backups user@new-machine:~

# On new machine, restore
$ n0man backup
```

## Backup Management

### List Available Backups

```bash
$ ls -la ~/.local/share/n0man/store/.backups/
total 12
drwxr-xr-x 3 user user 4096 Mar 12 15:04 20240312_150430
drwxr-xr-x 3 user user 4096 Mar 12 14:30 20240312_143052
drwxr-xr-x 3 user user 4096 Mar 11 09:20 20240311_092015
```

### Inspect Backup Contents

```bash
# View backup contents
ls -la ~/.local/share/n0man/store/.backups/20240312_150430/

# Compare with current
diff -r ~/.local/share/n0man/store/.backups/20240312_150430/ \
       ~/.local/share/n0man/store/ \
       --exclude=.git --exclude=.backups
```

### Delete Specific Backup

```bash
# Remove old backup manually
rm -rf ~/.local/share/n0man/store/.backups/20240311_092015
```

### Delete All Backups

```bash
# Remove all backups
rm -rf ~/.local/share/n0man/store/.backups/*
```

## Backup Size

### Estimate Backup Size

```bash
# Size of all backups
du -sh ~/.local/share/n0man/store/.backups/

# Size of specific backup
du -sh ~/.local/share/n0man/store/.backups/20240312_150430/
```

### Manage Backup Size

If backups are consuming too much space:

1. **Reduce retention**:
   ```toml
   [settings]
   housekeeping_max_backups = 3
   ```

2. **Clean up old backups**:
   ```bash
   rm -rf ~/.local/share/n0man/store/.backups/20240311*
   ```

3. **Exclude large files**:
   Use ignore patterns in dotfile configuration:
   ```toml
   [dotfiles.ignores]
   nvim = ["*.log", "cache/*", "node_modules/*"]
   ```

## Troubleshooting

### Backup Creation Fails

```bash
# Check disk space
df -h ~/.local/share/n0man

# Check permissions
ls -la ~/.local/share/n0man/store

# Check configuration
n0man list
```

### Restore Fails

```bash
# Verify backup exists
ls ~/.local/share/n0man/store/.backups/

# Check store path
n0man list

# Check file permissions
ls -la ~/.local/share/n0man/store
```

### Backup Missing After Sync

Backups are only kept if retention allows:

```toml
[settings]
housekeeping_max_backups = 5
```

If you have 10 backups and sync, the 5 oldest are deleted.

### No Backups Available

This happens when:
- Fresh install (no sync yet)
- All backups manually deleted
- `housekeeping_max_backups = 0`

## Best Practices

1. **Before major changes**: Always create manual backup
2. **Regular syncs**: Sync frequently to have more restore points
3. **Monitor space**: Check backup size periodically
4. **Test restores**: Periodically test that backups work
5. **Adjust retention**: Set `housekeeping_max_backups` based on needs

## Configuration

### Backup Retention

```toml
[settings]
housekeeping_max_backups = 5
```

- `0`: Keep all backups (unlimited space)
- `3-10`: Keep recent backups (recommended)
- `1`: Only keep latest backup (minimal space)

### Backup Timing

Backups are created:
- Before every `sync` (automatic)
- When running `n0man backup create` (manual)

## Notes

- Backups are not compressed (plain copies)
- Backups follow symlinks to actual content
- Backups are ignored by Git (in `.gitignore`)
- Backup directory is excluded from scans
- Backup creation is atomic (all or nothing)
- Old backups are cleaned up automatically

## Related Commands

- `n0man sync`: Creates automatic backups
- `n0man backup create`: Creates manual backup
- `n0man backup rollback`: Restores latest backup
- `n0man backup`: Interactive backup manager
- `n0man doctor`: Checks backup directory health
- Configuration in `n0man.toml`: Controls backup retention
