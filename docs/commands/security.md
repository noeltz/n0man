# `n0man security scan`

Scan dotfiles for sensitive information.

## Usage

```bash
n0man security scan [path]
```

## Arguments

- `path` (optional) - Path to scan. If omitted, scans the n0man store

## Description

The `security scan` command scans your dotfiles for potential secrets and sensitive information. It uses multiple detection methods:

1. **Pattern-Based Detection**: Identifies high-risk file patterns (`.env`, `.pem`, etc.)
2. **Content Scanning**: Detects API keys, passwords, tokens using regex
3. **Entropy Analysis**: Detects high-entropy strings that may indicate secrets

This command can be run manually to audit your dotfiles before syncing.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No security issues found |
| 1 | Security scan failed (error) |
| 1 | Security issues found (secrets detected) |

## Examples

### Scan Store (No Issues)

```bash
$ n0man security scan
n0man security scan: /home/user/.local/share/n0man/store
✅ No security issues found.
```

### Scan Store (Issues Found)

```bash
$ n0man security scan
n0man security scan: /home/user/.local/share/n0man/store
🚨 Found 3 potential security issue(s):

  1. [CRITICAL] /home/user/.local/share/n0man/store/.env
     → api_key pattern match

  2. [HIGH] /home/user/.local/share/n0man/store/config.yaml
     → password pattern match
     Line: 12: password: "s3cr3t1234"

  3. [HIGH] /home/user/.local/share/n0man/store/private.pem
     → private_key pattern match
     Line: 1: -----BEGIN PRIVATE KEY-----
```

### Scan Custom Path

```bash
$ n0man security scan ~/.config
n0man security scan: /home/user/.local/share/n0man/store/.env
🚨 Found 1 potential security issue(s):

  1. [HIGH] /home/user/.config/api/keys.txt
     → High entropy string
     Line: 3: sk-1234567890abcdef...
```

## What It Scans For

### File Pattern Detection (High Risk)

| Pattern | Risk Level | Examples |
|---------|------------|----------|
| `.env`, `.env.*` | Critical | `.env`, `.env.local`, `.env.production` |
| `*.key`, `*.pem`, `*.p12` | Critical | `api.key`, `cert.pem` |
| `.ssh/id_*`, `.ssh/*_rsa` | Critical | SSH private keys |
| `.aws/credentials` | Critical | AWS credentials |
| `.docker/config.json` | Critical | Docker auth config |
| `.kube/config` | Critical | Kubernetes config |
| `*_history` | High | `.bash_history`, `.zsh_history` |
| `*.db`, `*.sqlite` | High | Databases |
| `.gnupg/secring.gpg` | High | GPG keys |
| `.gitconfig`, `config.yml` | Medium | Configuration files |

### Content Patterns (Regex Detection)

| Pattern Type | Examples |
|---------------|----------|
| API Keys | `sk-...`, `sk-proj-...`, `sk-ant-api03-...` |
| AWS Keys | `AKIA...`, `aws_access_key_id`, `aws_secret_access_key` |
| GitHub Tokens | `ghp_...`, `github_pat_...` |
| JWT Tokens | `eyJ...` |
| Private Keys | `-----BEGIN PRIVATE KEY-----` |
| Passwords | `password = "...", pass: "...", pwd: "..."` |
| Database URLs | `postgres://user:pass@host/db`, `mysql://...` |
| Generic Tokens | `token = "...", auth_key: "..."` |

### Entropy Analysis

Detects high-entropy strings that typically indicate:

- API keys
- Secret tokens
- Encryption keys
- Random hashes

## Risk Levels

| Level | Meaning | Example |
|-------|---------|---------|
| CRITICAL | Certain secret, immediate action required | SSH private key, `.env` file |
| HIGH | High confidence secret, should be investigated | API key, password, database URL |
| MEDIUM | Possible secret, review recommended | High-entropy string |
| LOW | Low risk, informational | Configuration file with sensitive patterns |

## Output Format

```
🚨 Found N potential security issue(s):

  1. [LEVEL] /path/to/file
     → reason 1
     Line: X: context line (if available)
     → reason 2
```

## Configuration

Security scanning behavior is controlled in `n0man.toml`:

