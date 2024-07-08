package cli

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

var (
	defaultEventsPath       = "./events.json"
	defaultSnapshotsPath    = "./snapshots.json"
	defaultStateMachinePath = "./state_machine.json"
	defaultSMContextPath    = "./context.json"
)

type debuggerOptions struct {
	eventsPath       string
	snapshotsPath    string
	stateMachinePath string
	smContextPath    string
	smContext        any
}

func NewStateMachineDebugger() *StateMachineDebugger {
	return &StateMachineDebugger{}
}

type StateMachineDebugger struct {
	opt debuggerOptions
}

func (d *StateMachineDebugger) Run(stateMachineContext any) {
	if !isPointerOrNil(stateMachineContext) {
		fmt.Println("state machine context must be a pointer or nil")
		os.Exit(1)
	}

	d.opt.smContext = stateMachineContext
	d.loadDefaultOptions()

	container, err := buildContainer(&d.opt)
	if err != nil {
		fmt.Printf("error building container: %v", err)
		os.Exit(1)
	}

	m := buildInitialModel(container)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error running program: %v", err)
		os.Exit(1)
	}
}

func (d *StateMachineDebugger) SetEventsPath(path string) {
	d.opt.eventsPath = path
}

func (d *StateMachineDebugger) SetSnapshotsPath(path string) {
	d.opt.snapshotsPath = path
}

func (d *StateMachineDebugger) SetStateMachinePath(path string) {
	d.opt.stateMachinePath = path
}

func (d *StateMachineDebugger) SetSMContextPath(path string) {
	d.opt.smContextPath = path
}

func (d *StateMachineDebugger) loadDefaultOptions() {
	if d.opt.eventsPath == "" {
		d.opt.eventsPath = defaultEventsPath
	}

	if d.opt.snapshotsPath == "" {
		d.opt.snapshotsPath = defaultSnapshotsPath
	}

	if d.opt.stateMachinePath == "" {
		d.opt.stateMachinePath = defaultStateMachinePath
	}

	if d.opt.smContextPath == "" {
		d.opt.smContextPath = defaultSMContextPath
	}
}
