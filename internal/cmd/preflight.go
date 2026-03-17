package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/ui"
)

// PreflightCheckResult represents the result of a single pre-flight check
type PreflightCheckResult struct {
	Name    string
	Passed  bool
	Message string
	CanFix  bool
	FixFunc func() error
}

// RunPreflightChecks runs all pre-flight checks and returns results
func RunPreflightChecks(cfg *config.Config, homeDir string) []PreflightCheckResult {
	var results []PreflightCheckResult

	// Check 1: Config loaded
	if cfg == nil {
		results = append(results, PreflightCheckResult{
			Name:    "Config",
			Passed:  false,
			Message: "Configuration not loaded",
			CanFix:  false,
		})
		return results // Can't continue without config
	}
	results = append(results, PreflightCheckResult{
		Name:    "Config",
		Passed:  true,
		Message: "Configuration loaded",
	})

	// Check 2: Store directory exists
	if _, err := os.Stat(cfg.LocalPath); os.IsNotExist(err) {
		results = append(results, PreflightCheckResult{
			Name:    "Store",
			Passed:  false,
			Message: "Store directory missing",
			CanFix:  true,
			FixFunc: func() error {
				return os.MkdirAll(cfg.LocalPath, 0700)
			},
		})
	} else {
		results = append(results, PreflightCheckResult{
			Name:    "Store",
			Passed:  true,
			Message: "Store directory exists",
		})
	}

	// Check 3: Git repository exists
	if cfg.LocalPath != "" {
		gitDir := filepath.Join(cfg.LocalPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			results = append(results, PreflightCheckResult{
				Name:    "Git",
				Passed:  false,
				Message: "Git repository missing",
				CanFix:  true,
				FixFunc: func() error {
					cmd := exec.Command("git", "init")
					cmd.Dir = cfg.LocalPath
					return cmd.Run()
				},
			})
		} else {
			results = append(results, PreflightCheckResult{
				Name:    "Git",
				Passed:  true,
				Message: "Git repository exists",
			})
		}
	}

	// Check 4: Broken symlinks
	if cfg.LocalPath != "" {
		brokenCount := 0
		brokenNames := []string{}
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

			// Check if store file exists
			if _, err := os.Stat(storePath); os.IsNotExist(err) {
				brokenCount++
				brokenNames = append(brokenNames, name)
				continue
			}

			// Check if symlink exists
			if info, err := os.Lstat(realTarget); err != nil || info.Mode()&os.ModeSymlink == 0 {
				brokenCount++
				brokenNames = append(brokenNames, name)
			}
		}

		if brokenCount > 0 {
			results = append(results, PreflightCheckResult{
				Name:    "Symlinks",
				Passed:  false,
				Message: fmt.Sprintf("%d broken symlink(s)", brokenCount),
				CanFix:  true,
				FixFunc: func() error {
					for _, name := range brokenNames {
						targetPath := cfg.GetTargetPath(name)
						if targetPath == "" {
							continue
						}
						realTarget := targetPath
						if strings.HasPrefix(targetPath, "~") {
							realTarget = strings.Replace(targetPath, "~", homeDir, 1)
						}
						storePath := filepath.Join(cfg.LocalPath, name)

						// Remove existing file/symlink
						os.Remove(realTarget)

						// Recreate symlink
						if err := os.Symlink(storePath, realTarget); err != nil {
							return err
						}
					}
					return nil
				},
			})
		} else {
			results = append(results, PreflightCheckResult{
				Name:    "Symlinks",
				Passed:  true,
				Message: "All symlinks valid",
			})
		}
	}

	return results
}

// PrintPreflightResults prints the pre-flight check results
func PrintPreflightResults(results []PreflightCheckResult) {
	allPassed := true
	for _, r := range results {
		if r.Passed {
			ui.PrintSuccess(r.Message)
		} else {
			ui.PrintWarning(r.Message)
			allPassed = false
		}
	}

	if !allPassed {
		fmt.Println()
	}
}

// HandlePreflightFailure handles pre-flight check failures interactively
func HandlePreflightFailure(results []PreflightCheckResult, cfgPath string, cfg *config.Config, homeDir string) error {
	fixableCount := 0
	for _, r := range results {
		if r.CanFix {
			fixableCount++
		}
	}

	if fixableCount == 0 {
		return fmt.Errorf("pre-flight checks failed with unfixable issues")
	}

	shouldFix, _ := ui.PromptConfirm(fmt.Sprintf("Fix %d issue(s) before continuing?", fixableCount), true)
	if !shouldFix {
		return fmt.Errorf("pre-flight checks failed. Run 'n0man doctor' for more details")
	}

	fmt.Println()
	ui.PrintStep("Fixing issues...")

	fixedCount := 0
	for i := range results {
		if results[i].CanFix && !results[i].Passed {
			err := results[i].FixFunc()
			if err != nil {
				ui.PrintError(fmt.Sprintf("Failed to fix %s: %v", results[i].Name, err))
			} else {
				ui.PrintSuccess(fmt.Sprintf("✓ Fixed: %s", results[i].Name))
				results[i].Passed = true
				fixedCount++
			}
		}
	}

	if fixedCount > 0 {
		fmt.Println()
		ui.PrintSuccess(fmt.Sprintf("Fixed %d issue(s)!", fixedCount))

		// Re-run checks to verify
		fmt.Println()
		ui.PrintInfo("Verifying fixes...")
		fmt.Println()
		newResults := RunPreflightChecks(cfg, homeDir)
		PrintPreflightResults(newResults)

		// Check if all pass now
		for _, r := range newResults {
			if !r.Passed {
				return fmt.Errorf("some issues remain. Run 'n0man doctor' for details")
			}
		}
	}

	return nil
}
