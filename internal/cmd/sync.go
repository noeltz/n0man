package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/noeltz/n0man/internal/backup"
	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/git"
	"github.com/noeltz/n0man/internal/security"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/noeltz/n0man/internal/ui/conflict"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Bidirectional sync: commit local changes, pull remote changes, and push",
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

		ui.PrintHeader("n0man sync")

		// Pre-flight checks
		ui.PrintStep("Running pre-flight checks...")
		homeDir, _ := os.UserHomeDir()
		preflightResults := RunPreflightChecks(cfg, homeDir)
		PrintPreflightResults(preflightResults)

		// Check if any critical checks failed
		criticalFailed := false
		for _, r := range preflightResults {
			if !r.Passed && !r.CanFix {
				criticalFailed = true
				break
			}
		}

		if criticalFailed {
			ui.PrintError("Pre-flight checks failed. Run 'n0man doctor' for details.")
			return fmt.Errorf("pre-flight checks failed")
		}

		// Handle fixable issues
		hasIssues := false
		for _, r := range preflightResults {
			if !r.Passed {
				hasIssues = true
				break
			}
		}

		if hasIssues {
			if err := HandlePreflightFailure(preflightResults, cfgPath, cfg, homeDir); err != nil {
				return err
			}
		}

		client := git.NewClient()

		// 1. Backup
		ui.PrintStep("Creating pre-sync backup snapshot...")
		ts, err := backup.CreateSnapshot(cfg)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		if ts != "" {
			ui.PrintSuccess(fmt.Sprintf("Backup created: %s", ts))
		}

		if err := backup.CleanOldBackups(cfg); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to clean old backups: %v", err))
		}

		// 2. Check for local changes
		ui.PrintStep("Checking for local changes...")
		hasChanges, err := client.HasChanges(cfg.LocalPath)
		if err != nil {
			return fmt.Errorf("failed to check for local changes: %w", err)
		}

		if hasChanges {
			// Security scan
			noSecurity, _ := cmd.Flags().GetBool("no-security")
			if !noSecurity && cfg.Security.Enabled {
				ui.PrintStep("Running security scan...")
				scanner := security.NewScanner(&cfg.Security)
				report, err := scanner.ScanPathWithContext(GetContext(), cfg.LocalPath)
				if err != nil {
					return fmt.Errorf("security scan failed: %w", err)
				}
				if report.TotalFindings > 0 {
					security.PrintFindings(flattenFindings(report))
					return fmt.Errorf("security scan found %d issue(s). Fix them or re-run with --no-security", report.TotalFindings)
				}
				ui.PrintSuccess("Security scan passed")
			}

			ui.PrintStep("Committing local changes...")

			// Update .gitignore before adding files
			if err := updateGitignore(cfg); err != nil {
				ui.PrintWarning(fmt.Sprintf("Failed to update .gitignore: %v", err))
			}

			err = client.Add(cfg.LocalPath, ".")
			if err != nil {
				return fmt.Errorf("failed to stage changes: %w", err)
			}

			message := "chore: auto-sync dotfiles update"
			err = client.Commit(cfg.LocalPath, message)
			if err != nil {
				return fmt.Errorf("failed to commit changes: %w", err)
			}
			ui.PrintSuccess("Local changes committed")
		} else {
			ui.PrintInfo("No local changes to commit")
		}

		// 3. Remote sync
		if cfg.RemoteURL != "" {
			ui.PrintStep("Synchronizing with remote...")
			err = client.Pull(cfg.LocalPath)
			if err != nil {
				conflictStrategy, _ := cmd.Flags().GetString("conflict-strategy")

				if client.IsRebasing(cfg.LocalPath) {
					// Handle conflict based on strategy
					if conflictStrategy != "" {
						// Non-interactive mode with strategy
						var keepLocal bool
						switch conflictStrategy {
						case "keep-local":
							keepLocal = true
							ui.PrintStep("Conflict strategy: Keeping local changes...")
						case "keep-remote":
							keepLocal = false
							ui.PrintStep("Conflict strategy: Keeping remote changes...")
						case "abort":
							client.AbortRebase(cfg.LocalPath)
							return fmt.Errorf("conflict detected, sync aborted as per --conflict-strategy=abort")
						default:
							client.AbortRebase(cfg.LocalPath)
							return fmt.Errorf("invalid --conflict-strategy value: %s (use: keep-local, keep-remote, abort)", conflictStrategy)
						}

						if err := client.ResolveConflict(cfg.LocalPath, keepLocal); err != nil {
							return fmt.Errorf("failed to resolve: %w", err)
						}
						client.ContinueRebase(cfg.LocalPath)
					} else {
						// Interactive mode - use TUI
						res, pErr := conflict.PromptConflictResolution()
						if pErr != nil {
							client.AbortRebase(cfg.LocalPath)
							return fmt.Errorf("conflict resolution aborted: %w", pErr)
						}

						switch res {
						case conflict.KeepLocal:
							ui.PrintStep("Resolving: Keeping local changes...")
							if err := client.ResolveConflict(cfg.LocalPath, true); err != nil {
								return fmt.Errorf("failed to resolve: %w", err)
							}
							client.ContinueRebase(cfg.LocalPath)
						case conflict.KeepRemote:
							ui.PrintStep("Resolving: Keeping remote changes...")
							if err := client.ResolveConflict(cfg.LocalPath, false); err != nil {
								return fmt.Errorf("failed to resolve: %w", err)
							}
							client.ContinueRebase(cfg.LocalPath)
						case conflict.AbortAndManual:
							client.AbortRebase(cfg.LocalPath)
							return fmt.Errorf("sync aborted. Please resolve conflicts manually in %s", cfg.LocalPath)
						}
					}
				} else {
					ui.PrintError("Pull failed — possible connection or authentication issue")
					return fmt.Errorf("failed to pull: %w", err)
				}
			}

			err = client.Push(cfg.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to push: %w", err)
			}
			ui.PrintSuccess("Remote synchronization complete")
		} else {
			ui.PrintInfo("No remote URL configured — skipping pull/push")
		}

		fmt.Println()
		ui.PrintSuccess("Sync completed successfully!")
		return nil
	},
}

func updateGitignore(cfg *config.Config) error {
	gitignorePath := filepath.Join(cfg.LocalPath, ".gitignore")
	var lines []string
	lines = append(lines, "# Generated by n0man", ".backups/", "n0man.toml")

	for name, patterns := range cfg.GetIgnores() {
		// Only add ignore patterns for directories, not files
		storePath := filepath.Join(cfg.LocalPath, name)
		info, err := os.Stat(storePath)
		if err != nil {
			// If path doesn't exist, skip it
			continue
		}

		// Only add ignore patterns for directories
		if !info.IsDir() {
			continue
		}

		for _, p := range patterns {
			// In .gitignore, we want forward slashes.
			// name is the top-level path in the store.
			line := name + "/" + p
			lines = append(lines, line)
		}
	}

	content := ""
	if len(lines) > 0 {
		content = fmt.Sprintf("%s\n", strings.Join(lines, "\n"))
	}

	return os.WriteFile(gitignorePath, []byte(content), 0644)
}

func init() {
	syncCmd.Flags().Bool("no-security", false, "Skip security scanning before commit")
	syncCmd.Flags().String("conflict-strategy", "", "Conflict resolution strategy for non-interactive mode (keep-local, keep-remote, abort)")
	rootCmd.AddCommand(syncCmd)
}
