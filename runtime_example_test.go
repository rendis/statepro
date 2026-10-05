package statepro_test

import (
	"context"
	"fmt"
	"time"

	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// Application budgets are examples; choose them for the application's workload.
func ExampleNewQuantumMachineWithOptions() {
	if err := builtin.RegisterObserver("example:allow", func(ctx context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		panic(err)
	}
	if err := builtin.RegisterAction("example:capture", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Action arguments capture safely without reentering the owner's lock.
		_, err := instrumentation.GetSnapshotWithError(args)
		return err
	}); err != nil {
		panic(err)
	}
	if err := builtin.RegisterInvoke("example:wait", func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) {
		// Replace this with work that observes ctx.Done() or passes ctx to I/O.
		<-ctx.Done()
	}); err != nil {
		panic(err)
	}
	initial := "idle"
	machine, err := statepro.NewQuantumMachineWithOptions(&theoretical.QuantumMachineModel{
		ID: "example", Initials: []string{"U:main"}, Universes: map[string]*theoretical.UniverseModel{
			"main": {ID: "main", CanonicalName: "main", Initial: &initial, Realities: map[string]*theoretical.RealityModel{
				"idle": {ID: "idle", Type: theoretical.RealityTypeTransition, Observers: []*theoretical.ObserverModel{{Src: "example:allow"}}, EntryActions: []*theoretical.ActionModel{{Src: "example:capture"}}, EntryInvokes: []*theoretical.InvokeModel{{Src: "example:wait"}}, On: map[string][]*theoretical.TransitionModel{"GO": {{Targets: []string{"done"}}}}},
				"done": {ID: "done", Type: theoretical.RealityTypeFinal},
			}},
		},
	}, instrumentation.RuntimeOptions{
		MaxConcurrentInvokes: 8, CancelInvokesOnExit: true,
		MaxAccumulatedEvents: 128, MaxTrackingEntries: 64, StrictObservers: true,
	})
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := machine.Init(ctx, nil); err != nil {
		panic(err)
	}
	if _, err := machine.SendEvent(ctx, statepro.NewEventBuilder("GO").Build()); err != nil {
		panic(err)
	}
	snapshot, err := instrumentation.GetSnapshotContext(ctx, machine)
	if err != nil {
		panic(err)
	}
	fmt.Println(snapshot.Resume.FinalizedUniverses["main"])
	// Use a fresh deadline for shutdown even if the event context is canceled.
	lifecycle := machine.(instrumentation.RuntimeLifecycle)
	if err := lifecycle.Close(); err != nil {
		panic(err)
	}
	shutdownCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := lifecycle.WaitInvokes(shutdownCtx); err != nil {
		panic(err)
	}
	// Output: done
}
