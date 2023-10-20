package statepro

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
)

type QuantumMachine interface {
	Init(ctx context.Context, machineContext any) error

	SendEvent(ctx context.Context, event experimental.Event) error

	GetSnapshot() experimental.ExQuantumMachineSnapshot

	LoadSnapshot(snapshot experimental.ExQuantumMachineSnapshot) error
}
