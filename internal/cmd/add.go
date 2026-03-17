package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/security"
	"github.com/noeltz/n0man/internal/system"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path> [name]",
	Short: "Add a file or directory to n0man and replace it with a symlink",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]

		// Expand tilde to home directory
		if strings.HasPrefix(targetPath, "~") {
			homeDir, _ := os.UserHomeDir()
			targetPath = strings.Replace(targetPath, "~", homeDir, 1)
		}

		// Convert relative paths to absolute
		if !filepath.IsAbs(targetPath) {
			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				return fmt.Errorf("could not resolve absolute path: %w", err)
			}
			targetPath = absPath
		}

		// Validate path safety to prevent path traversal attacks
		if err := system.IsPathSafe(targetPath); err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		absTargetPath, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("could not resolve absolute path: %w", err)
		}

		// Re-validate after resolving to absolute path
		if err := system.IsPathSafe(absTargetPath); err != nil {
			return fmt.Errorf("invalid absolute path: %w", err)
		}

		if _, err := os.Stat(absTargetPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absTargetPath)
		}

		homeDir, _ := os.UserHomeDir()
		prettyPath := absTargetPath
		if strings.HasPrefix(absTargetPath, homeDir) {
			prettyPath = strings.Replace(absTargetPath, homeDir, "~", 1)
		}

		name := filepath.Base(absTargetPath)
		if len(args) == 2 {
			name = args[1]
		} else {
			// Smart default naming: if we're in ~/.config/something/file, try to use something/file
			configPrefix := filepath.Join(homeDir, ".config")
			if strings.HasPrefix(absTargetPath, configPrefix) {
				rel, err := filepath.Rel(configPrefix, absTargetPath)
				if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, "..") {
					name = rel
				}
			}
		}

		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config (did you run 'n0man init'?): %w", err)
		}

		if cfg.LocalPath == "" {
			return fmt.Errorf("store path not configured. Run 'n0man init' first")
		}

		if existingPath, ok := cfg.GetPaths()[name]; ok {
			if existingPath == prettyPath {
				ui.PrintInfo(fmt.Sprintf("'%s' is already tracked as '%s'", prettyPath, name))
				return nil
			}
			return fmt.Errorf("naming collision: name '%s' is already taken by '%s'.\nPlease provide a different name: n0man add <path> <custom-name>", name, existingPath)
		}

		// Security Scanning
		noSecurity, _ := cmd.Flags().GetBool("no-security")
		if !noSecurity && cfg.Security.Enabled {
			scanner := security.NewScanner(&cfg.Security)
			report, err := scanner.ScanPath(absTargetPath)
			if err != nil {
				return fmt.Errorf("security scan failed: %w", err)
			}

			if report.TotalFindings > 0 {
				security.PrintFindings(flattenFindings(report))
				if cfg.Security.FailOnSecrets {
					if cfg.Security.Interactive {
						confirm, _ := ui.PromptConfirm("Potential secrets found. Do you want to continue?", false)
						if !confirm {
							return fmt.Errorf("aborted due to security findings")
						}
					} else {
						return fmt.Errorf("security scan failed with %d findings", report.TotalFindings)
					}
				}
			}
		}

		ignores, _ := cmd.Flags().GetStringSlice("ignore")
		storePath := filepath.Join(cfg.LocalPath, name)

		ui.PrintStep(fmt.Sprintf("Moving %s → %s", absTargetPath, storePath))
		err = system.MovePath(absTargetPath, storePath)
		if err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}

		ui.PrintStep(fmt.Sprintf("Linking %s → %s", storePath, absTargetPath))
		err = system.CreateSymlink(storePath, absTargetPath)
		if err != nil {
			_ = system.MovePath(storePath, absTargetPath)
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		cfg.SetPath(name, prettyPath)
		if len(ignores) > 0 {
			cfg.SetIgnores(name, ignores)
		}

		if err := cfg.Save(cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// Verify that the config was actually saved
		verifiedCfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to verify config save: %w", err)
		}

		if verifiedPath, ok := verifiedCfg.GetPaths()[name]; !ok || verifiedPath != prettyPath {
			return fmt.Errorf("config verification failed: entry not saved correctly")
		}

		ui.PrintSuccess(fmt.Sprintf("Added '%s' (%s)", name, prettyPath))
		return nil
	},
}

func init() {
	addCmd.Flags().StringSliceP("ignore", "i", nil, "Pattern(s) to ignore during sync")
	addCmd.Flags().Bool("no-security", false, "Skip security scanning for this add")
	rootCmd.AddCommand(addCmd)
}

func flattenFindings(report *security.ScanReport) []security.Finding {
	var findings []security.Finding
	for _, res := range report.Results {
		findings = append(findings, res.Findings...)
	}
	return findings
}
