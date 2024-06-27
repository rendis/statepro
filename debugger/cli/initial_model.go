package cli

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)
)

var sendEventChoice = &choice{
	title:        "Send event",
	modelBuilder: buildSendEventModel,
}

var loadSnapshotChoice = &choice{
	title:        "Load snapshot",
	modelBuilder: buildLoadSnapshotModel,
}

var historyChoice = &choice{
	title:        "History",
	modelBuilder: buildHistoryViewerModel,
}

var initialItems = []list.Item{
	sendEventChoice,
	loadSnapshotChoice,
	historyChoice,
}

func buildInitialModel(container *smContainer) *model {
	w, h := getTerminalSize()
	x, y := appStyle.GetFrameSize()
	l := list.New(initialItems, list.NewDefaultDelegate(), w-x, h-y)
	l.Title = "State Machine Debugger - Choose an option"
	return &model{
		helperModel: &l,
		container:   container,
		update:      initialModelUpdate,
		view:        initialModelView,
	}
}

func initialModelView(m *model) string {
	hm := m.helperModel.(*list.Model)
	updateInitialModelChoiceTitles(m.container)
	return lipgloss.NewStyle().Margin(1, 2).Render(hm.View())
}

func initialModelUpdate(m *model, teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	hm := m.helperModel.(*list.Model)

	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m, nil
		case "enter":
			c := initialItems[hm.Index()].(*choice)
			if c.modelBuilder != nil {
				return c.modelBuilder(m, m.container)
			}
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		hm.SetSize(msg.Width-h, msg.Height-v)
	}

	nhm, cmd := hm.Update(teaMsg)
	m.helperModel = &nhm
	return m, cmd
}

func updateInitialModelChoiceTitles(container *smContainer) {
	if container == nil {
		container = &smContainer{}
	}

	// events
	sentEvents := 0
	for _, e := range container.events {
		if e.Sent {
			sentEvents++
		}
	}
	t := buildTitleWithRange(*buildTitle(sendEventChoice.title), sentEvents, len(container.events))
	if container.qm == nil || len(container.events) == 0 {
		t = buildDisabledTitle(*t)
	}
	sendEventChoice.editedTitle = t

	// snapshots
	t = buildTitleWithCounter(*buildTitle(loadSnapshotChoice.title), len(container.snapshots))
	if container.qm == nil || len(container.snapshots) == 0 {
		t = buildDisabledTitle(*t)
	}
	loadSnapshotChoice.editedTitle = t

	// history
	t = buildTitleWithCounter(*buildTitle(historyChoice.title), len(container.history))
	if len(container.history) == 0 {
		t = buildDisabledTitle(*t)
	}
	historyChoice.editedTitle = t
}
