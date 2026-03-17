package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Inspect divergence between your live machine, config, and the repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.LocalPath == "" {
			return fmt.Errorf("store path not configured. Run 'n0man init'")
		}

		ui.PrintHeader("n0man status")

		hasIssues := false
		homeDir, _ := os.UserHomeDir()

		// 1. Dotfiles mapping
		ui.PrintSection("Dotfiles Mapping")
		if len(cfg.GetPaths()) == 0 {
			ui.PrintInfo("No dotfiles tracked")
		}

		for name := range cfg.GetPaths() {
			targetPath := cfg.GetTargetPath(name)
			if targetPath == "" {
				continue
			}
			realTarget := targetPath
			if strings.HasPrefix(targetPath, "~") {
				realTarget = strings.Replace(targetPath, "~", homeDir, 1)
			}
			storePath := filepath.Join(cfg.LocalPath, name)

			info, err := os.Lstat(realTarget)
			if err != nil {
				if os.IsNotExist(err) {
					ui.PrintError(fmt.Sprintf("%s: Target missing (%s)", name, targetPath))
				} else {
					ui.PrintError(fmt.Sprintf("%s: Error (%v)", name, err))
				}
				hasIssues = true
				continue
			}

			if info.Mode()&os.ModeSymlink == 0 {
				ui.PrintWarning(fmt.Sprintf("%s: Exists but NOT a symlink (%s)", name, targetPath))
				hasIssues = true
				continue
			}

			linkTarget, err := os.Readlink(realTarget)
			if err != nil {
				ui.PrintError(fmt.Sprintf("%s: Failed to read symlink: %v", name, err))
				hasIssues = true
				continue
			}

			if linkTarget != storePath {
				ui.PrintWarning(fmt.Sprintf("%s: Wrong target (expected %s, got %s)", name, storePath, linkTarget))
				hasIssues = true
				continue
			}

			if _, err := os.Stat(storePath); os.IsNotExist(err) {
				ui.PrintError(fmt.Sprintf("%s: Broken symlink — store file missing", name))
				hasIssues = true
				continue
			}

			ui.PrintSuccess(fmt.Sprintf("%s (%s)", name, targetPath))
		}

		// 2. Git status
		ui.PrintSection("Git Repository")
		if _, err := os.Stat(filepath.Join(cfg.LocalPath, ".git")); os.IsNotExist(err) {
			ui.PrintWarning("Not a git repository")
			hasIssues = true
		} else {
			gitCmd := exec.Command("git", "status", "--short")
			gitCmd.Dir = cfg.LocalPath
			out, err := gitCmd.Output()
			if err != nil {
				ui.PrintError(fmt.Sprintf("git status failed: %v", err))
				hasIssues = true
			} else if len(out) == 0 {
				ui.PrintSuccess("Clean working tree")
			} else {
				ui.PrintWarning("Uncommitted changes:")
				lines := strings.Split(strings.TrimSpace(string(out)), "\n")
				for _, line := range lines {
					fmt.Printf("       %s\n", line)
				}
				hasIssues = true
			}
		}

		fmt.Println()
		if hasIssues {
			ui.PrintWarning("Divergence or issues detected")
		} else {
			ui.PrintSuccess("All systems go — no drift detected")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
