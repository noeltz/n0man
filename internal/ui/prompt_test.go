package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPromptModelInit(t *testing.T) {
	m := PromptModel{
		Question: "Test question",
		Choices:  []string{"A", "B", "C"},
	}

	cmd := m.Init()
	if cmd != nil {
		t.Errorf("Expected Init to return nil, got %v", cmd)
	}
}

func TestPromptModelUpdate(t *testing.T) {
	tests := []struct {
		name       string
		initial    PromptModel
		msg        tea.Msg
		wantMove   int
		wantQuit   bool
		wantChoice int
	}{
		{
			name: "up key moves cursor up",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   1,
			},
			msg:      tea.KeyMsg{Type: tea.KeyUp},
			wantMove: -1,
		},
		{
			name: "down key moves cursor down",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   0,
			},
			msg:      tea.KeyMsg{Type: tea.KeyDown},
			wantMove: +1,
		},
		{
			name: "down key does not move past last item",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   2,
			},
			msg:      tea.KeyMsg{Type: tea.KeyDown},
			wantMove: 0,
		},
		{
			name: "enter selects current choice and quits",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   1,
			},
			msg:        tea.KeyMsg{Type: tea.KeyEnter},
			wantMove:   0,
			wantQuit:   true,
			wantChoice: 1,
		},
		{
			name: "space selects like enter",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   0,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}},
			wantMove:   0,
			wantQuit:   true,
			wantChoice: 0,
		},
		{
			name: "ctrl+c sets quitting and choice -1",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   1,
			},
			msg:        tea.KeyMsg{Type: tea.KeyCtrlC},
			wantMove:   0,
			wantQuit:   true,
			wantChoice: -1,
		},
		{
			name: "k key moves up",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   1,
			},
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			wantMove: -1,
		},
		{
			name: "j key moves down",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   0,
			},
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantMove: +1,
		},
		{
			name: "q key quits",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   1,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantMove:   0,
			wantQuit:   true,
			wantChoice: -1,
		},
		{
			name: "esc key quits",
			initial: PromptModel{
				Question: "Choose",
				Choices:  []string{"A", "B", "C"},
				Cursor:   1,
			},
			msg:        tea.KeyMsg{Type: tea.KeyEsc},
			wantMove:   0,
			wantQuit:   true,
			wantChoice: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, cmd := tt.initial.Update(tt.msg)
			updated := m.(PromptModel)

			// Compute expected cursor with proper bounds
			expectedCursor := tt.initial.Cursor + tt.wantMove
			if expectedCursor < 0 {
				expectedCursor = 0
			}
			if expectedCursor > len(tt.initial.Choices)-1 {
				expectedCursor = len(tt.initial.Choices) - 1
			}

			// Check cursor position
			if updated.Cursor != expectedCursor {
				t.Errorf("Cursor: got %d, want %d", updated.Cursor, expectedCursor)
			}

			// Check quit and choice
			if updated.Quitting != tt.wantQuit {
				t.Errorf("Quitting: got %v, want %v", updated.Quitting, tt.wantQuit)
			}
			if updated.Choice != tt.wantChoice {
				t.Errorf("Choice: got %d, want %d", updated.Choice, tt.wantChoice)
			}

			// Command check: If quitting, cmd should be non-nil (tea.Quit)
			cmdReturned := cmd != nil
			if tt.wantQuit && !cmdReturned {
				t.Error("Expected a quit command but got nil")
			}
			if !tt.wantQuit && cmdReturned {
				t.Error("Expected no command but got one")
			}
		})
	}
}

func TestPromptModelView(t *testing.T) {
	m := PromptModel{
		Question: "Select an option",
		Choices:  []string{"Yes", "No"},
		Cursor:   0,
	}

	view := m.View()

	// Should contain the question
	if !strings.Contains(view, "Select an option") {
		t.Error("View missing question")
	}
	// Should contain choices
	if !strings.Contains(view, "Yes") || !strings.Contains(view, "No") {
		t.Error("View missing choices")
	}
	// Should contain help text
	if !strings.Contains(view, "arrow keys") {
		t.Error("View missing help text")
	}
}

func TestPromptModelViewWhenQuitting(t *testing.T) {
	m := PromptModel{
		Question: "Select",
		Choices:  []string{"A", "B"},
		Quitting: true,
	}

	view := m.View()
	if view != "" {
		t.Errorf("Expected empty view when quitting, got %q", view)
	}
}

func TestPromptModelViewCursorIndicator(t *testing.T) {
	m := PromptModel{
		Question: "Choose",
		Choices:  []string{"A", "B", "C"},
		Cursor:   1,
	}

	view := m.View()

	// The selected cursor should have a cursor indicator "▸"
	if !strings.Contains(view, "▸") {
		t.Error("View missing cursor indicator for selected item")
	}
}
