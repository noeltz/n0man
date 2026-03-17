# `n0man doctor`

Run comprehensive health checks on your dotfiles setup with interactive fixes.

## Usage

```bash
n0man doctor [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `-f, --fix` | Automatically fix issues without prompting |

## Description

The `doctor` command performs a thorough diagnostic of your n0man setup. It checks:

1. **Store Directory**: Verifies the store exists and is accessible
2. **Symlink Integrity**: Validates all tracked dotfiles have correct symlinks
3. **Stale Entries**: Identifies files in store not tracked in configuration
4. **Git Health**: Checks Git repository status and remote configuration
5. **Configuration**: Validates settings and paths

Unlike `status`, `doctor` is interactive and can automatically fix many issues.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 0 | Issues found (some may have been fixed) |
| 1 | Failed to load configuration |

## Examples

### All Systems Healthy

```bash
$ n0man doctor
n0man doctor
Store Directory
  ✅ /home/user/.local/share/n0man/store

Symlink Integrity
  ✅ vim
  ✅ nvim
  ✅ zsh

Stale Entries
  ✅ No stale entries

Git Health
  ✅ Clean working tree
  ℹ️  No remote configured (local-only mode)

Configuration
  ℹ️  No remote_url set (local-only mode)
  ✅ local_path: /home/user/.local/share/n0man/store
  ✅ max_backups: 5

✅ All checks passed! Your dotfiles are healthy.
```

### Issues Found (Interactive Fixes)

```bash
$ n0man doctor
n0man doctor
Store Directory
  ✅ /home/user/.local/share/n0man/store

Symlink Integrity
  ✅ vim
  ⚠️  nvim: Symlink → /tmp/nvim (expected /home/user/.local/share/n0man/store/nvim)
  ✅ zsh

Stale Entries
  ✅ No stale entries

Git Health
  ✅ Clean working tree
  ℹ️  No remote configured (local-only mode)

Configuration
  ℹ️  No remote_url set (local-only mode)
  ✅ local_path: /home/user/.local/share/n0man/store
  ✅ max_backups: 5

⚠️  Found 1 issue(s). Review above for details.

  ? Would you like to fix 1 issue(s) automatically? (Y/n)
  > Yes

  → Fixing issues...
  ✓ ✓ Fixed: Symlinks
  ✓ Fixed 1 issue(s)!

  → Running checks again to verify fixes...

✅ All checks passed! Your dotfiles are healthy.
```

### Automatic Fix Mode

```bash
$ n0man doctor --fix
n0man doctor
Store Directory
  ✅ /home/user/.local/share/n0man/store

Symlink Integrity
  ⚠️  nvim: Symlink missing
  ✅ vim

Stale Entries
  ✅ No stale entries

Git Health
  ✅ Clean working tree

Configuration
  ✅ local_path: /home/user/.local/share/n0man/store

⚠️  Found 1 issue(s). Review above for details.

  → Fixing issues...
  ✓ ✓ Fixed: Symlinks
  ✓ Fixed 1 issue(s)!

  → Running checks again to verify fixes...

✅ All checks passed! Your dotfiles are healthy.
```

### Store Directory Missing

```bash
$ n0man doctor
n0man doctor
Store Directory
  ❌ Store missing: /home/user/.local/share/n0man/store

Symlink Integrity
  ❌ .bashrc: Store file missing
  ❌ .vimrc: Store file missing

⚠️  Found 3 issue(s). Review above for details.
```

### Multiple Issues

```bash
$ n0man doctor
n0man doctor
Store Directory
  ✅ /home/user/.local/share/n0man/store

Symlink Integrity
  ✅ vim
  ❌ nvim: Symlink missing at ~/.config/nvim
  ⚠️  zsh: ~/.zshrc exists but is NOT a symlink

Stale Entries
  ✅ No stale entries

Git Health
  ⚠️  Not a Git repository
  ⚠️  Uncommitted changes detected

Configuration
  ℹ️  No remote_url set (local-only mode)
  ✅ local_path: /home/user/.local/share/n0man/store
  ✅ max_backups: 5

⚠️  Found 5 issue(s). Review above for details.

  ? Would you like to fix 2 issue(s) automatically? (Y/n)
  > Yes

  → Fixing issues...
  ✓ ✓ Fixed: Symlinks
  ✓ Fixed 1 issue(s)!
