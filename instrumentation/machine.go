package instrumentation

import (
	"context"
)

type QuantumMachine interface {
	Init(ctx context.Context, machineContext any) error

	SendEvent(ctx context.Context, event Event) error

	LazySendEvent(ctx context.Context, event Event) error

	LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error

	SnapshotExtractor
}
