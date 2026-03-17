package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/system"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Stop tracking a dotfile and restore it to its original location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		targetPath := cfg.GetTargetPath(name)
		if targetPath == "" {
			// Provide helpful error with suggestions
			tracked := cfg.GetPaths()
			if len(tracked) > 0 {
				var names []string
				for n := range tracked {
					names = append(names, n)
				}
				sort.Strings(names)

				// Check for similar name
				suggestion := ""
				for _, n := range names {
					if strings.Contains(n, name) || strings.Contains(name, n) {
						suggestion = n
						break
					}
				}

				if suggestion != "" {
					return fmt.Errorf("dotfile '%s' is not tracked\n\nDid you mean '%s'?\n\nTracked dotfiles:\n  %s\n\nOr use 'n0man add' to track a new file.", name, suggestion, strings.Join(names, "\n  "))
				}

				return fmt.Errorf("dotfile '%s' is not tracked\n\nTracked dotfiles:\n  %s\n\nOr use 'n0man add' to track a new file.", name, strings.Join(names, "\n  "))
			}

			return fmt.Errorf("dotfile '%s' is not tracked\n\nNo dotfiles are currently tracked.\n\nUse 'n0man add <path>' to track your first dotfile.", name)
		}

		homeDir, _ := os.UserHomeDir()
		if strings.HasPrefix(targetPath, "~") {
			targetPath = strings.Replace(targetPath, "~", homeDir, 1)
		}

		storePath := filepath.Join(cfg.LocalPath, name)

		if force {
			ui.PrintStep(fmt.Sprintf("Force removing '%s' from tracking", name))

			// Remove symlink if it exists
			if isLink, _ := system.IsSymlink(targetPath); isLink {
				err = os.Remove(targetPath)
				if err != nil {
					return fmt.Errorf("failed to remove symlink: %w", err)
				}
			}

			// Remove from store if it exists
			if _, err := os.Stat(storePath); err == nil {
				if err := os.RemoveAll(storePath); err != nil {
					return fmt.Errorf("failed to remove from store: %w", err)
				}
			}

			cfg.Delete(name)
		} else {
			ui.PrintStep(fmt.Sprintf("Restoring %s from %s", targetPath, storePath))

			if isLink, _ := system.IsSymlink(targetPath); isLink {
				err = os.Remove(targetPath)
				if err != nil {
					return fmt.Errorf("failed to remove symlink: %w", err)
				}
			}

			err = system.MovePath(storePath, targetPath)
			if err != nil {
				return fmt.Errorf("failed to restore: %w", err)
			}

			cfg.Delete(name)
		}

		if err := cfg.Save(cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if force {
			ui.PrintSuccess(fmt.Sprintf("Force removed '%s' from tracking", name))
		} else {
			ui.PrintSuccess(fmt.Sprintf("Removed '%s' and restored to %s", name, targetPath))
		}
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolP("force", "f", false, "Force removal without restoring file (useful if store file is missing)")
	rootCmd.AddCommand(rmCmd)
}
