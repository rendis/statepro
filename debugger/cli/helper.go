package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TylerBrock/colorjson"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

type winSize struct {
	Row uint16
	Col uint16
}

type version struct {
	title   string
	content any
}

func loadJSON(t any, path string, ignoreNotExist bool) error {
	// Check if file exists
	if !exists(path) {
		if ignoreNotExist {
			return nil
		}
		return fmt.Errorf("file does not exist: %s", path)
	}

	// Read file
	arrByte, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = json.Unmarshal(arrByte, t)
	if err != nil {
		return fmt.Errorf("error deserializing json (%s): %w", path, err)
	}

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func buildTitle(title string) *string {
	t := fmt.Sprintf("%s %s", "◯", title)
	return &t
}

func buildMarkedTitle(title string) *string {
	t := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Render("✓")
	t = fmt.Sprintf("%s (%s)", title, t)
	return &t
}

func buildDisabledTitle(title string) *string {
	t := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080")).
		Render(fmt.Sprintf("%s", title))
	return &t
}

func buildYellowTitle(title string) *string {
	t := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Render(fmt.Sprintf("%s", title))
	return &t
}

func buildTitleWithCounter(title string, counter int) *string {
	t := fmt.Sprintf("%s (%d)", title, counter)
	return &t
}

func buildTitleWithRange(title string, l, r int) *string {
	t := fmt.Sprintf("%s (%d/%d)", title, l, r)
	return &t
}

func getTerminalSize() (int, int) {
	ws := &winSize{}
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if err != 0 {
		return 170, 25
	}

	width := int(ws.Col)
	height := int(ws.Row)
	return width, height
}

func getBoolFromMD(key string, md map[string]any) bool {
	if v, ok := md[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}

	return false
}

func getStrFromMD(key string, md map[string]any) string {
	if v, ok := md[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getSnapshotEvent() *debuggerEvent {
	return &debuggerEvent{
		Name:  "Initial snapshot",
		Title: "State machine initial snapshot",
	}
}

func isPointerOrNil(v any) bool {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || !val.IsValid() {
		return false
	}
	if val.IsNil() {
		return true
	}

	return val.Elem().Kind() == reflect.Struct
}

func copyStructPointer(v any) any {
	val := reflect.ValueOf(v)

	if val.IsNil() {
		return nil
	}

	structValue := val.Elem()
	structCopy := reflect.New(structValue.Type()).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		structCopy.Field(i).Set(structValue.Field(i))
	}

	return structCopy.Addr().Interface()
}

func formatToJson(source any, enableFormater bool) ([]byte, error) {
	jsonData, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	if !enableFormater {
		var basicJson bytes.Buffer
		_ = json.Indent(&basicJson, jsonData, "", "  ")
		return basicJson.Bytes(), nil
	}

	var obj map[string]interface{}
	_ = json.Unmarshal(jsonData, &obj)

	formatter := colorjson.NewFormatter()
	formatter.Indent = 4
	coloredJson, err := formatter.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return coloredJson, nil
}

func resetEvents(container *smContainer) {
	for _, event := range container.events {
		event.Sent = false
	}
}

func markEventsFromHistory(container *smContainer) {
	resetEvents(container)
	var eventsMap = make(map[uuid.UUID]*debuggerEvent)
	for _, event := range container.events {
		eventsMap[event.uid] = event
	}

	for _, history := range container.history {
		if event, ok := eventsMap[history.event.uid]; ok {
			event.Sent = true
		}
	}
}

func isFiltering(l *list.Model) bool {
	return l.FilterState() == list.Filtering
}
