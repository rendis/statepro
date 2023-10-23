package admission

import (
	"context"
	"fmt"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
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

func (a *AdmissionUniverse) ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error) {
	//TODO implement me
	panic("implement me")
}

func (a *AdmissionUniverse) ExecuteObserver(ctx context.Context, universeContext any, accumulatorStatistics experimental.AccumulatorStatistics, event experimental.Event, observer theoretical.ObserverModel) (bool, error) {
	switch observer.Src {
	case "isAdmissionCompleted":
		return isAdmissionCompleted(ctx, accumulatorStatistics)
	default:
		log.Printf("ERROR: observer not found. Observer name: '%s'\n", observer.Src)
		return false, fmt.Errorf("observer not found. Observer name: '%s'", observer.Src)
	}
}

func (a *AdmissionUniverse) ExecuteAction(ctx context.Context, universeContext any, event experimental.Event, action theoretical.ActionModel) error {
	switch action.Src {
	case "logMultiUniverseTransition":
		return logMultiUniverseTransition(ctx)
	default:
		log.Printf("ERROR: action not found. Action name: '%s'\n", action.Src)
		return fmt.Errorf("action not found. Action name: '%s'", action.Src)
	}
}

func (a *AdmissionUniverse) ExecuteInvoke(ctx context.Context, universeContext any, event experimental.Event, invoke theoretical.InvokeModel) {
	//TODO implement me
	panic("implement me")
}

func (a *AdmissionUniverse) ExecuteCondition(ctx context.Context, conditionName string, args map[string]any, universeContext any, event experimental.Event) (bool, error) {
	//TODO implement me
	panic("implement me")
}
