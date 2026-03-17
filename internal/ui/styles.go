package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	Primary   = lipgloss.Color("#7C3AED") // violet
	Secondary = lipgloss.Color("#06B6D4") // cyan
	Success   = lipgloss.Color("#10B981") // emerald
	Warning   = lipgloss.Color("#F59E0B") // amber
	Danger    = lipgloss.Color("#EF4444") // red
	Muted     = lipgloss.Color("#6B7280") // gray
	White     = lipgloss.Color("#F9FAFB")
	Subtle    = lipgloss.Color("#9CA3AF")
)

// Styles
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		PaddingBottom(1)

	Subtitle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	PrimaryStyle = lipgloss.NewStyle().
			Foreground(Primary)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	DangerStyle = lipgloss.NewStyle().
			Foreground(Danger)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	Bold = lipgloss.NewStyle().
		Bold(true).
		Foreground(White)

	Divider = lipgloss.NewStyle().
		Foreground(Muted).
		SetString("─────────────────────────────────────────")

	Banner = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(0, 2)

	TableHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(Secondary).
			PaddingRight(4)

	TableCell = lipgloss.NewStyle().
			Foreground(White).
			PaddingRight(4)

	SectionHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(Secondary).
			MarginTop(1)
)

// Helper functions for common output patterns
func PrintSuccess(msg string) {
	fmt.Println(SuccessStyle.Render("  ✓ " + msg))
}

func PrintWarning(msg string) {
	fmt.Println(WarningStyle.Render("  ⚠ " + msg))
}

func PrintError(msg string) {
	fmt.Println(DangerStyle.Render("  ✗ " + msg))
}

func PrintInfo(msg string) {
	fmt.Println(MutedStyle.Render("  ℹ " + msg))
}

func PrintStep(msg string) {
	fmt.Println(Bold.Render("  → " + msg))
}

func PrintHeader(title string) {
	fmt.Println(Banner.Render(title))
	fmt.Println()
}

func PrintSection(title string) {
	fmt.Println(SectionHeader.Render("  " + title))
}
