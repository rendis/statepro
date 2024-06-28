package cli

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func buildJsonSplitViewerModel(title string, obj any) string {
	coloredJson, _ := formatToJson(obj, true)
	content := string(coloredJson)

	x, y := appStyle.GetFrameSize()
	w, h := getTerminalSize()
	vp := viewport.New(w-x, h-y)
	headerHeight := lipgloss.Height(jsonViewerHeaderView(title, &vp))
	vp.YPosition = headerHeight
	vp.HighPerformanceRendering = false
	vp.SetContent(content)
	vp.YPosition = headerHeight + 1

	return vp.View()
}
