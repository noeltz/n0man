# `n0man self-update`

Update n0man to the latest version from the configured repository.

## Usage

```bash
n0man self-update
```

## Description

The `self-update` command updates n0man to the latest version by building from source and atomically replacing the current binary. This ensures you always have the latest features and bug fixes without needing to manually download or reinstall.

The update process:
1. Acquires a lock to prevent concurrent updates
2. Detects the current binary location using [`os.Executable()`](https://pkg.go.dev/os#Executable)
3. Checks for Go installation and version
4. Checks for a newer version from GitHub releases API
5. Downloads and builds the latest version to a temporary location
6. Verifies the new binary (file size, functionality)
7. Creates a backup of the current binary
8. Atomically replaces the old binary with the new one
9. Sets executable permissions (0755)
10. Removes the backup on success, or restores it on failure

## Installation

### Quick Install

The recommended way to install n0man is using the one-line installation script:

```bash
curl -sSL https://raw.githubusercontent.com/noeltz/n0man/main/install.sh | bash
```

This will:
- Check prerequisites (bash, curl/wget, git, Go 1.22+)
- Acquire a lock to prevent concurrent installations
- Create `$HOME/.local/bin` if it doesn't exist
- Clone the repository with `--depth 1` for faster downloads
- Build the binary with version information embedded
- Install it to `$HOME/.local/bin/n0man`
- Set proper permissions (0755)
- Verify the installation
- Check if `$HOME/.local/bin` is in PATH
- Clean up temporary files and lock file

### Manual Install

If you prefer to install manually:

```bash
git clone https://github.com/noeltz/n0man.git
cd n0man
go build -o $HOME/.local/bin/n0man ./cmd/n0man
chmod 0755 $HOME/.local/bin/n0man
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully updated or already up to date |
| 1 | General error during update |
| 2 | Prerequisites not met (Go, git, or curl/wget not installed) |
| 3 | Lock acquisition failed (another update/install in progress) |
| 4 | Build/download failed |
| 5 | Verification failed |
| 6 | Permission denied |

### Exit Code Details

- **Code 0**: Success - either updated to new version or already at latest version
- **Code 1**: General error - catch-all for unexpected errors
- **Code 2**: Prerequisites not met - required tools (Go, git, curl/wget) are missing
- **Code 3**: Lock acquisition failed - another update or installation is in progress (lock file exists and is recent)
- **Code 4**: Build/download failed - failed to clone repository or build binary
- **Code 5**: Verification failed - binary doesn't meet size requirements or is not functional
- **Code 6**: Permission denied - cannot write to installation location

## Examples

### Successful Update

```bash
$ n0man self-update
🔄 Updating n0man...

  Current binary: /home/user/.local/bin/n0man
  Platform: linux/amd64
  Using Go at: /usr/local/go/bin/go
  go version go1.22.0 linux/amd64
  Current version: 1.2.3
  Latest version: 1.2.4

  Building new version...
  Verifying new binary...
  Replacing binary...

✅ n0man updated successfully!
  New version: 1.2.4

  Run 'n0man --help' to verify the update.
```

### Already Up to Date

```bash
$ n0man self-update
🔄 Updating n0man...

  Current binary: /home/user/.local/bin/n0man
  Platform: linux/amd64
  Using Go at: /usr/local/go/bin/go
  go version go1.22.0 linux/amd64
  Current version: 1.2.4
  Latest version: 1.2.4

✅ Already up to date!
```

### Another Update in Progress

```bash
$ n0man self-update
🔄 Updating n0man...

Error: failed to acquire update lock: another update is in progress
```

### Go Not Installed

```bash
$ n0man self-update
🔄 Updating n0man...

  Current binary: /home/user/.local/bin/n0man
  Platform: linux/amd64
Error: Prerequisite check: go is not installed or not in PATH

💡 Suggestion: Install Go from https://go.dev/dl/
```

### Build Failed

```bash
$ n0man self-update
🔄 Updating n0man...

  Current binary: /home/user/.local/bin/n0man
  Platform: linux/amd64
  Using Go at: /usr/local/go/bin/go
  go version go1.22.0 linux/amd64
  Current version: 1.2.3
  Latest version: 1.2.4

  Building new version...
Error: Build: build failed: ...
💡 Suggestion: Check your Go installation and try again
```

## How It Works

### 1. Lock Mechanism

The command uses a lock file (`/tmp/n0man-update.lock`) to prevent multiple concurrent updates. The lock file contains:
- Process ID (PID) of the update process
- Timestamp of lock acquisition

If a lock file exists and is less than 30 minutes old, the command will fail with an error. Stale locks (older than 30 minutes) are automatically removed and the update proceeds.

**Lock file format:**
```
<pid>
<timestamp in RFC3339 format>
```

The install script uses a separate lock file (`/tmp/n0man-install.lock`) with a 1-hour timeout to prevent concurrent installations.

### 2. Binary Location Detection

The command uses [`os.Executable()`](https://pkg.go.dev/os#Executable) to detect where the currently running binary is located. It also resolves symlinks using [`filepath.EvalSymlinks()`](https://pkg.go.dev/path/filepath#EvalSymlinks) to ensure it's working with the real file path, then converts to an absolute path.

This ensures that:
- The update happens at the actual binary location (not symlink location)
- The binary path is always absolute and canonical
- Symlinks don't interfere with the update process

### 3. Version Checking

The command fetches the latest version from the GitHub Releases API:
- URL: `https://api.github.com/repos/noeltz/n0man/releases/latest`
- Parses the `tag_name` field from the JSON response
- Compares with current version (obtained from `n0man version --short`)
- If versions match, no update is performed

### 4. Build Process

The command builds the new version from source:

1. **Create temporary directory**: Uses `os.MkdirTemp()` with prefix "n0man-build-*"
2. **Clone repository**: Uses `git clone --depth 1` for shallow clone (faster download)
3. **Build binary**: Uses `go build -o <path> ./cmd/n0man`
4. **Context timeout**: Build operations have a 5-minute timeout
5. **Automatic cleanup**: Temporary directory is removed after build (or on failure)

The build process requires:
- Go 1.22 or higher
- Network access to GitHub
- Sufficient disk space for the build

### 5. Atomic Replacement

The command uses atomic file operations to safely replace the binary:

1. **Create backup**: Renames current binary to `<path>.backup`
2. **Move new binary**: Renames new binary to original path
3. **Set permissions**: Sets executable permissions (0755)
4. **Remove backup**: Removes backup on success

**Automatic rollback:**
- If step 2 fails, the backup is automatically restored
- If step 3 fails, the backup is automatically restored
- This ensures the user is never left with a broken binary

### 6. Verification

The new binary is verified before replacement:

**File existence check:**
- Ensures the binary file exists at the expected path

**File size validation:**
- Minimum: 1MB (1,000,000 bytes)
- Maximum: 100MB (100,000,000 bytes)
- Prevents corrupted or incomplete downloads

**Functionality test:**
- Executes the binary with `--help` flag
- Ensures the binary is a valid, functional Go executable
- Catches corrupted or invalid binaries

### 7. Error Handling

The command uses a structured error type `UpdateError` that includes:

```go
type UpdateError struct {
    Step       string  // Which step failed
    Err        error   // The underlying error
    CanRecover bool    // Whether the user can retry
    Suggestion string  // Helpful suggestion for the user
}
```

Each error provides:
- Clear indication of what step failed
- The underlying error details
- Whether the error is recoverable
- Suggested actions for the user

## Requirements

### Go Installation

Go 1.22+ is required. Install from:

- **Linux**: `sudo apt install golang` (Ubuntu/Debian) or download from go.dev
- **macOS**: `brew install go` or download from go.dev
- **Windows**: Download installer from go.dev

The install script checks for Go version and will fail if version is less than 1.22.

### Network Connection

An internet connection is required to:
- Fetch the latest version from GitHub API
- Clone the repository
- Download Go dependencies during build

### Write Permissions

You need write permissions to the directory containing the n0man binary. The recommended installation location is `$HOME/.local/bin/n0man`, which should be writable by your user.

### Git

Git is required for cloning the repository during the update process.

## Binary Location

The binary is updated in-place at its current location. The recommended installation path is:

```
$HOME/.local/bin/n0man
```

### Common Paths

| Installation Method | Default Path |
|---------------------|--------------|
| install.sh script | `$HOME/.local/bin/n0man` |
| Manual build | Custom path you specify |
| go install | `~/go/bin/n0man` or `$GOBIN/n0man` |

Ensure your PATH includes the directory containing n0man.

## After Update

### Verify Update

```bash
# Check version
n0man version --short

# Verify command is working
n0man --help
```

No shell restart is required since the binary is updated in-place.

### Check Binary Location

```bash
# Find where n0man is installed
which n0man

# Check if it's a symlink
ls -la $(which n0man)
```

## Manual Update

If self-update doesn't work, you can update manually:

### Using install.sh

```bash
curl -sSL https://raw.githubusercontent.com/noeltz/n0man/main/install.sh | bash
```

This will:
- Check all prerequisites
- Acquire installation lock
- Clone and build the latest version
- Replace the existing binary
- Verify the installation

### From Source

```bash
git clone https://github.com/noeltz/n0man.git
cd n0man
go build -o $HOME/.local/bin/n0man ./cmd/n0man
chmod 0755 $HOME/.local/bin/n0man
```

### Using go install

```bash
go install github.com/noeltz/n0man/cmd/n0man@latest
```

Note: This will install to `~/go/bin/n0man` or `$GOBIN/n0man`, which may be different from your current installation location.

## Troubleshooting

### Permission Denied

```bash
# Check binary permissions
ls -la $HOME/.local/bin/n0man

# Should be executable (rwxr-xr-x)
chmod 0755 $HOME/.local/bin/n0man

# Check directory permissions
ls -la $HOME/.local/bin
```

### Another Update in Progress

If you see "another update is in progress" but no update is actually running:

```bash
# Check if lock file exists
ls -la /tmp/n0man-update.lock

# View lock file contents (PID and timestamp)
cat /tmp/n0man-update.lock

# Remove stale lock (if older than 30 minutes)
rm /tmp/n0man-update.lock
```

For installation lock issues:
```bash
# Check install lock
ls -la /tmp/n0man-install.lock

# Remove stale install lock (if older than 1 hour)
rm /tmp/n0man-install.lock
```

### Update Fails

```bash
# Check network connection
ping github.com

# Check Go version
go version

# Check disk space
df -h

# Check git is installed
git --version

# Try manual update
curl -sSL https://raw.githubusercontent.com/noeltz/n0man/main/install.sh | bash
```

### Binary Not Found After Update

```bash
# Check if binary exists
ls -la $HOME/.local/bin/n0man

# Check if PATH includes the directory
echo $PATH | grep .local/bin

# If not in PATH, add to your shell configuration
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc
```

### Backup Restoration

If an update fails and leaves you with a broken binary, check for the backup:

```bash
# Check if backup exists
ls -la $HOME/.local/bin/n0man.backup

# Restore from backup
mv $HOME/.local/bin/n0man.backup $HOME/.local/bin/n0man

# Make it executable
chmod 0755 $HOME/.local/bin/n0man
```

### Go Version Too Old

If you see an error about Go version:

```bash
# Check current Go version
go version

# If version is < 1.22, upgrade Go
# Ubuntu/Debian:
sudo apt update
sudo apt install golang-go

# macOS:
brew upgrade go

# Or download from https://go.dev/dl/
```

### Build Timeout

If the build times out (after 5 minutes):

```bash
# Check your internet connection
ping github.com

# Try manual build
git clone https://github.com/noeltz/n0man.git
cd n0man
go build -o $HOME/.local/bin/n0man ./cmd/n0man
```

## Use Cases

### Regular Updates

```bash
# Update monthly or after new features announced
$ n0man self-update
```

### After Bug Fixes

```bash
# If you encounter a bug that's been fixed
$ n0man self-update
```

### Before New Features

```bash
# Check release notes for new features
$ n0man self-update
```

### Automated Updates

You can set up a cron job or systemd timer to automatically update n0man:

```bash
# Add to crontab (monthly)
0 0 1 * * $HOME/.local/bin/n0man self-update

# Or weekly
0 0 * * 0 $HOME/.local/bin/n0man self-update
```

**Note:** Automated updates should only be used if you're comfortable with potential changes and have tested updates manually first.

## Security Considerations

### Verification

The self-update command verifies the new binary before replacement:
- Checks file size is within reasonable bounds (1MB - 100MB)
- Verifies the binary is functional (responds to `--help`)
- Prevents installation of corrupted or incomplete binaries

### Atomic Operations

The update uses atomic file operations to prevent corruption:
- Old binary is backed up before replacement
- New binary is written to a temporary location first
- Atomic rename ensures no partial updates
- Automatic rollback on failure at any step

### Lock File

The lock file prevents race conditions:
- Only one update can run at a time
- Stale locks are automatically cleaned up (after 30 minutes)
- Lock file contains PID and timestamp for debugging
- Separate lock files for updates and installations

### Permissions

The binary is installed with secure permissions:
- Owner read/write/execute (0755)
- No world-writable permissions
- Verified after installation

### Path Validation

The self-update command:
- Resolves symlinks to get the real binary path
- Converts to absolute path
- Ensures the binary is in a writable location
- Prevents path traversal attacks

### Network Security

- Uses HTTPS for all network requests
- Fetches from official GitHub repository
- No checksum verification currently implemented (future enhancement)
- Downloads from trusted source (GitHub)

## Cross-Platform Support

### Linux

- Fully supported
- Uses `/tmp` for lock files
- Supports ext4 and other Linux filesystems
- Atomic renames work within the same filesystem
- Tested on Ubuntu, Debian, and other distributions

### macOS

- Fully supported
- Uses `/tmp` for lock files
- Supports APFS filesystem
- Atomic renames work within the same filesystem
- Tested on macOS 10.15+

### Platform-Specific Differences

The install script handles platform differences for lock file age calculation:
- **Linux**: Uses `stat -c %Y` for modification time
- **macOS**: Uses `stat -f %m` for modification time

Both platforms use the same installation path (`$HOME/.local/bin/n0man`) and the same binary format.

## Installation Script Details

The [`install.sh`](../../install.sh) script provides a one-line installation method:

```bash
curl -sSL https://raw.githubusercontent.com/noeltz/n0man/main/install.sh | bash
```

### What the Install Script Does

1. **Prerequisites Check**: Verifies bash, curl/wget, git, and Go 1.22+ are installed
2. **Lock Acquisition**: Prevents concurrent installations using `/tmp/n0man-install.lock` (1-hour timeout)
3. **Directory Creation**: Creates `$HOME/.local/bin` if it doesn't exist
4. **Repository Clone**: Clones the repository to a temporary directory with `--depth 1` for speed
5. **Build**: Compiles the binary with version and build time information embedded via ldflags
6. **Installation**: Copies the binary to `$HOME/.local/bin/n0man`
7. **Permissions**: Sets executable permissions (0755)
8. **Verification**: Verifies the binary is functional
9. **PATH Check**: Checks if `$HOME/.local/bin` is in PATH and provides instructions if not
10. **Cleanup**: Removes temporary files and lock file automatically on exit

### Install Script Features

- **Idempotent**: Can be run multiple times safely
- **Error Handling**: Clear error messages with suggestions
- **Cleanup**: Automatic cleanup of temporary files on exit (using trap)
- **Cross-platform**: Works on Linux and macOS
- **Version Information**: Embeds version and build time in the binary
- **User Feedback**: Colored output with progress indicators (ℹ, ✓, ✗, ⚠)
- **Lock Mechanism**: Prevents concurrent installations with automatic stale lock cleanup
- **Go Version Check**: Verifies Go 1.22+ is installed before proceeding

### Install Script Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Prerequisites not met |
| 3 | Lock acquisition failed |
| 4 | Build failed |
| 5 | Verification failed |
| 6 | Permission denied |

## Comparison with Other Methods

| Method | Advantages | Disadvantages |
|--------|------------|---------------|
| `self-update` | In-place update, automatic rollback, works everywhere, atomic operations | Requires Go, git, and network |
| install.sh script | One-line install, consistent path, verification, lock mechanism | Requires Go, git, and curl/wget |
| Package manager (apt, brew) | Easy, automatic updates | May be outdated, not available for all platforms |
| Manual build | Full control, customizable | More steps, manual updates, no automatic rollback |
| go install | Simple command | Installs to different location, no automatic rollback |

## Best Practices

1. **Regular updates**: Update monthly to get bug fixes and new features
2. **Verify after update**: Run `n0man --help` to confirm the update worked
3. **Check prerequisites**: Ensure Go 1.22+, git, and curl/wget are installed before updating
4. **Review changes**: Check GitHub releases for what's new before updating
5. **Backup important data**: Although the update is safe, it's good practice to have backups
6. **Use recommended installation**: Install to `$HOME/.local/bin` for consistency
7. **Test updates**: Test updates in a non-critical environment first if possible
8. **Monitor logs**: Watch for error messages during update process
9. **Handle lock files**: Know how to remove stale lock files if needed
10. **Check PATH**: Ensure `$HOME/.local/bin` is in your PATH

## Notes

- Updates to the latest version from GitHub releases
- Requires internet connection to download
- Requires Go 1.22+ to compile
- Does not affect your dotfiles or configuration
- Old binary is backed up during update and removed on success
- Works on Linux and macOS
- Atomic operations prevent corruption
- Automatic rollback on failure
- Lock mechanism prevents concurrent updates
- Build timeout of 5 minutes for safety
- Binary size validation (1MB - 100MB)
- Symlinks are resolved to real paths

## Related Commands

- `n0man version --short`: Check current version
- `n0man --help`: Verify installation
- Configuration in `n0man.toml`: Not affected by updates

## Environment Variables

The following environment variables can be used to customize behavior (future enhancement):

| Variable | Purpose | Default |
|----------|---------|---------|
| `N0MAN_INSTALL_DIR` | Installation directory | `$HOME/.local/bin` |
| `N0MAN_LOCK_TIMEOUT` | Lock timeout in minutes | `30` |
| `N0MAN_UPDATE_TIMEOUT` | Update timeout in minutes | `5` |

Note: These are planned for future implementation and are not currently supported.

## Technical Details

### Binary Update Algorithm

```
1. Acquire lock (/tmp/n0man-update.lock)
2. Get current binary path (resolve symlinks)
3. Check Go installation
4. Fetch latest version from GitHub API
5. Compare versions
6. If update needed:
   a. Create temp directory
   b. Clone repository (shallow clone)
   c. Build binary with timeout (5 min)
   d. Verify binary (size, functionality)
   e. Create backup of current binary
   f. Atomic rename (new -> old)
   g. Set permissions (0755)
   h. Remove backup
7. Release lock
```

### Lock File Lifecycle

```
[No Lock] -> [Acquiring] -> [Locked] -> [Releasing] -> [No Lock]
                    |              |
                    v              v
                [Failed]     [Stale] -> [Cleanup] -> [No Lock]
```

### Error Recovery

```
Error occurs:
  1. Log error details with context
  2. Clean up temp files
  3. Release lock if held
  4. Restore backup if exists
  5. Show user-friendly error
  6. Provide suggestion for recovery
```

## Implementation Files

- [`internal/cmd/self_update.go`](../../internal/cmd/self_update.go): Self-update command implementation
- [`install.sh`](../../install.sh): Installation script
- [`docs/plans/installation-and-self-update-design.md`](../plans/installation-and-self-update-design.md): Design document

## See Also

- [Getting Started Guide](../guides/getting-started.md)
- [Troubleshooting Guide](../guides/troubleshooting.md)
- [Exit Codes Reference](../references/exit-codes.md)
