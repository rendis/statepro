package cli

import tea "github.com/charmbracelet/bubbletea"

type choice struct {
	title        string
	editedTitle  *string
	description  string
	modelBuilder func(prevModel *model, container *smContainer) (tea.Model, tea.Cmd)
	obj          any
}

func (c *choice) Title() string {
	if c.editedTitle != nil {
		return *c.editedTitle
	}
	return c.title
}

func (c *choice) Description() string {
	return c.description
}

func (c *choice) FilterValue() string {
	if c.editedTitle != nil {
		return *c.editedTitle
	}
	return c.title
}
