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

func Render(sm *state.StateManager, res []parser.Response) error {
	if len(res) == 0 {
		fmt.Println("Nenhuma resposta encontrada para exibir.")
		return nil
	}

	p := tea.NewProgram(initialModel(res, sm))
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
	responses    []parser.Response
	cursor       int
	keys         keyMap
	stateManager *state.StateManager
}

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Quit key.Binding
}

func initialModel(res []parser.Response, sm *state.StateManager) model {
	return model{
		responses:    res,
		cursor:       0,
		stateManager: sm,
		keys: keyMap{
			Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
			Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
			Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
		},
	}
}

// Init é a primeira função a ser executada.
func (m model) Init() tea.Cmd {
	return nil
}

// Update é chamada quando "algo acontece", como o pressionar de uma tecla.
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
			if m.cursor < len(m.responses)-1 {
				m.cursor++
			}
		}
	}

	m.stateManager.SetIndex(m.cursor)
	return m, nil
}

// View renderiza a UI com base no estado atual do model.
func (m model) View() string {
	var b strings.Builder

	b.WriteString("Selecione uma resposta para o servidor:\n\n")

	for i, res := range m.responses {
		cursor := "  " // Não selecionado
		line := fmt.Sprintf("[%d] %s", res.StatusCode, res.Title)

		if m.cursor == i {
			cursor = "> " // Selecionado
			line = selectedItemStyle.Render(line)
		}

		b.WriteString(cursor + line + "\n")
	}

	help := fmt.Sprintf("\n%s  %s  %s", m.keys.Up.Help(), m.keys.Down.Help(), m.keys.Quit.Help())
	b.WriteString(helpStyle.Render(help))

	return b.String()
}
