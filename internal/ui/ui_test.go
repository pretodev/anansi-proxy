package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pretodev/anansi-proxy/internal/endpoint"
	"github.com/pretodev/anansi-proxy/internal/state"
)

func createTestEndpoint() *endpoint.EndpointSchema {
	return &endpoint.EndpointSchema{
		Route:  "/api/test",
		Accept: "application/json",
		Body:   "{}",
		Responses: map[int][]endpoint.Response{
			200: {
				{
					Title:       "Success",
					Body:        `{"status": "ok"}`,
					ContentType: "application/json",
					StatusCode:  200,
				},
			},
			404: {
				{
					Title:       "Not Found",
					Body:        `{"error": "not found"}`,
					ContentType: "application/json",
					StatusCode:  404,
				},
			},
			500: {
				{
					Title:       "Server Error",
					Body:        `{"error": "internal server error"}`,
					ContentType: "application/json",
					StatusCode:  500,
				},
			},
		},
	}
}

func TestInitialModel(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()

	m := initialModel(sm, ep)

	if m.endpoint != ep {
		t.Error("endpoint not set correctly")
	}
	if m.cursor != 0 {
		t.Errorf("cursor should start at 0, got %d", m.cursor)
	}
	if m.stateManager != sm {
		t.Error("stateManager not set correctly")
	}

	// Verify key bindings
	if m.keys.Up.Keys()[0] != "up" {
		t.Error("Up key binding not set correctly")
	}
	if m.keys.Down.Keys()[0] != "down" {
		t.Error("Down key binding not set correctly")
	}
	if m.keys.Quit.Keys()[0] != "q" {
		t.Error("Quit key binding not set correctly")
	}
}

func TestModel_Init(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update_QuitKey(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	tests := []struct {
		name string
		key  string
	}{
		{"quit with q", "q"},
		{"quit with ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(tt.key[0])}}
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			_, cmd := m.Update(msg)
			if cmd == nil {
				t.Error("Update should return tea.Quit command")
			}
		})
	}
}

func TestModel_Update_UpKey(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)
	m.cursor = 2 // Start at position 2

	// Press up key
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after pressing up, got %d", m.cursor)
	}
	if sm.Index() != 1 {
		t.Errorf("state manager index should be 1, got %d", sm.Index())
	}

	// Press up again
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != 0 {
		t.Errorf("cursor should be 0 after pressing up twice, got %d", m.cursor)
	}
	if sm.Index() != 0 {
		t.Errorf("state manager index should be 0, got %d", sm.Index())
	}

	// Press up when at top (should stay at 0)
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 when at top, got %d", m.cursor)
	}
}

func TestModel_Update_DownKey(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	// Press down key
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after pressing down, got %d", m.cursor)
	}
	if sm.Index() != 1 {
		t.Errorf("state manager index should be 1, got %d", sm.Index())
	}

	// Press down again
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != 2 {
		t.Errorf("cursor should be 2 after pressing down twice, got %d", m.cursor)
	}
	if sm.Index() != 2 {
		t.Errorf("state manager index should be 2, got %d", sm.Index())
	}

	// Press down when at bottom (should stay at 2)
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != 2 {
		t.Errorf("cursor should stay at 2 when at bottom, got %d", m.cursor)
	}
}

func TestModel_Update_AlternativeKeys(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	// Test 'j' for down
	msgJ := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := m.Update(msgJ)
	m = updatedModel.(model)

	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after pressing 'j', got %d", m.cursor)
	}

	// Test 'k' for up
	msgK := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ = m.Update(msgK)
	m = updatedModel.(model)

	if m.cursor != 0 {
		t.Errorf("cursor should be 0 after pressing 'k', got %d", m.cursor)
	}
}

func TestModel_View(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	view := m.View()

	// Check that view contains expected text
	if !strings.Contains(view, "Select a response for the server:") {
		t.Error("View should contain header text")
	}

	// Check that all responses are shown
	if !strings.Contains(view, "[200] Success") {
		t.Error("View should contain first response")
	}
	if !strings.Contains(view, "[404] Not Found") {
		t.Error("View should contain second response")
	}
	if !strings.Contains(view, "[500] Server Error") {
		t.Error("View should contain third response")
	}

	// Check that help is shown
	if !strings.Contains(view, "move up") {
		t.Error("View should contain help text for up")
	}
	if !strings.Contains(view, "move down") {
		t.Error("View should contain help text for down")
	}
	if !strings.Contains(view, "quit") {
		t.Error("View should contain help text for quit")
	}
}

func TestModel_View_SelectedItem(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	// First item should be selected initially
	view := m.View()
	lines := strings.Split(view, "\n")

	// Find the line with first response
	var firstResponseLine string
	for _, line := range lines {
		if strings.Contains(line, "[200] Success") {
			firstResponseLine = line
			break
		}
	}

	if !strings.HasPrefix(strings.TrimSpace(firstResponseLine), ">") {
		t.Error("First item should show '>' when selected")
	}

	// Move to second item
	m.cursor = 1
	view = m.View()
	lines = strings.Split(view, "\n")

	// Check that first item is not selected
	for _, line := range lines {
		if strings.Contains(line, "[200] Success") {
			if strings.HasPrefix(strings.TrimSpace(line), ">") {
				t.Error("First item should not show '>' when not selected")
			}
		}
		if strings.Contains(line, "[404] Not Found") {
			if !strings.HasPrefix(strings.TrimSpace(line), ">") {
				t.Error("Second item should show '>' when selected")
			}
		}
	}
}

