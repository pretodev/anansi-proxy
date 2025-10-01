package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pretodev/anansi-proxy/internal/parser"
	"github.com/pretodev/anansi-proxy/internal/state"
)

func Render(sm *state.StateManager, endpoint *parser.EndpointSchema) error {
	p := tea.NewProgram(initialModel(sm, endpoint))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

var (
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	endpoint     *parser.EndpointSchema
	cursor       int
	keys         keyMap
	stateManager *state.StateManager
}

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Quit key.Binding
}

func initialModel(sm *state.StateManager, endpoint *parser.EndpointSchema) model {
	return model{
		endpoint:     endpoint,
		cursor:       0,
		stateManager: sm,
		keys: keyMap{
			Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
			Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
			Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		},
	}
}

// Init is the first function to be executed.
func (m model) Init() tea.Cmd {
	return nil
}

// Update is called when "something happens", like a key press.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.endpoint.Responses)-1 {
				m.cursor++
			}
		}
	}

	m.stateManager.SetIndex(m.cursor)
	return m, nil
}

// View renders the UI based on the current model state.
func (m model) View() string {
	var b strings.Builder

	b.WriteString("Select a response for the server:\n\n")

	for i, res := range m.endpoint.Responses {
		cursor := "  " // Not selected
		line := fmt.Sprintf("[%d] %s", res.StatusCode, res.Title)

		if m.cursor == i {
			cursor = "> " // Selected
			line = selectedItemStyle.Render(line)
		}

		b.WriteString(cursor + line + "\n")
	}

	help := fmt.Sprintf("\n%s  %s  %s", m.keys.Up.Help(), m.keys.Down.Help(), m.keys.Quit.Help())
	b.WriteString(helpStyle.Render(help))

	return b.String()
}