```toml
[security]
enabled = true              # Enable/disable security scanning
scan_content = true         # Scan file content (not just patterns)
exclude_patterns = true     # Use pattern-based exclusion
sensitivity = "medium"      # low, medium, high, paranoid
fail_on_secrets = true      # Fail if secrets found (in add/sync)
interactive = true          # Prompt before blocking operations

[security.content_scan]
entropy_threshold = 4.5     # Entropy threshold (0-8)
min_secret_length = 20      # Minimum length for secret detection
max_file_size = 10485760    # Max file size to scan (10MB)
scan_binary_files = false   # Skip binary files
context_window = 50         # Lines of context for analysis

[security.pattern_config]
custom = ["*.mysecret"]     # Custom patterns to exclude

[security.allowlist]
patterns = ["*test*", "*example*"]  # Allowlist patterns
files = ["safe_config.yaml"]        # Allowlist specific files
```

## Use Cases

### Before Initial Sync

```bash
$ n0man init git@github.com:user/dotfiles.git
$ n0man add ~/.vimrc
$ n0man add ~/.config/nvim
$ n0man security scan
# Check for secrets before syncing
$ n0man sync
```

### After Adding Files

```bash
$ n0man add ~/.config/api/keys.yaml --no-security
# Added with --no-security because you know it's safe
$ n0man security scan
# Still verify no other secrets were added
```

### Audit Existing Repository

```bash
$ n0man security scan
# Periodic audit of all tracked files
```

### Scan Specific Directory

```bash
$ n0man security scan ~/.config/nvim
# Check specific directory before adding
```

## Handling False Positives

### Add to Allowlist

If a file is flagged but you know it's safe:

```toml
[security.allowlist]
files = ["safe_config.yaml", "test_keys.txt"]
patterns = ["*test*", "*example*", "*demo*"]
```

### Interactive Mode

If `interactive = true`, n0man will prompt you before blocking operations:

```bash
$ n0man add ~/.config/app
🚨 Found 1 potential security issue(s):

  1. [HIGH] .env
     → api_key pattern match

? Potential secrets found. Do you want to continue? (y/N)
> y
✅ Added 'app' (~/.config/app)
```

### Skip Scanning

For files you know are safe:

```bash
n0man add ~/.config/app --no-security
n0man sync --no-security
```

**Warning**: Only skip scanning if you're absolutely certain files are safe.

## Comparison with `add`/`sync` Scanning

| Feature | `add`/`sync` | `security scan` |
|---------|--------------|-----------------|
| Automatic | Yes | No (manual) |
| Blocks operation | Yes (if `fail_on_secrets=true`) | No (report only) |
| Interactive | Yes (if `interactive=true`) | No |
| Scope | Only new/changed files | All files or specified path |

## Notes

- Scans follow symlinks to actual file content
- Binary files are skipped by default
- Large files (>10MB by default) are skipped
- Git and backup directories are automatically excluded
- Entropy detection uses Shannon entropy calculation
- Pattern matching is case-insensitive for keywords
- Secret values are redacted in output (shown as `sk-...****...xyz`)

## Best Practices

1. **Scan before sync**: Always run before first sync of new dotfiles
2. **Review findings**: Don't blindly accept false positives
3. **Use allowlists**: For safe files with secret-like patterns
4. **Regular audits**: Periodically scan your entire dotfiles
5. **Secure secrets**: Store true secrets in password managers, not dotfiles

## Troubleshooting

### Too Many False Positives

```toml
[security]
sensitivity = "low"  # Lower sensitivity
exclude_patterns = false  # Disable pattern-based detection
```

### Missing Secrets

```toml
[security]
sensitivity = "paranoid"  # Higher sensitivity
entropy_threshold = 4.0  # Lower threshold
min_secret_length = 15  # Shorter minimum length
```

### Scan Takes Too Long

```toml
[security.content_scan]
max_file_size = 1048576  # Reduce from 10MB to 1MB
scan_binary_files = false  # Ensure binary files are skipped
```

## Related Commands

- `n0man add`: Can use `--no-security` to skip scanning
- `n0man sync`: Can use `--no-security` to skip scanning
- Configuration in `n0man.toml`: Controls scanning behavior
