# Contributing to n0man

Thank you for your interest in contributing to n0man! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites

- **Go 1.26.1 or later** - Required for building the project
- **golangci-lint** - Required for code linting
- **git** - Required for version control

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/noeltz/n0man.git
   cd n0man
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Verify the build works:
   ```bash
   make build
   ```

## Building and Testing

| Command | Description |
|---------|-------------|
| `make build` | Build the n0man binary |
| `make test` | Run all tests with race detection |
| `make lint` | Run golangci-lint |
| `make vet` | Run go vet |
| `make all` | Run full CI pipeline (test + lint + vet + build) |

## Code Standards

### Go Conventions

- Follow standard Go coding conventions
- Use meaningful variable and function names
- Add comments for exported functions following Go doc standards
- Keep functions focused and small (under 50 lines when possible)

### AI/Agent Context Comments

This project uses structured comments to help AI agents understand code intent. When adding new code:

```go
// PURPOSE: Explains WHY this code exists
// Example:
// PURPOSE: Prevent accidental commit of sensitive data

// PATTERN: Identifies design patterns in use
// Example:
// PATTERN: Facade pattern - provides simple interface to complex subsystem

// CONSTRAINT: Documents limitations or requirements
// Example:
// CONSTRAINT: Only set N0MAN_ALLOW_OUTSIDE_HOME in test environments

// SEE: References related code
// Example:
// SEE: Similar implementation in services/auth.ts

// TEST CASE: Documents expected test scenarios
// Example:
// TEST CASE: Empty file → Passed=true, Findings=0
```

### Pre-commit Checklist

Before submitting a PR, ensure:

- [ ] `make all` passes without errors
- [ ] Tests cover new functionality
- [ ] New commands have documentation in `docs/commands/`
- [ ] Configuration options are documented in `docs/references/`

## Submitting Changes

### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/my-feature`
3. **Make** your changes with appropriate tests
4. **Run** `make all` to ensure code quality
5. **Commit** with clear, conventional commit messages
6. **Push** to your fork
7. **Submit** a Pull Request with a clear description

### Commit Message Format

Use conventional commits for clear history:

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature or command
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat: add backup rollback command

Allows direct rollback to latest backup:
n0man backup rollback

Fixes #123
```

```
fix: resolve symlink target validation

Ensures symlinks point within home directory even
when target is a relative path.
```

## Reporting Issues

### Bug Reports

Use GitHub Issues and include:

1. **Clear title** describing the problem
2. **Steps to reproduce** the issue
3. **Expected behavior** vs actual behavior
4. **Environment details** (OS, Go version, n0man version)
5. **Relevant logs** or error messages

### Feature Requests

Include:

1. **Clear description** of the feature
2. **Use cases** - why is this useful?
3. **Alternative solutions** you've considered

## Project Structure

```
n0man/
├── cmd/n0man/main.go       # Entry point
├── internal/
│   ├── cmd/               # CLI commands (Cobra)
│   ├── config/            # TOML configuration
│   ├── git/               # Git operations
│   ├── security/          # Secret detection
│   ├── system/            # File operations
│   ├── backup/            # Backup system
│   └── ui/                # TUI components
├── docs/                  # Documentation
└── Makefile              # Build targets
```

## Security Considerations

When contributing code that handles:

- **File paths** - Always use `system.IsPathSafe()` for validation
- **Git URLs** - Always validate with `validateGitURL()` to prevent injection
- **Sensitive data** - Never log or expose secrets in error messages

See `docs/guides/security.md` for full security guidelines.

## License

By contributing to n0man, you agree that your contributions will be licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
