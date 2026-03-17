# `n0man list`

List all tracked dotfiles.

## Usage

```bash
n0man list
```

## Description

The `list` command displays all dotfiles currently tracked by n0man, including their names, target paths, and ignore patterns. This is useful for:

- Quickly seeing what's being managed
- Verifying dotfile configuration
- Checking ignore patterns
- Identifying files before removal

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully listed (or no files tracked) |
| 1 | Failed to load configuration |

## Examples

### Multiple Tracked Files

```bash
$ n0man list
n0man list
  NAME                 TARGET PATH                      IGNORES
  ──────────────────────────────────────────────────────────────
  vim                  ~/.vimrc                          
  nvim                 ~/.config/nvim                   *.swap, backup/*
  zsh                  ~/.zshrc                          *.log
  git                  ~/.gitconfig                     
  ssh                  ~/.ssh/config                    *.pem

  Store: /home/user/.local/share/n0man/store
```

### No Files Tracked

```bash
$ n0man list
ℹ️  No dotfiles tracked. Use 'n0man add <path>' to start.
```

### Single File

```bash
$ n0man list
n0man list
  NAME                 TARGET PATH                      IGNORES
  ──────────────────────────────────────────────────────────────
  vim                  ~/.vimrc                          

  Store: /home/user/.local/share/n0man/store
```

## Output Format

The output is displayed in a table with three columns:

| Column | Description |
|--------|-------------|
| NAME | The name used to identify the dotfile in n0man |
| TARGET PATH | The original location where the symlink is created |
| IGNORES | Comma-separated list of ignore patterns (if any) |

At the bottom, the store path is shown for reference.

## What It Shows

### Name

The internal name used by n0man to track the dotfile. This is the name used with commands like:

```bash
n0man rm <name>
```

### Target Path

The original location of the file/directory. This is where the symlink is created.

- Paths may contain `~` for home directory
- Shows the actual symlink location on your system
- Used for reference and verification

### Ignores

Any ignore patterns configured for this dotfile:

- If empty: No ignore patterns
- If present: Comma-separated list (e.g., `*.log, cache/*, tmp/*`)
- These patterns are added to `.gitignore`

## Use Cases

### Check What's Tracked Before Sync

```bash
$ n0man list
# Review the list
$ n0man status
# Check health
$ n0man sync
# Sync
```

### Find File to Remove

```bash
$ n0man list
# Find the name you want to remove
$ n0man rm nvim
```

### Verify Configuration

```bash
$ n0man list
# Check target paths are correct
# Check ignore patterns are applied
```

### Review After Adding

```bash
$ n0man add ~/.config/nvim
✅ Added 'nvim' (~/.config/nvim)

$ n0man list
# Verify it's in the list
```

## Comparison with `status`

| Command | Shows | Purpose |
|---------|-------|---------|
| `list` | Tracked dotfiles, paths, ignores | Overview of what's managed |
| `status` | Symlink health, Git state, drift | Health check and diagnostics |

Use `list` to see **what** is tracked, use `status` to see **how** it's tracking.

## Notes

- Only shows dotfiles in configuration
- Does not verify symlinks (use `status` for that)
- Does not check Git state (use `status` for that)
- Ignores are shown as comma-separated for brevity
- Empty ignore column means no patterns

## Tips

1. **Pipe to grep**: Filter the list
   ```bash
   n0man list | grep nvim
   ```

2. **Count files**: See how many you're tracking
   ```bash
   n0man list | grep -c "^  "
   ```

3. **Review regularly**: Periodically review your tracked files

4. **Document in README**: Use the output to document your dotfiles setup

5. **Check ignores**: Ensure sensitive patterns are being ignored
