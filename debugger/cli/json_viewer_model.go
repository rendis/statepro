package cli

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

const (
	jsonViewerReadyKey   = "ready"
	jsonViewerHeaderKey  = "header"
	jsonViewerContentKey = "content"
)

var (
	jsonViewerTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	jsonViewerInfoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return jsonViewerTitleStyle.BorderStyle(b)
	}()
)

func buildJsonViewerModel(prevModel *model, title string, obj any) (tea.Model, tea.Cmd) {
	vp := viewport.Model{}
	header := title
	m := &model{
		header:      header,
		prevModel:   prevModel,
		helperModel: &vp,
		md: map[string]interface{}{
			jsonViewerReadyKey:  false,
			jsonViewerHeaderKey: header,
		},
		view:   jsonViewerView,
		update: jsonViewerUpdate,
	}

	return m, jsonViewerInit(obj, m)
}

func jsonViewerInit(source any, m *model) tea.Cmd {
	coloredJson, err := formatToJson(source, true)
	if err != nil {
		m.err = err
		return nil
	}

	content := string(coloredJson)
	m.md[jsonViewerContentKey] = content

	return tea.Tick(0, func(time.Time) tea.Msg {
		w, h := getTerminalSize()
		return tea.WindowSizeMsg{Width: w, Height: h}
	})
}

func jsonViewerView(m *model) string {
	ready := getBoolFromMD(jsonViewerReadyKey, m.md)
	if !ready {
		return "\n  Initializing..."
	}
	vp := m.helperModel.(*viewport.Model)
	header := getStrFromMD(jsonViewerHeaderKey, m.md)
	return fmt.Sprintf("%s\n%s\n%s", jsonViewerHeaderView(header, vp), vp.View(), jsonViewerFooterView(vp))
}

func jsonViewerUpdate(m *model, teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	vp := m.helperModel.(*viewport.Model)

	switch msg := teaMsg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.prevModel, nil
		case "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		header := getStrFromMD(jsonViewerHeaderKey, m.md)
		headerHeight := lipgloss.Height(jsonViewerHeaderView(header, vp))
		footerHeight := lipgloss.Height(jsonViewerFooterView(vp))
		verticalMarginHeight := headerHeight + footerHeight

		ready := getBoolFromMD(jsonViewerReadyKey, m.md)

		if !ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			nvp := viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			vp = &nvp
			vp.YPosition = headerHeight
			vp.HighPerformanceRendering = false
			content := getStrFromMD(jsonViewerContentKey, m.md)
			vp.SetContent(content)
			m.md[jsonViewerReadyKey] = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			vp.YPosition = headerHeight + 1
		} else {
			vp.Width = msg.Width
			vp.Height = msg.Height - verticalMarginHeight
		}

	}

	// Handle keyboard and mouse events in the viewport
	nvp, cmd := vp.Update(teaMsg)
	m.helperModel = &nvp
	return m, cmd
}

func jsonViewerHeaderView(header string, vp *viewport.Model) string {
	title := jsonViewerTitleStyle.Render(header)
	line := strings.Repeat("─", max(0, vp.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func jsonViewerFooterView(vp *viewport.Model) string {
	info := jsonViewerInfoStyle.Render(fmt.Sprintf("%3.f%%", vp.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, vp.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
