package conflict

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if len(m.choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(m.choices))
	}
	if m.resolved != false {
		t.Errorf("Expected resolved=false, got %v", m.resolved)
	}
}

func TestModelInit(t *testing.T) {
	m := initialModel()
	cmd := m.Init()
	if cmd != nil {
		t.Errorf("Expected Init to return nil, got %v", cmd)
	}
}

func TestModelUpdate(t *testing.T) {
	tests := []struct {
		name           string
		initial        model
		msg            tea.Msg
		wantCursor     int
		wantResolved   bool
		wantQuitting   bool
		wantResolution Resolution
	}{
		{
			name: "up key moves cursor up",
			initial: model{
				choices: []string{"Keep Local", "Keep Remote", "Abort"},
				cursor:  1,
			},
			msg:        tea.KeyMsg{Type: tea.KeyUp},
			wantCursor: 0,
		},
		{
			name: "down key moves cursor down",
			initial: model{
				choices: []string{"Keep Local", "Keep Remote", "Abort"},
				cursor:  0,
			},
			msg:        tea.KeyMsg{Type: tea.KeyDown},
			wantCursor: 1,
		},
		{
			name: "down key does not move past last item",
			initial: model{
				choices: []string{"Keep Local", "Keep Remote", "Abort"},
				cursor:  2,
			},
			msg:        tea.KeyMsg{Type: tea.KeyDown},
			wantCursor: 2,
		},
		{
			name: "enter selects current resolution",
			initial: model{
				choices: []string{"Keep Local", "Keep Remote", "Abort"},
				cursor:  1,
			},
			msg:            tea.KeyMsg{Type: tea.KeyEnter},
			wantCursor:     1,
			wantResolved:   true,
			wantResolution: KeepRemote,
		},
		{
			name: "ctrl+c sets quitting (does not resolve)",
			initial: model{
				choices: []string{"Keep Local", "Keep Remote", "Abort"},
				cursor:  1,
			},
			msg:          tea.KeyMsg{Type: tea.KeyCtrlC},
			wantCursor:   1,
			wantQuitting: true,
			// resolved remains false, resolution remains 0
		},
		{
			name: "q key sets quitting",
			initial: model{
				choices: []string{"Keep Local", "Keep Remote", "Abort"},
				cursor:  0,
			},
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantCursor:   0,
			wantQuitting: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, cmd := tt.initial.Update(tt.msg)
			updated := m.(model)

			if updated.cursor != tt.wantCursor {
				t.Errorf("Cursor: got %d, want %d", updated.cursor, tt.wantCursor)
			}
			if updated.resolved != tt.wantResolved {
				t.Errorf("Resolved: got %v, want %v", updated.resolved, tt.wantResolved)
			}
			if updated.quitting != tt.wantQuitting {
				t.Errorf("Quitting: got %v, want %v", updated.quitting, tt.wantQuitting)
			}
			if updated.resolution != tt.wantResolution {
				t.Errorf("Resolution: got %v, want %v", updated.resolution, tt.wantResolution)
			}
			if cmd != nil && !tt.wantResolved && !tt.wantQuitting {
				t.Error("Expected no command but got one")
			}
		})
	}
}

func TestModelView(t *testing.T) {
	m := model{
		choices: []string{"Keep Local", "Keep Remote", "Abort"},
		cursor:  1,
	}

	view := m.View()

	// Should contain danger warning
	if !strings.Contains(view, "Sync Conflict") {
		t.Error("View missing conflict header")
	}
	// Should contain choices
	if !strings.Contains(view, "Keep Local") || !strings.Contains(view, "Keep Remote") || !strings.Contains(view, "Abort") {
		t.Error("View missing choices")
	}
	// Should contain help text
	if !strings.Contains(view, "Press up/down") {
		t.Error("View missing help text")
	}
}

func TestModelViewWhenResolved(t *testing.T) {
	m := model{
		choices:    []string{"Keep Local", "Keep Remote", "Abort"},
		cursor:     0,
		resolved:   true,
		resolution: KeepLocal,
	}

	view := m.View()
	expected := "Selected: Keep Local\n"
	if view != expected {
		t.Errorf("Expected %q, got %q", expected, view)
	}
}

func TestModelViewWhenQuitting(t *testing.T) {
	m := model{
		choices:  []string{"Keep Local", "Keep Remote", "Abort"},
		quitting: true,
	}

	view := m.View()
	if view != "Aborting...\n" {
		t.Errorf("Expected abort message, got %q", view)
	}
}

func TestPromptConflictResolution(t *testing.T) {
	// This test would run the full TUI which is hard to test without user input
	// We'll just test that the function exists and returns appropriate types
	// In a real test, we'd mock the tea.Program
	_ = PromptConflictResolution
	// Placeholder to ensure function is present and compiles
}
