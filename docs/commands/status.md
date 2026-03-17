# `n0man status`

Inspect divergence between your live machine, config, and the repo.

## Usage

```bash
n0man status
```

## Description

The `status` command provides a comprehensive health check of your dotfiles setup. It inspects:

1. **Dotfiles Mapping**: Verifies all tracked dotfiles have correct symlinks pointing to the store
2. **Git Repository**: Shows uncommitted changes and repository status
3. **Divergence Detection**: Identifies drift between live system and n0man tracking

This is the primary diagnostic command for troubleshooting issues.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All systems go — no drift detected |
| 0 | Divergence or issues detected (but command succeeded) |

**Note**: Unlike many commands, `status` returns 0 even when issues are found. Exit code indicates command success, not system health.

## Examples

### Healthy System

```bash
$ n0man status
n0man status
Dotfiles Mapping
  ✅ vim (~/.vimrc)
  ✅ nvim (~/.config/nvim)

Git Repository
  ✅ Clean working tree

✅ All systems go — no drift detected
```

### Missing Symlink

```bash
$ n0man status
n0man status
Dotfiles Mapping
  ✅ vim (~/.vimrc)
  ⚠️  nvim: Exists but NOT a symlink (~/.config/nvim)

Git Repository
  ✅ Clean working tree

⚠️  Divergence or issues detected
```

### Broken Symlink

```bash
$ n0man status
n0man status
Dotfiles Mapping
  ✅ vim (~/.vimrc)
  ❌ nvim: Broken symlink — store file missing

Git Repository
  ✅ Clean working tree

⚠️  Divergence or issues detected
```

### Wrong Symlink Target

```bash
$ n0man status
n0man status
Dotfiles Mapping
  ✅ vim (~/.vimrc)
  ⚠️  nvim: Wrong target (expected /home/user/.local/share/n0man/store/nvim, got /tmp/nvim)

Git Repository
  ✅ Clean working tree

⚠️  Divergence or issues detected
```

### Uncommitted Changes

```bash
$ n0man status
n0man status
Dotfiles Mapping
  ✅ vim (~/.vimrc)
  ✅ nvim (~/.config/nvim)

Git Repository
  ⚠️  Uncommitted changes:
       M vimrc
       ?? newfile.txt

⚠️  Divergence or issues detected
```

### Multiple Issues

```bash
$ n0man status
n0man status
Dotfiles Mapping
  ✅ vim (~/.vimrc)
  ❌ nvim: Broken symlink — store file missing
  ⚠️  ssh: Exists but NOT a symlink (~/.ssh/config)
  ❌ zsh: Target missing (~/.zshrc)

Git Repository
  ⚠️  Uncommitted changes:
       D removed.txt

⚠️  Divergence or issues detected
```

## What It Checks

### 1. Dotfiles Mapping

For each tracked dotfile in configuration:

- **Target exists**: Checks if file exists at the configured path
- **Is symlink**: Verifies it's actually a symlink (not a regular file)
- **Correct target**: Ensures symlink points to the correct store location
- **Store file exists**: Confirms the file in the store is not missing

Status indicators:
- ✅ **Success**: Everything is correct
- ⚠️ **Warning**: Non-critical issue (e.g., not a symlink but file exists)
- ❌ **Error**: Critical issue (e.g., missing store file)

### 2. Git Repository

- **Repository exists**: Checks if `.git` directory exists in store
- **Clean working tree**: Runs `git status --short`
- **Uncommitted changes**: Lists modified/deleted/untracked files

## Output Sections

### Dotfiles Mapping

Shows each tracked dotfile with its target path and status:

```
✅ name (target/path)
⚠️  name: Warning message (target/path)
❌ name: Error message (target/path)
```

### Git Repository

- **Repository exists**: Checks if `.git` directory exists in store
- **Clean working tree**: Runs `git status --short`
- **Uncommitted changes**: Lists modified/deleted/untracked files

**Note**: Files newly added with `n0man add` will appear as untracked (`??`) in Git status until the first `sync` operation, which stages and commits them. This is normal behavior.

**Note**: Files newly added with `n0man add` will appear as untracked (`??`) in Git status until the first `sync` operation, which stages and commits them. This is normal behavior.

### Git Repository

Shows the Git repository status:

```
✅ Clean working tree
⚠️  Uncommitted changes:
     M modified_file
     ?? untracked_file
     D deleted_file
```

### Summary

Final message indicating overall health:

```
✅ All systems go — no drift detected
⚠️  Divergence or issues detected
```

## Common Issues and Solutions

### Issue: "Exists but NOT a symlink"

**Cause**: File exists but is a regular file, not a symlink.

**Solution**:
```bash
# Option 1: Manually remove and let n0man fix it
rm ~/.config/nvim
n0man doctor  # Will offer to recreate symlink

# Option 2: Re-add the file (will overwrite)
n0man add ~/.config/nvim --no-security
```

### Issue: "Broken symlink — store file missing"

**Cause**: Symlink points to store, but store file doesn't exist.

**Solution**:
```bash
# Restore from backup
n0man backup rollback

# Or manually copy file back
cp <original-location> ~/.local/share/n0man/store/<name>
```

### Issue: "Wrong target"

**Cause**: Symlink exists but points to wrong location.

**Solution**:
```bash
# Remove and recreate
rm ~/.config/nvim
ln -s ~/.local/share/n0man/store/nvim ~/.config/nvim
```

### Issue: "Uncommitted changes"

**Cause**: Files in store have been modified outside of n0man.

**Solution**:
```bash
# Review changes
cd ~/.local/share/n0man/store
git status
git diff

# Commit changes
git add .
git commit -m "Describe changes"

# Or discard changes
git reset --hard HEAD
```

## Comparison with `doctor`

| Command | Purpose | Interactive? | Fixes Issues? |
|---------|---------|--------------|---------------|
| `status` | Passive inspection | No | No |
| `doctor` | Active health check | Yes | Can fix issues |

Use `status` for quick checks and `doctor` for comprehensive diagnostics and fixes.

## Notes

- Always run before `sync` to ensure clean state
- Useful for troubleshooting after manual file operations
- Detects drift from external modifications
- Does not modify any files or configuration
- Git must be installed for repository checks

## Best Practices

1. **Run regularly**: Check status periodically to catch issues early
2. **Before sync**: Always run `n0man status` before `n0man sync`
3. **After manual changes**: If you manually edit files in the store, run status
4. **Troubleshooting**: When something seems wrong, start with status
5. **Combine with doctor**: Run `status` first, then `doctor` for fixes
