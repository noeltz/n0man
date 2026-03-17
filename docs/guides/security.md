# Security Scanning

n0man includes comprehensive security scanning to prevent committing sensitive information like API keys, passwords, and secrets to your dotfiles repository.

## Overview

The security scanner uses multiple detection methods:

1. **Pattern-Based Detection**: Identifies high-risk file patterns (`.env`, `.pem`, SSH keys)
2. **Content Scanning**: Detects secrets in file content using regex patterns
3. **Entropy Analysis**: Detects high-entropy strings that indicate random secrets

## How It Works

### Detection Pipeline

```
File → Pattern Check → Content Scan → Entropy Analysis → Findings
```

1. **Pattern Check**: Is this a known high-risk file type?
2. **Content Scan**: Does it contain known secret patterns?
3. **Entropy Analysis**: Are there high-entropy strings?

### Risk Levels

| Level | Meaning | Example |
|-------|---------|---------|
| CRITICAL | Certain secret, immediate action required | SSH private key |
| HIGH | High confidence secret, should be investigated | API key, password |
| MEDIUM | Possible secret, review recommended | High-entropy string |
| LOW | Low risk, informational | Configuration with sensitive patterns |
| NONE | No issues detected |

## Detection Methods

### 1. Pattern-Based Detection

Checks file names and paths against high-risk patterns.

#### Critical Risk Patterns

- `.env`, `.env.*`, `*.env` - Environment files
- `*.key`, `*.pem`, `*.p12`, `*.pfx`, `*.jks` - Key/certificate files
- `.ssh/id_*`, `.ssh/*_rsa` - SSH private keys
- `.aws/credentials` - AWS credentials
- `.docker/config.json` - Docker authentication
- `.kube/config` - Kubernetes config

#### High Risk Patterns

- `*_history` - Shell history files
- `.git-credentials` - Git credentials storage
- `.netrc` - FTP credentials
- `*.db`, `*.sqlite` - Database files
- `.gnupg/secring.gpg` - GPG secret keyring

#### Medium Risk Patterns

- `.gitconfig` - Git configuration
- `config.yml`, `config.yaml` - Configuration files
- `*.log` - Log files
- `logs/*`, `tmp/*`, `.cache/*` - Temporary directories

#### Low Risk Patterns

- `.DS_Store` - macOS metadata
- `node_modules/*`, `vendor/*` - Dependency directories
- `__pycache__/*`, `*.pyc` - Python cache
- `dist/*`, `build/*` - Build artifacts

### 2. Content Scanning

Uses regex patterns to detect secrets in file content.

#### API Keys

- OpenAI: `sk-...`, `sk-proj-...`
- Anthropic: `sk-ant-api03-...`
- Generic: `api_key = "..."`, `sk-[a-zA-Z0-9_-]{20,}`

#### Cloud Keys

- AWS: `AKIA[0-9A-Z]{16}`, `aws_access_key_id`
- GitHub: `ghp_...`, `github_pat_...`

#### Tokens

- JWT: `eyJ...` (Base64 encoded JWT)
- Generic tokens: `token = "..."`, `auth_key: "..."`

#### Passwords

- `password = "..."`, `pass: "..."`, `pwd: "..."`
- `database: "postgres://user:pass@host/db"`

#### Private Keys

- `-----BEGIN PRIVATE KEY-----`
- `-----BEGIN RSA PRIVATE KEY-----`
- `-----BEGIN OPENSSH PRIVATE KEY-----`

### 3. Entropy Analysis

Detects high-entropy strings that typically indicate secrets.

#### How It Works

Shannon entropy measures randomness in a string:

```
High entropy (4.5+) → Likely a secret
Medium entropy (3.5-4.5) → Possible secret
Low entropy (<3.5) → Likely normal text
```

#### Examples

| String | Entropy | Verdict |
|--------|---------|---------|
| `hello world` | 2.85 | Normal |
| `sk-1234567890abcdef` | 4.2 | Likely secret |
| `gV8kQ9mP2xL5nR7sT1uW3yA6zC4` | 5.8 | Secret |

#### Factors Considered

- String length (must be 20+ characters)
- Character variety (letters, numbers, symbols)
- Secret prefixes (`sk-`, `pk-`, `api-`, etc.)
- Structural patterns (URLs, paths)
- Natural language detection

## Configuration

### Enable/Disable Scanning

```toml
[security]
enabled = true              # Enable/disable all scanning
scan_content = true         # Scan file content
exclude_patterns = true     # Use pattern-based detection
```

### Sensitivity Levels

```toml
[security]
sensitivity = "medium"     # low, medium, high, paranoid
```

| Level | Entropy Threshold | False Positives | False Negatives |
|-------|------------------|-----------------|-----------------|
| low | 5.5 | Few | Many |
| medium | 4.5 | Balanced | Balanced |
| high | 3.5 | Many | Few |
| paranoid | 3.0 | Very many | Very few |

### Fail Behavior

```toml
[security]
fail_on_secrets = true     # Block operations on secrets
interactive = true         # Prompt before blocking
```

When `fail_on_secrets = true`:
- Operations (`add`, `sync`) fail if secrets detected
- If `interactive = true`: Prompts to continue
- If `interactive = false`: Fails immediately

### Advanced Settings

```toml
[security.content_scan]
entropy_threshold = 4.5        # Entropy threshold
min_secret_length = 20         # Minimum string length
max_file_size = 10485760      # Max file size (10MB)
scan_binary_files = false       # Skip binary files
context_window = 50            # Lines of context
```

### Custom Patterns

```toml
[security.pattern_config]
custom = ["*.internal_keys", "secrets/*"]
```

Add custom patterns to flag as high-risk.

### Allowlist

