package cmd

import (
	"fmt"
	"os"

	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/security"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Security related operations",
}

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan dotfiles for sensitive information",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfg, err := config.Load(cfgPath)
		if err != nil && err != config.ErrConfigNotFound {
			return err
		}

		var scanPath string
		if len(args) > 0 {
			scanPath = args[0]
		} else if cfg != nil && cfg.LocalPath != "" {
			scanPath = cfg.LocalPath
		} else {
			return fmt.Errorf("no path specified and local store not configured")
		}

		if cfg == nil {
			return fmt.Errorf("config not loaded. Run 'n0man init' first")
		}

		ui.PrintHeader(fmt.Sprintf("n0man security scan: %s", scanPath))
		scanner := security.NewScanner(&cfg.Security)
		report, err := scanner.ScanPath(scanPath)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		findings := flattenFindingsFull(report)
		security.PrintFindings(findings)
		if len(findings) > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	securityCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(securityCmd)
}

func flattenFindingsFull(report *security.ScanReport) []security.Finding {
	var findings []security.Finding
	for _, res := range report.Results {
		findings = append(findings, res.Findings...)
	}
	return findings
}
