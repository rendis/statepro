package cli

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var loadSnapshotKeys = []key.Binding{
	key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "load"),
	),
	key.NewBinding(
		key.WithKeys("v", "m", "t", "s"),
		key.WithHelp("v/m/t/s", "view snapshot"),
	),
	key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

func buildLoadSnapshotModel(prevModel *model, container *smContainer) (tea.Model, tea.Cmd) {
	if len(container.snapshots) == 0 {
		return prevModel, nil
	}

	var items []list.Item
	for _, snapshot := range container.snapshots {
		c := &choice{
			editedTitle: buildTitle(snapshot.Title),
			obj:         snapshot,
		}
		items = append(items, c)
	}

	w, h := getTerminalSize()
	x, y := appStyle.GetFrameSize()
	l := list.New(items, list.NewDefaultDelegate(), w-x, h-y)
	l.Title = "Choose a snapshot"
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return loadSnapshotKeys
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return loadSnapshotKeys
	}

	m := &model{
		helperModel: &l,
		prevModel:   prevModel,
		container:   container,
		view:        loadSnapshotModelView,
		update:      loadSnapshotModelUpdate,
	}

	return m, nil
}

func loadSnapshotModelView(m *model) string {
	hm := m.helperModel.(*list.Model)
	v1 := appStyle.Render(hm.View())

	item, _ := hm.SelectedItem().(*choice)
	h := item.obj.(*debuggerSnapshot)
	part := buildSnapshotPart(h, "m")

	splitView2 := buildJsonSplitViewerModel(part.title, part.content, 3)
	v2 := appStyle.Render(splitView2)

	return lipgloss.JoinHorizontal(lipgloss.Top, v1, v2)
}

func loadSnapshotModelUpdate(m *model, teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	hm := m.helperModel.(*list.Model)

	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.prevModel, nil
		case "v", "m", "t", "s":
			if isFiltering(hm) {
				break
			}
			item, _ := hm.SelectedItem().(*choice)
			dn := item.obj.(*debuggerSnapshot)
			v := buildSnapshotPart(dn, msg.String())
			return buildJsonViewerModel(m, v.title, v.content)
		case "l":
			item, _ := hm.SelectedItem().(*choice)
			dn := item.obj.(*debuggerSnapshot)
			if err := m.container.qm.LoadSnapshot(dn.Snapshot, m.container.smContext); err != nil {
				return m, nil
			}

			m.container.history = []*containerHistory{
				{
					snapshot: m.container.qm.GetSnapshot(),
					context:  copyStructPointer(m.container.smContext),
					event:    getSnapshotEvent(),
					pos:      0,
				},
			}

			resetEvents(m.container)
			return m.prevModel, nil
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		hm.SetSize(msg.Width-h, msg.Height-v)
	}

	nhm, cmd := hm.Update(teaMsg)
	m.helperModel = &nhm
	return m, cmd
}
