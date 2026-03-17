# `n0man self-update`

Update n0man to the latest version from the configured repository.

## Usage

```bash
n0man self-update
```

## Description

The `self-update` command updates n0man to the latest version using Go's `go install` mechanism. This ensures you always have the latest features and bug fixes.

The update process:
1. Checks if Go is installed
2. Fetches and compiles the latest version from `github.com/noeltz/n0man/cmd/n0man@latest`
3. Installs the binary to your Go bin directory
4. Reports the new binary location

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successfully updated |
| 1 | Go is not installed |
| 1 | Failed to fetch or compile latest version |

## Examples

### Successful Update

```bash
$ n0man self-update
🔄 Updating n0man...
  Using Go at: /usr/local/go/bin/go
  Current binary: /home/user/go/bin/n0man
  Platform: linux/amd64
  Fetching and building latest version...
go: downloading github.com/noeltz/n0man/cmd/n0man v1.2.3

✅ n0man updated successfully!
  New binary at: /home/user/go/bin/n0man
  Restart your shell or run 'n0man --help' to verify.
```

### Go Not Installed

```bash
$ n0man self-update
🔄 Updating n0man...
Error: Go is not installed or not in PATH. Cannot self-update.
Install Go from https://go.dev/dl/ or update manually
```

### Verify Update

```bash
$ n0man --version
# If version output is available
$ n0man --help
# Verify command is working
```

## What It Does

### 1. Check for Go

```bash
# Looks for go in PATH
which go
```

If Go is not found, the command fails with instructions to install Go.

### 2. Show Current Info

- Go installation path
- Current binary location
- Platform (OS/architecture)

### 3. Fetch and Build

```bash
go install github.com/noeltz/n0man/cmd/n0man@latest
```

This:
- Downloads the latest source code
- Compiles it for your platform
- Installs to your Go bin directory

### 4. Report New Binary Location

Determines where Go installed the new binary:
- `GOBIN` environment variable, or
- `$GOPATH/bin`, or
- `~/go/bin`

## Binary Location

The new binary is installed to:

```
$GOBIN/n0man         # If GOBIN is set
$GOPATH/bin/n0man    # If GOPATH is set
~/go/bin/n0man       # Default location
```

### Common Paths

| Platform | Default Path |
|----------|--------------|
| Linux | `~/go/bin/n0man` |
| macOS | `~/go/bin/n0man` |
| Windows | `%USERPROFILE%\go\bin\n0man.exe` |

Ensure your PATH includes the Go bin directory.

## After Update

### Restart Shell

The new binary won't be in effect until your shell is restarted:

```bash
# Option 1: Close and reopen terminal
# Option 2: Source your shell config
source ~/.bashrc  # or ~/.zshrc

# Option 3: Use the binary directly
~/go/bin/n0man --help
```

### Verify Version

If n0man has a version command:

```bash
$ n0man --version
n0man version 1.2.3
```

## Requirements

### Go Installation

Go 1.22+ is required. Install from:

- **Linux**: `sudo apt install golang` (Ubuntu) or download from go.dev
- **macOS**: `brew install go` or download from go.dev
- **Windows**: Download installer from go.dev

### PATH Configuration

Ensure Go bin directory is in your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:~/go/bin
```

## Manual Update

If self-update doesn't work, update manually:

### From Source

```bash
git clone https://github.com/noeltz/n0man.git
cd n0man
go build -o n0man ./cmd/n0man
sudo mv n0man /usr/local/bin/
```

### From Release (if available)

```bash
wget https://github.com/noeltz/n0man/releases/latest/download/n0man-linux-amd64
chmod +x n0man-linux-amd64
sudo mv n0man-linux-amd64 /usr/local/bin/n0man
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

## Troubleshooting

### Command Not Found After Update

```bash
# Check Go bin is in PATH
echo $PATH | grep go

# Add to PATH if missing
export PATH=$PATH:~/go/bin

# Verify binary exists
ls -la ~/go/bin/n0man
```

### Permission Denied

```bash
# Check binary permissions
ls -la ~/go/bin/n0man

# Should be executable
chmod +x ~/go/bin/n0man
```

### Update Fails

```bash
# Check network connection
ping github.com

# Check Go version
go version

# Manually install
go install github.com/noeltz/n0man/cmd/n0man@latest
```

## Comparison with Package Managers

| Method | Advantages | Disadvantages |
|--------|------------|---------------|
| `self-update` | Latest version, works everywhere | Requires Go, requires PATH setup |
| Package manager (apt, brew) | Easy, automatic | May be outdated |
| Manual build | Full control | More steps |

## Notes

- Updates to the `@latest` tag (main branch or latest release)
- Requires internet connection to download
- Requires Go 1.22+ to compile
- Does not affect your dotfiles or configuration
- Old binary is overwritten (not backed up)
- Works on all platforms (Linux, macOS, Windows)

## Best Practices

1. **Regular updates**: Update monthly to get bug fixes
2. **Verify after update**: Run `n0man --help` to confirm
3. **Restart shell**: Required for PATH changes to take effect
4. **Check Go version**: Ensure you have Go 1.22+
5. **Review changes**: Check GitHub releases for what's new

## Related Commands

- `n0man --version`: Check current version (if available)
- `n0man --help`: Verify installation
- Configuration in `n0man.toml`: Not affected by updates
