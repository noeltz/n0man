# `n0man completion`

Generate shell autocompletion scripts for n0man.

## Usage

```bash
n0man completion <shell>
```

## Arguments

| Shell | Description |
|-------|-------------|
| `bash` | Bash shell completion |
| `zsh` | Zsh shell completion |
| `fish` | Fish shell completion |
| `powershell` | PowerShell completion |

## Description

The `completion` command generates shell autocompletion scripts that provide:

- Command autocompletion
- Flag autocompletion
- Argument suggestions
- Subcommand completion

This improves the command-line experience by reducing typing and preventing errors.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Completion script generated successfully |
| 1 | Invalid shell specified |

## Examples

### Bash

```bash
# Generate and source immediately
source <(n0man completion bash)

# Or save and source from .bashrc
n0man completion bash > ~/.n0man-completion.bash
echo "source ~/.n0man-completion.bash" >> ~/.bashrc
source ~/.bashrc
```

### Zsh

```bash
# Generate and source immediately
source <(n0man completion zsh)

# Or save to completions directory
n0man completion zsh > "${fpath[1]}/_n0man"
autoload -U compinit && compinit
```

### Fish

```bash
# Generate and save
n0man completion fish > ~/.config/fish/completions/n0man.fish
```

### PowerShell

```powershell
# Generate and profile
n0man completion powershell | Out-String | Invoke-Expression

# Or save to profile
n0man completion powershell >> $PROFILE
```

## Autocompletion Features

### Command Completion

```bash
$ n0man <TAB>
add         backup      completion  doctor      help        init        list        rm          security    self-update status      sync        version
```

### Flag Completion

```bash
$ n0man sync --<TAB>
--conflict-strategy    --help                 --no-security
```

### Subcommand Completion

```bash
$ n0man backup <TAB>
create    rollback
```

### Argument Completion

```bash
$ n0man rm <TAB>
.bashrc    .vimrc    nvim    config
```

## Installation by Shell

### Bash

**Prerequisites:** Bash 4.2+, `bash-completion` package.

**Method 1: System-wide**
```bash
sudo n0man completion bash > /etc/bash_completion.d/n0man
```

**Method 2: User-specific**
```bash
n0man completion bash > ~/.local/share/bash-completion/completions/n0man
```

**Method 3: Manual**
```bash
echo "source <(n0man completion bash)" >> ~/.bashrc
source ~/.bashrc
```

### Zsh

**Prerequisites:** Zsh 5.2+

**Method 1: Using fpath**
```bash
echo "fpath=(~/.zsh/completion $fpath)" >> ~/.zshrc
n0man completion zsh > ~/.zsh/completion/_n0man
autoload -Uz compinit && compinit
```

**Method 2: Using oh-my-zsh**
```bash
n0man completion zsh > ~/.oh-my-zsh/completions/_n0man
```

### Fish

**Prerequisites:** Fish 3.0+

**Method 1: Using completions directory**
```bash
n0man completion fish > ~/.config/fish/completions/n0man.fish
```

**Method 2: Using fish_add_path**
```bash
n0man completion fish | source
```

### PowerShell

**Prerequisites:** PowerShell 5.0+

**Method 1: Current session**
```powershell
n0man completion powershell | Out-String | Invoke-Expression
```

**Method 2: All sessions**
```powershell
n0man completion powershell >> $PROFILE
. $PROFILE
```

## Troubleshooting

### Bash: Completion Not Working

**Problem:** Tab completion doesn't work.

**Solution:**
```bash
# Ensure bash-completion is installed
# Debian/Ubuntu
sudo apt install bash-completion

# RHEL/CentOS
sudo yum install bash-completion

# Source completion
source /etc/bash_completion
```

### Zsh: Completion Not Working

**Problem:** Tab completion doesn't work.

**Solution:**
```bash
# Ensure compinit is loaded
echo "autoload -Uz compinit && compinit" >> ~/.zshrc
source ~/.zshrc

# Check fpath
echo $fpath
# Should include ~/.zsh/completion
```

### Fish: Completion Not Working

**Problem:** Tab completion doesn't work.

**Solution:**
```bash
# Ensure completions directory exists
mkdir -p ~/.config/fish/completions

# Regenerate completion
n0man completion fish > ~/.config/fish/completions/n0man.fish

# Restart fish or run:
functions -e n0man
source ~/.config/fish/completions/n0man.fish
```

### PowerShell: Script Execution Policy

**Problem:** Script execution blocked.

**Solution:**
```powershell
# Check current policy
Get-ExecutionPolicy

# Set to RemoteSigned (requires admin)
Set-ExecutionPolicy RemoteSigned

# Or bypass for current session
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
```

### Completion Slow

**Problem:** Tab completion is slow.

**Solution:**
```bash
# Bash: Ensure completion is cached
hash -r

# Zsh: Rebuild completion cache
rm ~/.zcompdump*
autoload -Uz compinit && compinit
```

## Customization

### Bash: Custom Completion

Add custom completions after loading n0man completion:

```bash
# ~/.bashrc
source <(n0man completion bash)

# Custom completion for specific paths
complete -F _n0man -o default -o bashdefault n0man
```

### Zsh: Custom Completion

```zsh
# ~/.zshrc
autoload -Uz compinit && compinit

# Custom style
zstyle ':completion:*:n0man:*' menu select
```

## Uninstall

### Bash

```bash
sudo rm /etc/bash_completion.d/n0man
# Or
rm ~/.local/share/bash-completion/completions/n0man
```

### Zsh

```bash
rm ~/.zsh/completion/_n0man
# Or
rm ~/.oh-my-zsh/completions/_n0man
```

### Fish

```bash
rm ~/.config/fish/completions/n0man.fish
```

### PowerShell

```powershell
# Remove from profile
# Edit $PROFILE and remove the completion line
```

## See Also

- [Getting Started](../guides/getting-started.md) - Installation guide
- [Commands](../commands/) - All n0man commands