func TestModel_View_EmptyResponses(t *testing.T) {
	sm := state.New(0)
	ep := &endpoint.EndpointSchema{
		Route:     "/api/test",
		Accept:    "application/json",
		Body:      "{}",
		Responses: map[int][]endpoint.Response{},
	}
	m := initialModel(sm, ep)

	view := m.View()

	// Should still show header and help
	if !strings.Contains(view, "Select a response for the server:") {
		t.Error("View should contain header text even with no responses")
	}
	if !strings.Contains(view, "quit") {
		t.Error("View should contain help text even with no responses")
	}
}

func TestModel_View_SingleResponse(t *testing.T) {
	sm := state.New(1)
	ep := &endpoint.EndpointSchema{
		Route:  "/api/test",
		Accept: "application/json",
		Body:   "{}",
		Responses: map[int][]endpoint.Response{
			200: {
				{
					Title:       "Only Response",
					Body:        `{"status": "ok"}`,
					ContentType: "application/json",
					StatusCode:  200,
				},
			},
		},
	}
	m := initialModel(sm, ep)

	view := m.View()

	if !strings.Contains(view, "[200] Only Response") {
		t.Error("View should contain the single response")
	}
	if !strings.Contains(view, ">") {
		t.Error("Single response should be selected by default")
	}
}

func TestModel_Update_StateManagerSync(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	// Verify initial state
	if sm.Index() != 0 {
		t.Errorf("initial state manager index should be 0, got %d", sm.Index())
	}

	// Move down and verify state manager is updated
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(model)

	if sm.Index() != m.cursor {
		t.Errorf("state manager index (%d) should match cursor (%d)", sm.Index(), m.cursor)
	}

	// Move down again
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(model)

	if sm.Index() != m.cursor {
		t.Errorf("state manager index (%d) should match cursor (%d)", sm.Index(), m.cursor)
	}

	// Move up and verify
	msgUp := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(msgUp)
	m = updatedModel.(model)

	if sm.Index() != m.cursor {
		t.Errorf("state manager index (%d) should match cursor (%d)", sm.Index(), m.cursor)
	}
}

func TestKeyMap_Bindings(t *testing.T) {
	km := keyMap{
		Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
	}

	// Test Up binding
	upKeys := km.Up.Keys()
	if len(upKeys) != 2 {
		t.Errorf("Up should have 2 keys, got %d", len(upKeys))
	}
	if upKeys[0] != "up" || upKeys[1] != "k" {
		t.Error("Up keys should be 'up' and 'k'")
	}

	// Test Down binding
	downKeys := km.Down.Keys()
	if len(downKeys) != 2 {
		t.Errorf("Down should have 2 keys, got %d", len(downKeys))
	}
	if downKeys[0] != "down" || downKeys[1] != "j" {
		t.Error("Down keys should be 'down' and 'j'")
	}

	// Test Quit binding
	quitKeys := km.Quit.Keys()
	if len(quitKeys) != 2 {
		t.Errorf("Quit should have 2 keys, got %d", len(quitKeys))
	}
	if quitKeys[0] != "q" || quitKeys[1] != "ctrl+c" {
		t.Error("Quit keys should be 'q' and 'ctrl+c'")
	}
}

func TestModel_Update_UnknownKey(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)
	initialCursor := m.cursor

	// Press unknown key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(model)

	if m.cursor != initialCursor {
		t.Error("cursor should not change for unknown key")
	}
	if cmd != nil {
		t.Error("unknown key should not trigger any command")
	}
}

func TestModel_Update_NavigationBoundaries(t *testing.T) {
	sm := state.New(3)
	ep := createTestEndpoint()
	m := initialModel(sm, ep)

	// Test multiple ups at top
	msgUp := tea.KeyMsg{Type: tea.KeyUp}
	for i := 0; i < 5; i++ {
		updatedModel, _ := m.Update(msgUp)
		m = updatedModel.(model)
	}
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 after multiple ups, got %d", m.cursor)
	}

	// Test multiple downs at bottom
	msgDown := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 10; i++ {
		updatedModel, _ := m.Update(msgDown)
		m = updatedModel.(model)
	}
	maxIndex := ep.CountResponses() - 1
	if m.cursor != maxIndex {
		t.Errorf("cursor should stay at %d after multiple downs, got %d", maxIndex, m.cursor)
	}
}

// Note: Render() function is not tested here as it's an interactive function
// that starts the Bubble Tea program and blocks until user interaction.
// It's better suited for integration/manual testing rather than unit tests.
// The underlying model, Init(), Update(), and View() functions are all tested above,
// which provides comprehensive coverage of the UI logic.
