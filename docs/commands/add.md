# `n0man add`

Add a file or directory to n0man and replace it with a symlink.

## Usage

```bash
n0man add <path> [name] [flags]
```

## Arguments

- `path` (required) - Path to file or directory to add (e.g., `~/.vimrc`, `~/.config/nvim`)
- `name` (optional) - Custom name for the dotfile. If omitted, auto-generated from the path

## Flags

| Flag | Description |
|------|-------------|
| `-i, --ignore strings` | Pattern(s) to ignore during sync (e.g., `--ignore "*.log" --ignore "tmp/*"`) |
| `--no-security` | Skip security scanning for this add operation |

## Description

The `add` command moves a file or directory to the n0man store and creates a symlink in its original location. This allows you to:

1. Track dotfiles in version control (Git)
2. Sync them across machines
3. Keep your home directory clean with symlinks
4. Maintain a backup before destructive operations

### Auto-Naming Logic

If no custom `name` is provided, n0man attempts to generate a sensible name:

- For `~/.config/app/file`: Uses `app/file`
- For `~/.vimrc`: Uses `.vimrc` or `vimrc`
- For other paths: Uses the base filename

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully added |
| 1 | Path does not exist |
| 1 | Configuration not found (run `n0man init`) |
| 1 | Store path not configured |
| 1 | Naming collision with existing tracked file |
| 1 | Security scan failed (secrets detected) |
| 1 | Failed to move file or create symlink |

## Examples

### Add a Single File

```bash
$ n0man add ~/.vimrc
Moving /home/user/.vimrc → /home/user/.local/share/n0man/store/.vimrc
Linking /home/user/.local/share/n0man/store/.vimrc → /home/user/.vimrc
✅ Added '.vimrc' (~/.vimrc)
```

### Add with Custom Name

```bash
$ n0man add ~/.config/nvim/init.lua nvim-config
Moving /home/user/.config/nvim/init.lua → /home/user/.local/share/n0man/store/nvim-config
Linking /home/user/.local/share/n0man/store/nvim-config → /home/user/.config/nvim/init.lua
✅ Added 'nvim-config' (~/.config/nvim/init.lua)
```

### Add a Directory with Ignore Patterns

```bash
$ n0man add ~/.config/nvim nvim -i "*.swap" -i "backup/*" -i "*.log"
Moving /home/user/.config/nvim → /home/user/.local/share/n0man/store/nvim
Linking /home/user/.local/share/n0man/store/nvim → /home/user/.config/nvim
✅ Added 'nvim' (~/.config/nvim)
```

### Add Without Security Scan

```bash
$ n0man add ~/.ssh/config --no-security
Moving /home/user/.ssh/config → /home/user/.local/share/n0man/store/ssh-config
Linking /home/user/.local/share/n0man/store/ssh-config → /home/user/.ssh/config
✅ Added 'ssh-config' (~/.ssh/config)
```

## What It Does

1. Resolves absolute path and checks if it exists
2. Generates or validates the dotfile name
3. **Security Scan** (unless `--no-security`):
   - Scans file/directory for secrets
   - If secrets found and `fail_on_secrets=true`: blocks operation
   - If secrets found and `interactive=true`: prompts to continue
4. Moves the file/directory to the store
5. Creates a symlink at the original location pointing to the store
6. Updates configuration with the new dotfile mapping
7. Applies ignore patterns to `.gitignore` (on next `sync`)

## Security Scanning

By default, `add` runs a security scan before adding files. It checks for:

- High-risk file patterns (`.env`, `.pem`, SSH keys, etc.)
- API keys, passwords, tokens in content
- High-entropy strings that may indicate secrets

If secrets are detected, the operation will fail. Use `--no-security` to skip scanning only for files you know are safe.

## Ignore Patterns

Use the `-i, --ignore` flag to specify patterns that should be ignored during sync. These are added to `.gitignore` automatically:

```bash
n0man add ~/.config/app app -i "*.log" -i "cache/*" -i "tmp/*"
```

This will save to configuration and be applied to `.gitignore` during the next `sync` operation:
```
app/*.log
app/cache/*
app/tmp/*
```

**Note**: Ignore patterns are saved to configuration immediately but added to `.gitignore` during `sync` to ensure proper Git tracking.

## Error Handling

- **Path doesn't exist**: Command fails with descriptive message
- **Already tracked**: If path is already tracked with the same name, informs and exits cleanly
- **Naming collision**: If name exists with a different path, suggests using a custom name
- **Symlink creation fails**: Automatically rolls back the file move to original location

## After Adding

After running `add`, the file is now tracked. You can:

- View tracked files: `n0man list`
- Check status: `n0man status`
- Sync to remote: `n0man sync`

## Notes

- The original file is moved, not copied. Use backups before adding critical files.
- Symlinks are created at the original location
- Directories are added recursively
- Configuration is updated automatically
- The operation is atomic: either fully succeeds or fully rolls back
