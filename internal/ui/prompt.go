package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for responsive TUI layout
var (
	promptStyle   = lipgloss.NewStyle().MaxWidth(80).Padding(0, 2)
	choiceStyle   = lipgloss.NewStyle().Foreground(White)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(Secondary)
	cursorStyle   = lipgloss.NewStyle().Foreground(Secondary)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)

type PromptModel struct {
	Question string
	Choices  []string
	Cursor   int
	Choice   int
	Quitting bool
}

func (m PromptModel) Init() tea.Cmd {
	return nil
}

func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Choice = -1
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter", " ":
			m.Choice = m.Cursor
			m.Quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m PromptModel) View() string {
	if m.Quitting {
		return ""
	}

	s := promptStyle.Render(fmt.Sprintf("\n%s\n", m.Question))

	for i, choice := range m.Choices {
		cursor := "  "
		style := choiceStyle
		if m.Cursor == i {
			cursor = cursorStyle.Render("▸ ")
			style = selectedStyle
		}
		s += promptStyle.Render(fmt.Sprintf("%s%s\n", cursor, style.Render(choice)))
	}

	s += helpStyle.Render("\n  (use arrow keys or j/k to navigate, enter to select)\n")
	return s
}

// SelectPrompt displays a question and returns the index of the selected choice
func SelectPrompt(question string, choices []string) (int, error) {
	m := PromptModel{
		Question: question,
		Choices:  choices,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return -1, err
	}

	return finalModel.(PromptModel).Choice, nil
}

// PromptConfirm displays a yes/no question and returns true if the user selects yes
func PromptConfirm(question string, defaultYes bool) (bool, error) {
	choices := []string{"Yes", "No"}
	if !defaultYes {
		choices = []string{"No", "Yes"}
	}

	idx, err := SelectPrompt(question, choices)
	if err != nil {
		return false, err
	}

	if defaultYes {
		return idx == 0, nil
	}
	return idx == 1, nil
}

// InputPrompt displays a question and returns the user's text input
func InputPrompt(question string) (string, error) {
	fmt.Printf("\n  %s ", question)
	var input string
	_, err := fmt.Scanln(&input)
	return input, err
}
