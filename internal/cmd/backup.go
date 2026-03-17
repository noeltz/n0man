package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/noeltz/n0man/internal/backup"
	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/git"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

// runStoreRecovery handles recovery when store directory is missing
func runStoreRecovery(cfg *config.Config, cfgPath string) error {
	ui.PrintWarning("Store directory is missing!")

	hasBackups := false
	backups, err := backup.ListBackups(cfg)
	if err == nil && len(backups) > 0 {
		hasBackups = true
	}

	hasRemote := cfg.RemoteURL != ""

	if !hasBackups && !hasRemote {
		ui.PrintError("No recovery options available:")
		fmt.Println("  - No backups found")
		fmt.Println("  - No remote configured")
		fmt.Println("\nConsider running 'n0man init' to reinitialize.")
		return fmt.Errorf("no recovery options available")
	}

	// Build recovery options
	var choices []string
	if hasBackups {
		choices = append(choices, fmt.Sprintf("Restore from backup (%d snapshots available)", len(backups)))
	}
	if hasRemote {
		choices = append(choices, fmt.Sprintf("Re-clone from remote (%s)", cfg.RemoteURL))
	}
	choices = append(choices, "Reinitialize store (fresh start)")
	choices = append(choices, "Cancel")

	choice, err := ui.SelectPrompt("Choose recovery method:", choices)
	if err != nil || choice < 0 || choice == len(choices)-1 {
		return fmt.Errorf("recovery cancelled")
	}

	// Execute chosen recovery
	switch {
	case hasBackups && choice == 0:
		// Restore from backup
		restoreChoice, err := ui.SelectPrompt("Choose a backup to restore:", backups)
		if err != nil || restoreChoice < 0 {
			return fmt.Errorf("recovery cancelled")
		}

		timestamp := backups[restoreChoice]
		ui.PrintStep(fmt.Sprintf("Restoring from backup %s", timestamp))

		// Create store directory
		if err := os.MkdirAll(cfg.LocalPath, 0700); err != nil {
			return fmt.Errorf("failed to create store directory: %w", err)
		}

		if err := backup.RestoreBackup(cfg, timestamp); err != nil {
			return err
		}

		ui.PrintSuccess("Store restored from backup")

	case hasRemote && ((hasBackups && choice == 1) || (!hasBackups && choice == 0)):
		// Re-clone from remote
		ui.PrintStep(fmt.Sprintf("Cloning %s → %s", cfg.RemoteURL, cfg.LocalPath))

		client := git.NewClient()
		if err := client.Clone(cfg.RemoteURL, cfg.LocalPath); err != nil {
			return fmt.Errorf("failed to clone: %w", err)
		}

		ui.PrintSuccess("Store re-cloned from remote")

	default:
		// Reinitialize
		ui.PrintStep("Initializing fresh store")

		if err := os.MkdirAll(cfg.LocalPath, 0700); err != nil {
			return fmt.Errorf("failed to create store directory: %w", err)
		}

		client := git.NewClient()
		if err := client.Init(cfg.LocalPath); err != nil {
			return fmt.Errorf("failed to init git: %w", err)
		}

		// Configure git user if not set
		cmd := exec.Command("git", "config", "user.name")
		cmd.Dir = cfg.LocalPath
		if out, _ := cmd.Output(); len(out) == 0 {
			_ = exec.Command("git", "config", "user.name", "n0man").Run()
			_ = exec.Command("git", "config", "user.email", "n0man@localhost").Run()
		}

		ui.PrintSuccess("Store reinitialized")
	}

	// Recreate symlinks after recovery
	ui.PrintStep("Recreating symlinks...")
	homeDir, _ := os.UserHomeDir()

	for name, targetPath := range cfg.GetPaths() {
		realTarget := targetPath
		if filepath.IsAbs(targetPath) {
			realTarget = targetPath
		} else if targetPath[0] == '~' {
			realTarget = filepath.Join(homeDir, targetPath[1:])
		}

		storePath := filepath.Join(cfg.LocalPath, name)

		// Check if store file exists
		if _, err := os.Stat(storePath); os.IsNotExist(err) {
			continue // Skip if store file doesn't exist
		}

		// Remove existing symlink if present
		_ = os.Remove(realTarget)

		// Create symlink
		if err := os.Symlink(storePath, realTarget); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to recreate symlink for %s: %v", name, err))
		}
	}

	ui.PrintSuccess("Symlinks recreated")
	return nil
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage dotfile backups interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// Check if store is missing
		if _, err := os.Stat(cfg.LocalPath); os.IsNotExist(err) {
			return runStoreRecovery(cfg, cfgPath)
		}

		backups, err := backup.ListBackups(cfg)
		if err != nil {
			return err
		}

		ui.PrintHeader("n0man Backup Manager")

		// 1. Choose Action
		actionChoices := []string{"[+] Create New Backup", "[↺] Restore from List", "[x] Exit"}
		action, err := ui.SelectPrompt("What would you like to do?", actionChoices)
		if err != nil || action < 0 || action == 2 {
			return nil
		}

		if action == 0 {
			// Create New
			ui.PrintStep("Creating manual backup snapshot...")
			ts, err := backup.CreateSnapshot(cfg)
			if err != nil {
				return err
			}
			if ts == "" {
				ui.PrintInfo("No dotfiles tracked, nothing to backup.")
				return nil
			}
			ui.PrintSuccess(fmt.Sprintf("Backup created: %s", ts))
			return nil
		}

		// 2. Restore from List
		if len(backups) == 0 {
			ui.PrintInfo("No backups available to restore.")
			return nil
		}

		ui.PrintSection("Restorable Backups")
		choice, err := ui.SelectPrompt("Choose a backup to restore:", backups)
		if err != nil || choice < 0 {
			return nil // Cancelled
		}

		timestamp := backups[choice]
		ui.PrintStep(fmt.Sprintf("Restoring from backup %s", timestamp))
		if err := backup.RestoreBackup(cfg, timestamp); err != nil {
			return err
		}
		ui.PrintSuccess("Backup restore complete")

		return nil
	},
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Directly create a backup snapshot (non-interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		ui.PrintStep("Creating manual backup snapshot...")
		ts, err := backup.CreateSnapshot(cfg)
		if err != nil {
			return err
		}
		if ts == "" {
			ui.PrintInfo("No dotfiles tracked, nothing to backup.")
			return nil
		}

		ui.PrintSuccess(fmt.Sprintf("Backup created: %s", ts))
		return nil
	},
}

var backupRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Directly restore the latest backup (non-interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// Check if store is missing
		if _, err := os.Stat(cfg.LocalPath); os.IsNotExist(err) {
			return runStoreRecovery(cfg, cfgPath)
		}

		backups, err := backup.ListBackups(cfg)
		if err != nil {
			return err
		}

		if len(backups) == 0 {
			return fmt.Errorf("no backups available")
		}

		latestBackup := backups[0]
		ui.PrintStep(fmt.Sprintf("Restoring from latest backup %s", latestBackup))

		if err := backup.RestoreBackup(cfg, latestBackup); err != nil {
			return err
		}

		ui.PrintSuccess("Rollback complete")
		return nil
	},
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRollbackCmd)
	rootCmd.AddCommand(backupCmd)
}
