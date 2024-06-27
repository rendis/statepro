package cli

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rendis/statepro/v3/instrumentation"
)

var historyViewerKeys = []key.Binding{
	key.NewBinding(
		key.WithKeys("v", "m", "t", "s"),
		key.WithHelp("v/m/t/s", "view snapshot"),
	),
	key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "rollback"),
	),
}

type snapshotExtractor func(snapshot *instrumentation.MachineSnapshot) (string, any)

var historySnapshotKeys = map[string]snapshotExtractor{
	"m": func(snapshot *instrumentation.MachineSnapshot) (string, any) {
		return "resume segment", snapshot.Resume
	},
	"t": func(snapshot *instrumentation.MachineSnapshot) (string, any) {
		return "tracking segment", snapshot.Tracking
	},
	"s": func(snapshot *instrumentation.MachineSnapshot) (string, any) {
		return "snapshot segment", snapshot.Snapshots
	},
}

func buildHistoryViewerModel(prevModel *model, container *smContainer) (tea.Model, tea.Cmd) {
	if len(container.snapshots) == 0 {
		return prevModel, nil
	}

	var items []list.Item
	for _, h := range container.history {
		c := &choice{
			editedTitle: buildTitle(h.event.Name),
			description: h.event.Title,
			obj:         h,
		}
		items = append(items, c)
	}

	w, h := getTerminalSize()
	x, y := appStyle.GetFrameSize()
	l := list.New(items, list.NewDefaultDelegate(), w-x, h-y)
	l.Title = "View history"
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return historyViewerKeys
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return historyViewerKeys
	}

	m := &model{
		helperModel: &l,
		prevModel:   prevModel,
		container:   container,
		view:        historyViewerModelView,
		update:      historyViewerModelUpdate,
	}

	return m, nil
}

func historyViewerModelView(m *model) string {
	hm := m.helperModel.(*list.Model)
	return lipgloss.NewStyle().Margin(1, 2).Render(hm.View())
}

func historyViewerModelUpdate(m *model, teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	hm := m.helperModel.(*list.Model)

	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.prevModel, nil
		case "r":
			item, _ := hm.SelectedItem().(*choice)
			h := item.obj.(*containerHistory)
			if err := m.container.qm.LoadSnapshot(h.snapshot, m.container.smContext); err != nil {
				m.err = err
				return m, nil
			}
			m.container.history = m.container.history[:h.pos+1]
			markEventsFromHistory(m.container)
			return m.prevModel, nil
		case "v", "m", "t", "s":
			if isFiltering(hm) {
				break
			}
			item, _ := hm.SelectedItem().(*choice)
			h := item.obj.(*containerHistory)
			v := buildSnapshotPartFromHistory(h, msg.String())
			return buildJsonViewerModel(m, v.title, v.content)
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		hm.SetSize(msg.Width-h, msg.Height-v)
	}

	nhm, cmd := hm.Update(teaMsg)
	m.helperModel = &nhm
	return m, cmd
}

func buildSnapshotPartFromHistory(history *containerHistory, key string) *version {

	extractor, ok := historySnapshotKeys[key]
	var snapshot any = history.snapshot
	var titleSegment = "All"
	if ok {
		titleSegment, snapshot = extractor(history.snapshot)
	}

	title := fmt.Sprintf("Snapshot: %s (%s)", history.event.Name, *buildYellowTitle(titleSegment))
	return &version{
		title:   title,
		content: snapshot,
	}
}

func buildSnapshotPart(ds *debuggerSnapshot, key string) *version {

	extractor, ok := historySnapshotKeys[key]
	var snapshot any = ds.Snapshot
	var titleSegment = "All"
	if ok {
		titleSegment, snapshot = extractor(ds.Snapshot)
	}

	title := fmt.Sprintf("Snapshot: %s (%s)", ds.Title, *buildYellowTitle(titleSegment))
	return &version{
		title:   title,
		content: snapshot,
	}
}
