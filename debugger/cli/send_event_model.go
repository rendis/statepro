package cli

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rendis/statepro"
	"sort"
)

var (
	sendEventTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
)

var sendEventKeys = []key.Binding{
	key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send"),
	),
	key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	key.NewBinding(
		key.WithKeys("v", "m", "t", "s"),
		key.WithHelp("v/m/t/s", "view snapshot"),
	),
}

func buildSendEventModel(prevModel *model, container *smContainer) (tea.Model, tea.Cmd) {
	if len(container.events) == 0 {
		return prevModel, nil
	}

	var items []list.Item
	for _, event := range container.events {
		c := &choice{
			editedTitle: buildTitle(event.Title),
			description: event.Name,
			obj:         event,
		}
		items = append(items, c)
	}

	w, h := getTerminalSize()
	x, y := appStyle.GetFrameSize()
	l := list.New(items, list.NewDefaultDelegate(), w-x, h-y)
	l.Title = "Send an event"
	l.Styles.Title = sendEventTitleStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return sendEventKeys
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return sendEventKeys
	}
	l.SetShowTitle(true)
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)

	m := &model{
		helperModel: &l,
		prevModel:   prevModel,
		container:   container,
		view:        sendEventModelView,
		update:      sendEventModelUpdate,
	}

	return m, m.Init()
}

func sendEventModelView(m *model) string {
	hm := m.helperModel.(*list.Model)

	items := hm.Items()

	for _, item := range items {
		c := item.(*choice)
		e := c.obj.(*debuggerEvent)
		if e.Sent {
			c.editedTitle = buildMarkedTitle(*buildTitle(e.Title))
		}
	}

	sort.Slice(items, func(i, j int) bool {
		c1 := items[i].(*choice)
		e1 := c1.obj.(*debuggerEvent)

		c2 := items[j].(*choice)
		e2 := c2.obj.(*debuggerEvent)

		if e1.Sent && e2.Sent {
			return e1.Name < e2.Name
		}

		return !e1.Sent && e2.Sent
	})
	v1 := appStyle.Render(hm.View())

	h := m.container.history[len(m.container.history)-1]
	part := buildSnapshotPartFromHistory(h, "m")

	splitView2 := buildJsonSplitViewerModel(part.title, part.content, 3)
	v2 := appStyle.Render(splitView2)

	splitView3 := buildJsonSplitViewerModel(part.title, h.context, 3)
	v3 := appStyle.Render(splitView3)

	return lipgloss.JoinHorizontal(lipgloss.Top, v1, v2, v3)
}

func sendEventModelUpdate(m *model, teaMsg tea.Msg) (tea.Model, tea.Cmd) {
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
			h := m.container.history[len(m.container.history)-1]
			v := buildSnapshotPartFromHistory(h, msg.String())
			return buildJsonViewerModel(m, v.title, v.content)
		case "enter":
			item, _ := hm.SelectedItem().(*choice)
			event := item.obj.(*debuggerEvent)
			evt := statepro.NewEventBuilder(event.Name).
				SetData(event.Params).
				Build()

			success, err := m.container.qm.SendEvent(context.Background(), evt)
			if err != nil {
				fmt.Printf("error sending event: %v\n", err)
				return m, nil
			}

			if !success {
				fmt.Println("event not handled")
				return m, nil
			}

			event.Sent = true
			m.container.history = append(m.container.history, &containerHistory{
				event:    event,
				snapshot: m.container.qm.GetSnapshot(),
				context:  copyStructPointer(m.container.smContext),
				pos:      len(m.container.history),
			})

			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		hm.SetSize(msg.Width-h, msg.Height-v)
	}

	nhm, cmd := hm.Update(teaMsg)
	m.helperModel = &nhm
	return m, cmd
}
