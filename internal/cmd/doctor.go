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

// DoctorIssue represents a detected issue with potential fix action
type DoctorIssue struct {
	Category    string
	Description string
	CanFix      bool
	FixFunc     func() error
	Fixed       bool
}

// runDoctorChecks performs all health checks and returns detected issues
func runDoctorChecks(cfgPath string, cfg *config.Config, homeDir string) []DoctorIssue {
	var issues []DoctorIssue

	// 1. Store directory
	if _, err := os.Stat(cfg.LocalPath); os.IsNotExist(err) {
		issues = append(issues, DoctorIssue{
			Category:    "Store",
			Description: fmt.Sprintf("Store directory missing: %s", cfg.LocalPath),
			CanFix:      true,
			FixFunc: func() error {
				return os.MkdirAll(cfg.LocalPath, 0700)
			},
		})
	}

	// 2. Symlink integrity
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

		if _, err := os.Stat(storePath); os.IsNotExist(err) {
			issues = append(issues, DoctorIssue{
				Category:    "Symlink",
				Description: fmt.Sprintf("%s: Store file missing", name),
				CanFix:      false,
			})
			continue
		}

		info, err := os.Lstat(realTarget)
		if err != nil {
			if os.IsNotExist(err) {
				finalName := name
				finalStorePath := storePath
				finalTarget := realTarget
				issues = append(issues, DoctorIssue{
					Category:    "Symlink",
					Description: fmt.Sprintf("%s: Symlink missing", finalName),
					CanFix:      true,
					FixFunc: func() error {
						return os.Symlink(finalStorePath, finalTarget)
					},
				})
			}
			continue
		}

		if info.Mode()&os.ModeSymlink == 0 {
			issues = append(issues, DoctorIssue{
				Category:    "Symlink",
				Description: fmt.Sprintf("%s: Not a symlink", name),
				CanFix:      false,
			})
			continue
		}

		linkDest, err := os.Readlink(realTarget)
		if err != nil {
			continue
		}

		if linkDest != storePath {
			issues = append(issues, DoctorIssue{
				Category:    "Symlink",
				Description: fmt.Sprintf("%s: Wrong symlink target", name),
				CanFix:      true,
				FixFunc: func() error {
					os.Remove(realTarget)
					return os.Symlink(storePath, realTarget)
				},
			})
		}
	}

	// 3. Git health
	gitDir := filepath.Join(cfg.LocalPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		issues = append(issues, DoctorIssue{
			Category:    "Git",
			Description: "Store is not a Git repository",
			CanFix:      true,
			FixFunc: func() error {
				cmd := exec.Command("git", "init")
				cmd.Dir = cfg.LocalPath
				return cmd.Run()
			},
		})
	}

	// 4. Stale entries (files in store but not in config)
	entries, err := os.ReadDir(cfg.LocalPath)
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if name == ".git" || name == ".backups" || name == ".gitignore" {
				continue
			}
			if _, ok := cfg.GetPaths()[name]; !ok {
				finalName := name
				finalStorePath := filepath.Join(cfg.LocalPath, name)
				issues = append(issues, DoctorIssue{
					Category:    "Stale",
					Description: fmt.Sprintf("'%s' in store but NOT in config", finalName),
					CanFix:      true,
					FixFunc: func() error {
						if entry.IsDir() {
							return os.RemoveAll(finalStorePath)
						}
						return os.Remove(finalStorePath)
					},
				})
			}
		}
	}

	return issues
}