```

## What It Checks

### 1. Store Directory

- **Exists**: Checks if store directory exists
- **Accessible**: Can read/write to the directory
- **Status**: ✅ Success or ❌ Missing

No interactive fix - you must create the directory manually.

### 2. Symlink Integrity

For each tracked dotfile:

- **Store file exists**: Checks if file exists in store
- **Symlink exists**: Checks if symlink exists at target
- **Is symlink**: Verifies target is actually a symlink
- **Correct target**: Ensures symlink points to correct store location

Status indicators:
- ✅: Everything correct
- ⚠️: Fixable issue (offers to recreate)
- ❌: Critical issue (cannot fix automatically)

**Interactive Fixes:**
- Wrong symlink target: Offers to recreate
- Missing symlink: No auto-fix (manual intervention needed)

### 3. Stale Entries

Identifies files/directories in store that are not tracked in configuration. This typically happens when:
- A file was removed from tracking with `n0man rm` (the restored file remains in the store)
- Files were manually added to the store
- Configuration was manually edited

Status indicators:
- ✅: No stale entries
- ⚠️: Stale entries found

**Interactive Fixes:**
- Add to config: Prompts for target path and adds to configuration
- Remove from store: Deletes file/directory from store
- Ignore: Leaves as-is

**Note**: When you run `n0man rm`, the file is restored from the store to its original location as a regular file. The copy remains in the store as a "stale entry" until you either:
- Re-add the file with `n0man add`
- Manually remove it from the store
- Use the doctor to clean it up

### 4. Git Health

- **Repository exists**: Checks if `.git` directory exists
- **Clean working tree**: Runs `git status --porcelain`
- **Remote configured**: Checks `git remote -v`

Status indicators:
- ✅: Healthy
- ⚠️: Warning (not critical)

No interactive fixes - manual Git intervention required.

### 5. Configuration

- **Remote URL**: Shows configured remote (or notes local-only mode)
- **Local Path**: Shows store path
- **Max Backups**: Shows backup retention setting

Always shows ✅ (informational only).

## Interactive Prompts

### Fix Symlink

When a symlink has the wrong target:

```
? Fix symlink for 'nvim'?
  Recreate symlink
  Ignore
```

- **Recreate symlink**: Removes existing symlink and creates correct one
- **Ignore**: Leaves as-is (manual fix required)

### Handle Stale Entry

When a file exists in store but not in configuration:

```
? What to do with stale entry 'oldconfig'?
  Add to config
  Remove from store
  Ignore
```

- **Add to config**: Prompts for target path and adds to configuration
- **Remove from store**: Deletes the file/directory from store
- **Ignore**: Leaves as-is

## Output Sections

### Store Directory

```
✅ /home/user/.local/share/n0man/store
```

### Symlink Integrity

```
✅ name
⚠️  name: Warning message
❌ name: Error message
```

### Stale Entries

```
✅ No stale entries
⚠️  'name' in store but NOT in config
```

### Git Health

```
✅ Clean working tree
✅ Remote configured
⚠️  Uncommitted changes detected
⚠️  Not a Git repository
```

### Configuration

```
✅ remote_url: git@github.com:user/dotfiles.git
✅ local_path: /home/user/.local/share/n0man/store
✅ max_backups: 5
ℹ️  No remote_url set (local-only mode)
```

## Comparison with `status`

| Feature | `status` | `doctor` |
|---------|----------|----------|
| Checks store | No | Yes |
| Checks symlinks | Yes | Yes |
| Detects stale entries | No | Yes |
| Checks Git | Yes | Yes |
| Interactive fixes | No | Yes |
| Passive/Active | Passive | Active |

Use `status` for quick checks, `doctor` for comprehensive diagnostics and fixes.

## Use Cases

### After Manual Changes

```bash
# You manually edited files in store
$ n0man doctor
# Check for issues and fix them
```

### Before Sync

```bash
$ n0man doctor
# Ensure everything is healthy
$ n0man sync
# Sync with confidence
```

### Troubleshooting Issues

```bash
$ n0man status
# See what's wrong
$ n0man doctor
# Fix what can be fixed
```

### Regular Maintenance

```bash
# Run periodically to catch issues early
$ n0man doctor
```

## Manual Fixes for Non-Interactive Issues

### Store Directory Missing

```bash
mkdir -p ~/.local/share/n0man/store
cd ~/.local/share/n0man/store
git init
```

### Missing Symlink

```bash
ln -s ~/.local/share/n0man/store/<name> <target-path>
```

### Git Issues

```bash
cd ~/.local/share/n0man/store
# Fix uncommitted changes
git status
git add .
git commit -m "Fix issues"

# Initialize if not a repo
git init
git add .
git commit -m "Initial commit"

# Add remote
git remote add origin git@github.com:user/dotfiles.git
```

## Notes

- Doctor is safe to run at any time
- It reads and inspects but only modifies when you confirm
- Some issues require manual intervention
- Configuration is updated automatically when fixes are applied
- Git operations are not performed automatically

## Best Practices

1. **Run regularly**: Periodically check system health
2. **Before sync**: Run before `n0man sync` to ensure clean state
3. **After manual edits**: If you manually modify files, run doctor
4. **Troubleshooting**: Start with doctor when something seems wrong
5. **Review fixes**: Don't blindly accept all fixes - review each prompt
