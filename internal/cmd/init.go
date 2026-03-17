package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/git"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

// configureGitUser configures git user.name and user.email if not already set
func configureGitUser(repoPath string, isLocalOnly bool) error {
	// Check if git user.name is already configured (check both global and local)
	checkCmd := exec.Command("git", "config", "--global", "user.name")
	if out, err := checkCmd.Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return nil // Already configured globally
	}

	// Check local config as fallback
	checkCmd = exec.Command("git", "config", "user.name")
	checkCmd.Dir = repoPath
	if out, err := checkCmd.Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return nil // Already configured locally
	}

	if isLocalOnly {
		// Local-only mode: auto-configure with generic identity
		ui.PrintStep("Configuring Git for local use...")
		if err := exec.Command("git", "config", "user.name", "n0man").Run(); err != nil {
			return err
		}
		if err := exec.Command("git", "config", "user.email", "n0man@localhost").Run(); err != nil {
			return err
		}
		return nil
	}

	// Remote mode: return error to trigger prompt
	return fmt.Errorf("Git user not configured")
}

// promptGitSetup interactively sets up git user.name and user.email
func promptGitSetup() error {
	ui.PrintStep("Git user not configured")

	fmt.Println()
	name, err := ui.InputPrompt("Enter your name (for Git commits)")
	if err != nil || name == "" {
		return fmt.Errorf("git setup cancelled")
	}

	email, err := ui.InputPrompt("Enter your email (for Git commits)")
	if err != nil || email == "" {
		return fmt.Errorf("git setup cancelled")
	}

	// Configure git globally
	if err := exec.Command("git", "config", "--global", "user.name", name).Run(); err != nil {
		return fmt.Errorf("failed to set user.name: %w", err)
	}
	if err := exec.Command("git", "config", "--global", "user.email", email).Run(); err != nil {
		return fmt.Errorf("failed to set user.email: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Git configured: %s <%s>", name, email))
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init [remote_url]",
	Short: "Initialize n0man and optionally clone a remote dotfiles repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			if err == config.ErrConfigNotFound {
				defaultCfg := config.DefaultConfig()
				cfg = &defaultCfg

				storePath, err := config.DefaultStorePath()
				if err != nil {
					return err
				}
				cfg.LocalPath = storePath
			} else {
				return err
			}
		}

		ui.PrintHeader("n0man init")
		client := git.NewClient()

		isLocalOnly := len(args) == 0

		// Configure git user BEFORE operations that need it
		// This ensures git is ready for both clone and initial commit
		if err := configureGitUser(cfg.LocalPath, isLocalOnly); err != nil {
			// For remote mode, prompt for git setup instead of failing
			if !isLocalOnly {
				if promptErr := promptGitSetup(); promptErr != nil {
					return fmt.Errorf("git setup required: %w", promptErr)
				}
			} else {
				return err
			}
		}

		if len(args) == 1 {
			remoteURL := args[0]
			ui.PrintStep(fmt.Sprintf("Cloning %s → %s", remoteURL, cfg.LocalPath))
			err = client.Clone(remoteURL, cfg.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to clone: %w", err)
			}
			cfg.RemoteURL = remoteURL
		} else {
			ui.PrintStep(fmt.Sprintf("Initializing local repository at %s", cfg.LocalPath))
			err = client.Init(cfg.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to init: %w", err)
			}
		}

		err = cfg.Save(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.PrintSuccess("n0man initialized successfully")

		if isLocalOnly {
			fmt.Println()
			fmt.Println("  Local-only mode active. Your dotfiles are tracked but not backed up remotely.")
			fmt.Println("  Add a remote later with: n0man init git@github.com:user/dotfiles.git")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
