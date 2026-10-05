package cli

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	header      string
	helperModel any
	prevModel   *model
	md          map[string]any
	err         error
	init        func(m *model) tea.Cmd
	update      func(m *model, msg tea.Msg) (tea.Model, tea.Cmd)
	view        func(m *model) string
	container   *smContainer
}

func (m *model) Init() tea.Cmd {
	if m.init != nil {
		return m.init(m)
	}
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.update(m, msg)
}

func (m *model) View() string {
	view := m.view(m)
	if m.err != nil {
		return fmt.Sprintf("%s\nError: %v", view, m.err)
	}
	return view
}
