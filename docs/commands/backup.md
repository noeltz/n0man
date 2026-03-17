# `n0man backup`

Manage dotfile backups interactively.

## Usage

```bash
n0man backup [command]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `backup` (no args) | Interactive backup manager |
| `backup create` | Directly create a backup snapshot (non-interactive) |
| `backup rollback` | Directly restore the latest backup (non-interactive) |

## Description

The `backup` command provides manual control over n0man's backup system. While backups are created automatically before `sync` operations, you can also:

- Create manual backups at any time
- List available backups
- Restore from specific backups
- Rollback to the latest backup

Backups are timestamped snapshots of all tracked dotfiles stored in `.backups/` directory.

## Exit Codes

| Command | Exit Codes |
|---------|------------|
| `backup` | 0: Success, 1: Error, user cancel |
| `backup create` | 0: Success, 1: Error |
| `backup rollback` | 0: Success, 1: No backups available |

## Examples

### Interactive Backup Manager

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

### Create Manual Backup

```bash
$ n0man backup create
  Creating manual backup snapshot...
✅ Backup created: 20240312_151023
```

### Rollback to Latest Backup

```bash
$ n0man backup rollback
  Restoring from latest backup 20240312_151023
✅ Rollback complete
```

### Create Backup When No Files Tracked

```bash
$ n0man backup create
  Creating manual backup snapshot...
ℹ️  No dotfiles tracked, nothing to backup.
```

## What It Does

### Create Backup

1. Checks if any dotfiles are tracked
2. Creates timestamped directory: `.backups/20240312_151023/`
3. For each tracked dotfile:
   - Resolves the actual file (follows symlinks if needed)
   - Copies the file/directory to backup directory
   - Preserves file permissions and directory structure

### List Backups

Shows all available backup directories in `.backups/`, sorted from newest to oldest.

### Restore Backup

1. Prompts for backup selection (if using interactive mode)
2. Restores all files from backup directory to the store
3. Overwrites existing files in the store
4. Updates symlinks remain pointing to store (no changes)

## Backup Structure

Backups are organized in the store:

```
~/.local/share/n0man/store/
  .backups/
    20240312_151023/          # Backup timestamp
      vimrc                   # Backed up file
      nvim/                   # Backed up directory
        init.lua
        ...
    20240312_143052/
      ...
  vimrc                       # Current store files
  nvim/
```

Timestamp format: `YYYYMMDD_HHMMSS`

## Interactive Flow

When running `n0man backup` without arguments:

1. **Choose Action**:
   - `[+] Create New Backup`: Create a new snapshot
   - `[↺] Restore from List`: Choose a backup to restore
   - `[x] Exit`: Cancel

2. **If Create New Backup**:
   - Creates backup immediately
   - Shows timestamp

3. **If Restore from List**:
   - Shows all available backups (newest first)
   - Select backup to restore
   - Restores files from selected backup

## Automatic Backup Cleanup

After each `sync` operation, old backups are automatically cleaned up. Only the N most recent backups are kept, where N is `housekeeping_max_backups` in configuration (default: 5).

## Use Cases

### Before Risky Operations

```bash
$ n0man backup create
# Create manual backup before major changes
$ n0man add ~/.config/important-app --no-security
# Add potentially risky file
```

### After Sync Issues

```bash
$ n0man sync
# Something went wrong, files corrupted
$ n0man backup rollback
# Restore to pre-sync state
```

### Point-in-Time Recovery

```bash
$ n0man backup
# List all backups
$ n0man backup
# Choose specific timestamp to restore
```

## Comparison with `sync` Backups

| Aspect | Automatic (sync) | Manual (backup) |
|--------|-----------------|-----------------|
| When created | Before every `sync` | On demand |
| User control | Automatic | User-initiated |
| Naming | Timestamp | Timestamp |
| Cleanup | Automatic | Automatic |
| Restore | Via `backup` command | Via `backup` command |

## Notes

- Backups only track files in n0man configuration
- Backup follows symlinks to copy actual content
- Backup preserves permissions and directory structure
- Backup size depends on tracked files
- No limit on backup count (only cleanup limits it)
- Backups are not compressed (plain copies)
- Backup directory is ignored by Git (in `.gitignore`)

## Best Practices

1. **Before major changes**: Create manual backup
2. **After sync issues**: Rollback to investigate
3. **Regular cleanups**: Let automatic cleanup manage space
4. **Test restores**: Periodically test that backups work
5. **Monitor space**: Large dotfiles can consume backup space

## Troubleshooting

### Backup Creation Fails

```bash
# Check disk space
df -h ~/.local/share/n0man

# Check permissions
ls -la ~/.local/share/n0man
```

### Restore Fails

```bash
# Verify backup exists
ls ~/.local/share/n0man/store/.backups

# Check store path
n0man list
```

### No Backups Available

This can happen if:
- This is a fresh install (no sync yet)
- Backups were manually deleted
- `housekeeping_max_backups` is set to 0

## Related Commands

- `n0man sync`: Creates automatic backups
- `n0man add`: Can use `--no-security` to bypass scanning
- `n0man doctor`: Checks backup directory health
