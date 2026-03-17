package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/noeltz/n0man/internal/config"
	"github.com/noeltz/n0man/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked dotfiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.DefaultConfigPath()
		if err != nil {
			return err
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.GetPaths()) == 0 {
			ui.PrintInfo("No dotfiles tracked. Use 'n0man add <path>' to start.")
			return nil
		}

		ui.PrintHeader("n0man list")

		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.Secondary).Width(20)
		pathStyle := lipgloss.NewStyle().Foreground(ui.White).Width(35)
		ignoreStyle := lipgloss.NewStyle().Foreground(ui.Muted)

		// Header
		header := lipgloss.JoinHorizontal(lipgloss.Top,
			nameStyle.Copy().Foreground(ui.Primary).Render("NAME"),
			pathStyle.Copy().Foreground(ui.Primary).Render("TARGET PATH"),
			ignoreStyle.Copy().Foreground(ui.Primary).Render("IGNORES"),
		)
		fmt.Println("  " + header)
		fmt.Println("  " + ui.Divider.Render())

		for name, targetPath := range cfg.GetPaths() {
			ignores := strings.Join(cfg.GetIgnores()[name], ", ")
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				nameStyle.Render(name),
				pathStyle.Render(targetPath),
				ignoreStyle.Render(ignores),
			)
			fmt.Println("  " + row)
		}

		fmt.Printf("\n  %s %s\n", ui.MutedStyle.Render("Store:"), cfg.LocalPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
