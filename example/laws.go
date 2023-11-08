package main

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"
)

// actions

func logMultiUniverseTransitionAction(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
	log.Println("initializing states in multiple universes (Payment, Signature and Form)")
	return nil
}

func logEntryToStatusAction(_ context.Context, args instrumentation.ActionExecutorArgs) error {
	realityName := args.GetRealityName()
	universeName := args.GetUniverseCanonicalName()
	log.Printf("entry to reality '%s' from universe '%s'", realityName, universeName)
	return nil
}

func logExitFromStatusAction(_ context.Context, args instrumentation.ActionExecutorArgs) error {
	realityName := args.GetRealityName()
	universeName := args.GetUniverseCanonicalName()
	log.Printf("exit from reality '%s' from universe '%s'", realityName, universeName)
	return nil
}

// invokes

func notifyStatusChangedInvk(ctx context.Context, args instrumentation.InvokeExecutorArgs) {
	log.Println("Notify admission changed")
}

func logTransitionInvk(ctx context.Context, args instrumentation.InvokeExecutorArgs) {
	log.Printf("log transition from reality '%s'", args.GetRealityName())
}

// observers

func isAdmissionCompletedObserver(_ context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
	realityName := args.GetRealityName()
	accumulator := args.GetAccumulatorStatistics()

	log.Printf("checking if admission on reality '%s' is completed with accumulator: %s\n", realityName, accumulator)

	events := accumulator.GetRealityEvents(realityName)

	var approvedEvents = []string{"sign", "fill"}
	for _, approvedEvent := range approvedEvents {
		if _, ok := events[approvedEvent]; !ok {
			return false, nil
		}
	}

	return true, nil
}