```toml
[security.allowlist]
patterns = ["*test*", "*example*", "*demo*"]
files = ["safe_config.yaml", "sample_keys.txt"]
```

Files and patterns that are never flagged.

## Running Scans

### Scan All Dotfiles

```bash
n0man security scan
```

Scans the entire n0man store.

### Scan Specific Path

```bash
n0man security scan ~/.config/nvim
```

Scans a specific directory or file.

### Scan Before Sync

```bash
n0man security scan
# Review findings
n0man sync
```

### Scan in `add` Command

By default, `add` runs security scan:

```bash
$ n0man add ~/.config/api
🚨 Found 1 potential security issue(s):
  1. [HIGH] .env
     → api_key pattern match

Error: security scan failed with 1 findings
```

Skip scanning if you know it's safe:

```bash
n0man add ~/.config/api --no-security
```

### Scan in `sync` Command

By default, `sync` runs security scan before committing:

```bash
$ n0man sync
  Running security scan...
🚨 Found 2 potential security issue(s):
  1. [HIGH] config.yaml
     → password pattern match
  2. [CRITICAL] private.pem
     → private_key pattern match

Error: security scan found 2 issue(s). Fix them or re-run with --no-security
```

## Handling Findings

### Review Findings

When secrets are found, review them carefully:

```
🚨 Found 3 potential security issue(s):

  1. [CRITICAL] /path/to/.env
     → api_key pattern match

  2. [HIGH] /path/to/config.yaml
     → password pattern match
     Line: 12: password: "s3cr3t1234"

  3. [HIGH] /path/to/private.pem
     → private_key pattern match
```

### Fix True Positives

For actual secrets:

1. **Remove the secret** from the file
2. **Use environment variables** or `.env.local`
3. **Add to .gitignore**: `echo ".env" >> .gitignore`
4. **Revoke and rotate** compromised secrets

### Handle False Positives

For false alarms:

#### Option 1: Add to Allowlist

```toml
[security.allowlist]
patterns = ["*test*", "*example*"]
files = ["demo_config.yaml"]
```

#### Option 2: Lower Sensitivity

```toml
[security]
sensitivity = "low"
```

#### Option 3: Disable Specific Scans

```toml
[security]
scan_content = false        # Disable content scanning
exclude_patterns = false     # Disable pattern checking
```

#### Option 4: Skip Scanning (Not Recommended)

```bash
n0man add ~/.config/app --no-security
n0man sync --no-security
```

**Warning**: Only skip if you're absolutely certain files are safe.

## Output Format

### No Issues

```
✅ No security issues found.
```

### Issues Found

```
🚨 Found 3 potential security issue(s):

  1. [CRITICAL] /path/to/.env
     → api_key pattern match

  2. [HIGH] /path/to/config.yaml
     → password pattern match
     Line: 12: password: "s3cr3t1234"

  3. [HIGH] /path/to/private.pem
     → private_key pattern match
     Line: 1: -----BEGIN PRIVATE KEY-----
```

## Use Cases

### Before Initial Commit

```bash
n0man init git@github.com:user/dotfiles.git
n0man add ~/.vimrc
n0man add ~/.config/nvim
n0man security scan
# Verify no secrets
n0man sync
```

### After Adding Files

```bash
n0man add ~/.config/api --no-security
# Added with --no-security because you know it's safe
n0man security scan
# Verify no OTHER files have secrets
```

### Regular Audits

```bash
# Monthly security audit
n0man security scan
# Review and remediate findings
```

### CI/CD Integration

```bash
# In CI/CD pipeline
n0man security scan || exit 1
```

## Best Practices

1. **Always scan before sync**: Never sync without scanning
2. **Use environment variables**: Store secrets in `.env.local`, not dotfiles
3. **Rotate compromised secrets**: If a secret was committed, revoke it
4. **Regular audits**: Periodically scan your entire repository
5. **Review findings**: Don't blindly accept false positives
6. **Use allowlists**: For safe files with secret-like patterns
7. **Keep secrets separate**: Use password managers for true secrets

## Common False Positives

### Test Files

```toml
[security.allowlist]
patterns = ["*test*", "*example*", "*demo*", "*sample*"]
```

### UUIDs and Identifiers

High-entropy strings that aren't secrets:
- UUIDs (`550e8400-e29b-41d4-a716-446655440000`)
- Hashes (SHA256, MD5)
- Random identifiers

### Configuration Files

Safe config files with secret-like patterns:
```toml
[security.allowlist]
files = [
  "demo_config.yaml",
  "sample_keys.txt",
  "test_secrets.env"
]
```

## Limitations

### What It Detects

- API keys, passwords, tokens
- Private keys, certificates
- High-entropy secrets
- Known secret patterns

### What It Misses

- Custom encryption schemes
- Base64-encoded secrets (if not recognized)
- Secrets split across multiple files
- Secrets in comments or documentation
- Social engineering (user error)

### What It Flags Incorrectly

- High-entropy UUIDs
- Hash values (SHA256, MD5)
- Random identifiers
- Configuration examples

## Troubleshooting

### Too Many False Positives

```toml
[security]
sensitivity = "low"
entropy_threshold = 5.5
```

### Missing Secrets

```toml
[security]
sensitivity = "paranoid"
entropy_threshold = 3.0
min_secret_length = 15
```

### Scan Takes Too Long

```toml
[security.content_scan]
max_file_size = 1048576  # Reduce from 10MB
scan_binary_files = false
```

### Specific File Always Flagged

```toml
[security.allowlist]
files = ["that_file.yaml"]
```

## Related Documentation

- [Configuration](configuration.md): Security settings
- [Commands](commands/): How commands use scanning
- [Commands: Security](commands/security.md): `n0man security scan`
