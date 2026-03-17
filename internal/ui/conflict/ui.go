package conflict

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noeltz/n0man/internal/ui"
)

type Resolution int

const (
	KeepLocal Resolution = iota
	KeepRemote
	AbortAndManual
)

type model struct {
	cursor     int
	choices    []string
	resolved   bool
	resolution Resolution
	quitting   bool
}

func initialModel() model {
	return model{
		choices: []string{"Keep Local Changes", "Keep Remote Changes", "Abort and Fix Manually"},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.resolved = true
			m.resolution = Resolution(m.cursor)
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Aborting...\n"
	}
	if m.resolved {
		return fmt.Sprintf("Selected: %s\n", m.choices[m.cursor])
	}

	s := ui.DangerStyle.Render("\n  🚨 Sync Conflict Detected!\n")
	s += ui.MutedStyle.Render("  Choose how to resolve the conflict in the repository:\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ui.PrimaryStyle.Render(">")
			s += fmt.Sprintf("  %s %s\n", cursor, ui.Bold.Render(choice))
		} else {
			s += fmt.Sprintf("    %s\n", ui.MutedStyle.Render(choice))
		}
	}

	s += ui.MutedStyle.Render("\n  (Press up/down to move, enter to select, q to abort)\n")

	return s
}

// PromptConflictResolution runs the interactive TUI and returns the selected resolution
func PromptConflictResolution() (Resolution, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return AbortAndManual, err
	}

	finalModel := m.(model)
	if finalModel.quitting {
		return AbortAndManual, fmt.Errorf("user aborted")
	}

	return finalModel.resolution, nil
}