// printDoctorOutput prints the health check results
func printDoctorOutput(cfg *config.Config, issues []DoctorIssue, homeDir string) {
	ui.PrintSection("Store Directory")
	storeMissing := false
	for _, issue := range issues {
		if issue.Category == "Store" {
			ui.PrintError(issue.Description)
			storeMissing = true
		}
	}
	if !storeMissing {
		ui.PrintSuccess(cfg.LocalPath)
	}

	ui.PrintSection("Symlink Integrity")
	for name := range cfg.GetPaths() {
		hasIssue := false
		for _, issue := range issues {
			if issue.Category == "Symlink" && strings.HasPrefix(issue.Description, name+":") {
				ui.PrintError(issue.Description)
				hasIssue = true
				break
			}
		}
		if !hasIssue {
			ui.PrintSuccess(name)
		}
	}

	ui.PrintSection("Stale Entries")
	staleFound := false
	for _, issue := range issues {
		if issue.Category == "Stale" {
			ui.PrintWarning(issue.Description)
			staleFound = true
		}
	}
	// Also check for stale entries not yet in issues (edge case)
	entries, err := os.ReadDir(cfg.LocalPath)
	if err == nil {
		alreadyReported := make(map[string]bool)
		for _, issue := range issues {
			if issue.Category == "Stale" {
				alreadyReported[issue.Description] = true
			}
		}
		for _, entry := range entries {
			name := entry.Name()
			if name == ".git" || name == ".backups" || name == ".gitignore" {
				continue
			}
			desc := fmt.Sprintf("'%s' in store but NOT in config", name)
			if _, ok := cfg.GetPaths()[name]; !ok && !alreadyReported[desc] {
				ui.PrintWarning(desc)
				staleFound = true
			}
		}
	}
	if !staleFound {
		ui.PrintSuccess("No stale entries")
	}

	ui.PrintSection("Git Health")
	gitIssues := 0
	for _, issue := range issues {
		if issue.Category == "Git" {
			ui.PrintWarning(issue.Description)
			gitIssues++
		}
	}
	if gitIssues == 0 {
		gitDir := filepath.Join(cfg.LocalPath, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			gitStatus := exec.Command("git", "status", "--porcelain")
			gitStatus.Dir = cfg.LocalPath
			out, err := gitStatus.Output()
			if err != nil {
				ui.PrintError(fmt.Sprintf("git status failed: %v", err))
			} else if len(out) > 0 {
				ui.PrintWarning("Uncommitted changes detected")
			} else {
				ui.PrintSuccess("Clean working tree")
			}

			gitRemote := exec.Command("git", "remote", "-v")
			gitRemote.Dir = cfg.LocalPath
			remoteOut, err := gitRemote.Output()
			if err != nil || len(remoteOut) == 0 {
				if cfg.RemoteURL != "" {
					ui.PrintWarning("Remote URL in config but no Git remote set")
				} else {
					ui.PrintInfo("No remote configured (local-only mode)")
				}
			} else {
				ui.PrintSuccess("Remote configured")
			}
		}
	}

	ui.PrintSection("Configuration")
	if cfg.RemoteURL == "" {
		ui.PrintInfo("No remote_url set (local-only mode)")
	} else {
		ui.PrintSuccess(fmt.Sprintf("remote_url: %s", cfg.RemoteURL))
	}
	ui.PrintSuccess(fmt.Sprintf("local_path: %s", cfg.LocalPath))
	ui.PrintSuccess(fmt.Sprintf("max_backups: %d", cfg.Settings.HousekeepingMaxBackups))
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run comprehensive health checks on your dotfiles setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ui.PrintHeader("n0man doctor")

		homeDir, _ := os.UserHomeDir()
		autoFix, _ := cmd.Flags().GetBool("fix")

		// Run checks
		issues := runDoctorChecks(cfgPath, cfg, homeDir)

		// Print results
		printDoctorOutput(cfg, issues, homeDir)

		fmt.Println()
		if len(issues) == 0 {
			ui.PrintSuccess("All checks passed! Your dotfiles are healthy.")
		} else {
			ui.PrintWarning(fmt.Sprintf("Found %d issue(s). Review above for details.", len(issues)))

			// Offer interactive fix
			fixableCount := 0
			for _, issue := range issues {
				if issue.CanFix {
					fixableCount++
				}
			}

			if fixableCount > 0 {
				shouldFix := autoFix
				if !autoFix {
					fmt.Println()
					shouldFix, _ = ui.PromptConfirm(fmt.Sprintf("Would you like to fix %d issue(s) automatically?", fixableCount), true)
				}

				if shouldFix {
					fmt.Println()
					ui.PrintStep("Fixing issues...")

					fixedCount := 0
					for i := range issues {
						if issues[i].CanFix && !issues[i].Fixed {
							err := issues[i].FixFunc()
							if err != nil {
								ui.PrintError(fmt.Sprintf("Failed to fix %s: %v", issues[i].Description, err))
							} else {
								ui.PrintSuccess(fmt.Sprintf("✓ Fixed: %s", issues[i].Description))
								issues[i].Fixed = true
								fixedCount++
							}
						}
					}

					fmt.Println()
					if fixedCount > 0 {
						ui.PrintSuccess(fmt.Sprintf("Fixed %d issue(s)!", fixedCount))
					}

					// Re-run doctor to show remaining issues
					if fixedCount > 0 && len(issues) > fixedCount {
						fmt.Println()
						ui.PrintInfo("Running checks again to verify fixes...")
						fmt.Println()
						// Reload config and re-run
						cfg, err = config.Load(cfgPath)
						if err != nil {
							return fmt.Errorf("failed to reload config: %w", err)
						}
						issues = runDoctorChecks(cfgPath, cfg, homeDir)
						printDoctorOutput(cfg, issues, homeDir)
					}
				}
			}
		}

		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolP("fix", "f", false, "Automatically fix issues without prompting")
	rootCmd.AddCommand(doctorCmd)
}
