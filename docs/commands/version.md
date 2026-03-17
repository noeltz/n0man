# `n0man version`

Print version information for n0man.

## Usage

```bash
n0man version
```

**Alternative:**
```bash
n0man --version
```

## Description

The `version` command displays version information for the installed n0man binary, including:

- n0man version number
- Go version used to build
- Operating system and architecture
- Git commit hash (if available)
- Build timestamp (if available)

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Version information displayed |

## Examples

### Basic Version

```bash
$ n0man version
n0man version 1.0.0
  Go version: go1.26.1
  OS/Arch: linux/amd64
```

### With Build Information (if available)

```bash
$ n0man version
n0man version 1.0.0
  Go version: go1.26.1
  OS/Arch: linux/amd64
  Git commit: abc123def
  Built: 2024-03-12T15:04:05Z
```

### Using --version Flag

```bash
$ n0man --version
n0man version 1.0.0
```

## Output Format

### Standard Output

```
n0man version <version>
  Go version: <go-version>
  OS/Arch: <os>/<arch>
  Git commit: <commit>        # If available
  Built: <timestamp>          # If available
```

### Fields

| Field | Description | Always Present |
|-------|-------------|----------------|
| `n0man version` | Release version number | Yes |
| `Go version` | Go compiler version | Yes |
| `OS/Arch` | Target platform | Yes |
| `Git commit` | Git commit hash | No (build-time) |
| `Built` | Build timestamp | No (build-time) |

## Use Cases

### Verify Installation

```bash
# Check if n0man is installed
n0man version

# Check version in script
if ! n0man version &>/dev/null; then
    echo "n0man not installed"
    exit 1
fi
```

### Debugging

```bash
# Include version in bug report
n0man version >> bug-report.txt
```

### CI/CD

```yaml
# GitHub Actions
- name: Check n0man version
  run: n0man version
```

### Multiple Installations

If you have multiple n0man installations:

```bash
# Check which binary is being used
which n0man
n0man version

# Check specific binary
/usr/local/bin/n0man version
~/go/bin/n0man version
```

## Build-Time Variables

Version information can be enhanced at build time:

```bash
# Build with version info
go build -ldflags "-X github.com/noeltz/n0man/internal/cmd.version=1.0.0 \
                   -X github.com/noeltz/n0man/internal/cmd.gitCommit=abc123 \
                   -X github.com/noeltz/n0man/internal/cmd.buildTime=2024-03-12T15:04:05Z" \
     -o n0man ./cmd/n0man
```

## Comparison with Other Commands

| Command | Output | Use Case |
|---------|--------|----------|
| `n0man version` | Detailed version info | Debugging, bug reports |
| `n0man --version` | Short version | Quick check, scripts |
| `n0man --help` | Help text | Learning commands |

## Troubleshooting

### "command not found"

**Problem:** n0man is not in PATH.

**Solution:**
```bash
# Find n0man
find ~ -name "n0man" -type f 2>/dev/null

# Add to PATH
export PATH=$PATH:~/go/bin
```

### Version Mismatch

**Problem:** Different versions from different locations.

**Solution:**
```bash
# Check all installations
which -a n0man

# Remove old installations
sudo rm /usr/local/bin/n0man

# Reinstall
go install github.com/noeltz/n0man/cmd/n0man@latest
```

## See Also

- [`n0man self-update`](self-update.md) - Update to latest version
- [Getting Started](../guides/getting-started.md) - Installation guide
