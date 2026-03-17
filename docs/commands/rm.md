# `n0man rm`

Stop tracking a dotfile and restore it to its original location.

## Usage

```bash
n0man rm <name>
```

## Arguments

- `name` (required) - Name of the dotfile to remove (as shown in `n0man list`)

## Description

The `rm` command is the inverse of `add`. It:

1. Removes the symlink at the original location
2. Moves the actual file/directory from the store back to its original location
3. Removes the dotfile from n0man configuration
4. Preserves all file contents and metadata

This operation is fully reversible in the sense that the file is restored to its pre-add state.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully removed |
| 1 | Dotfile not found in configuration |
| 1 | Failed to remove symlink |
| 1 | Failed to restore file |
| 1 | Failed to save configuration |

## Examples

### Remove a Tracked Dotfile

```bash
$ n0man list
  NAME                 TARGET PATH                      IGNORES
  ──────────────────────────────────────────────────────────────
  vim                  ~/.vimrc                          
  nvim                 ~/.config/nvim                   *.swap, backup/*

$ n0man rm nvim
Restoring ~/.config/nvim from /home/user/.local/share/n0man/store/nvim
✅ Removed 'nvim' and restored to ~/.config/nvim
```

### Remove After Verification

```bash
$ n0man status
...
  nvim (~/.config/nvim)
...
All systems go — no drift detected

$ n0man rm nvim
Restoring ~/.config/nvim from /home/user/.local/share/n0man/store/nvim
✅ Removed 'nvim' and restored to ~/.config/nvim
```

## What It Does

1. Looks up the dotfile by `name` in configuration
2. Resolves the target path (expands `~` to home directory)
3. Removes the symlink at the target location (if it exists and is a symlink)
4. Moves the file/directory from the store to the target location
5. Removes the dotfile entry from configuration
6. Removes any associated ignore patterns
7. Saves the updated configuration

## Error Handling

- **Dotfile not found**: If `name` is not in configuration, command fails
- **Missing store file**: If the store file doesn't exist, restoration fails
- **Symlink issues**: If target is not a symlink or cannot be removed, command fails
- **Configuration save failure**: Warns but file is already restored

## After Removing

After running `rm`:

- The file is back at its original location as a regular file/directory
- No symlink remains
- The dotfile is no longer tracked in configuration
- The file remains in Git history (if previously synced)

To completely remove from Git, use:

```bash
cd ~/.local/share/n0man/store
git rm -r <name>
git commit -m "Remove <name>"
```

## Notes

- The file/directory content is fully preserved
- The operation is atomic: either fully succeeds or fully fails
- Configuration is updated automatically
- Git history is not modified (use `git rm` for that)
- If you want to remove from all machines, run `n0man rm` then `n0man sync`

## Comparison with `add`

| Operation | `add` | `rm` |
|-----------|-------|------|
| File location | Home → Store | Store → Home |
| Symlink status | Creates symlink | Removes symlink |
| Configuration | Adds entry | Removes entry |
| Git status | Adds to tracking | Not modified |

## Safety

- The `rm` command only removes from n0man tracking
- It does **not** delete files
- It does **not** modify Git history
- The file is safely restored to its original location
