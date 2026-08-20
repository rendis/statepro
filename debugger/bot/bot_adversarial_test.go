package bot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rendis/statepro/v3/debugger/bot"
	"github.com/rendis/statepro/v3/instrumentation"
)

type hostileQM struct {
	MockQuantumMachine
	loadErr   error
	initErr   error
	sendErr   error
	sendCalls int
}

func (h *hostileQM) LoadSnapshot(s *instrumentation.MachineSnapshot, ctx any) error {
	if h.loadErr != nil {
		return h.loadErr
	}
	return h.MockQuantumMachine.LoadSnapshot(s, ctx)
}

func (h *hostileQM) Init(ctx context.Context, machineContext any) error {
	if h.initErr != nil {
		return h.initErr
	}
	return h.MockQuantumMachine.Init(ctx, machineContext)
}

func (h *hostileQM) SendEvent(ctx context.Context, event instrumentation.Event) (bool, error) {
	h.sendCalls++
	if h.sendErr != nil {
		return false, h.sendErr
	}
	return h.MockQuantumMachine.SendEvent(ctx, event)
}

func TestAdv_NewBot_RejectsNilDeps(t *testing.T) {
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		return nil, nil
	}
	if _, err := bot.NewBot(nil, provider, false); err == nil {
		t.Fatal("expected error for nil quantum machine")
	}
	qm := &MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}
	if _, err := bot.NewBot(qm, nil, false); err == nil {
		t.Fatal("expected error for nil event provider")
	}
}

func TestAdv_Bot_GetQuantumMachine(t *testing.T) {
	qm := &MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		return nil, nil
	}
	b, err := bot.NewBot(qm, provider, false)
	if err != nil {
		t.Fatalf("NewBot: %v", err)
	}
	if b.GetQuantumMachine() != qm {
		t.Fatal("GetQuantumMachine must return the injected machine")
	}
}

func TestAdv_Bot_ProviderError(t *testing.T) {
	qm := &MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		return nil, fmt.Errorf("provider boom")
	}
	b, err := bot.NewBot(qm, provider, false)
	if err != nil {
		t.Fatalf("NewBot: %v", err)
	}
	if err := b.Run(context.Background(), nil); err == nil {
		t.Fatal("expected provider error")
	}
}

func TestAdv_Bot_LoadSnapshotError(t *testing.T) {
	qm := &hostileQM{
		MockQuantumMachine: MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}},
		loadErr:            fmt.Errorf("load boom"),
	}
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		return &MockEvent{name: "handled"}, nil
	}
	b, err := bot.NewBot(qm, provider, false)
	if err != nil {
		t.Fatalf("NewBot: %v", err)
	}
	if err := b.Run(context.Background(), nil); err == nil {
		t.Fatal("expected LoadSnapshot error")
	}
}

func TestAdv_Bot_InitError(t *testing.T) {
	qm := &hostileQM{
		MockQuantumMachine: MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}},
		initErr:            fmt.Errorf("init boom"),
	}
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		return nil, nil
	}
	b, err := bot.NewBot(qm, provider, true)
	if err != nil {
		t.Fatalf("NewBot: %v", err)
	}
	if err := b.Run(context.Background(), nil); err == nil {
		t.Fatal("expected Init error")
	}
}

func TestAdv_Bot_SendEventError(t *testing.T) {
	qm := &hostileQM{
		MockQuantumMachine: MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}},
		sendErr:            fmt.Errorf("send boom"),
	}
	calls := 0
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		calls++
		if calls == 1 {
			return &MockEvent{name: "handled"}, nil
		}
		return nil, nil
	}
	b, err := bot.NewBot(qm, provider, false)
	if err != nil {
		t.Fatalf("NewBot: %v", err)
	}
	if err := b.Run(context.Background(), nil); err == nil {
		t.Fatal("expected SendEvent error")
	}
}

func TestAdv_Bot_HappyPathHistory(t *testing.T) {
	qm := &MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}
	events := []instrumentation.Event{&MockEvent{name: "handled"}, &MockEvent{name: "handled"}}
	provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		if len(events) == 0 {
			return nil, nil
		}
		e := events[0]
		events = events[1:]
		return e, nil
	}
	b, err := bot.NewBot(qm, provider, true)
	if err != nil {
		t.Fatalf("NewBot: %v", err)
	}
	if err := b.Run(context.Background(), "ctx"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := len(b.GetHistory()); got != 2 {
		t.Fatalf("history len=%d, want 2", got)
	}
}
