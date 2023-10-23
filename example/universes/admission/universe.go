package admission

import (
	"context"
	"fmt"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"
)

func NewAdmissionUniverse() *AdmissionUniverse {
	return &AdmissionUniverse{}
}

type AdmissionUniverse struct{}

func (a *AdmissionUniverse) GetUniverseId() string {
	return "admission_default_universe"
}

func (a *AdmissionUniverse) GetUniverseDescription() string {
	//TODO implement me
	panic("implement me")
}

func (a *AdmissionUniverse) ExecuteObserver(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
	observer := args.GetObserver()
	accumulatorStatistics := args.GetAccumulatorStatistics()

	switch observer.Src {
	case "isAdmissionCompleted":
		return isAdmissionCompleted(ctx, args.GetRealityName(), accumulatorStatistics)
	default:
		log.Printf("ERROR: observer not found. Observer name: '%s'\n", observer.Src)
		return false, fmt.Errorf("observer not found. Observer name: '%s'", observer.Src)
	}
}

func (a *AdmissionUniverse) ExecuteAction(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	action := args.GetAction()

	switch action.Src {
	case "logMultiUniverseTransition":
		return logMultiUniverseTransition(ctx)
	default:
		return nil
	}
}
