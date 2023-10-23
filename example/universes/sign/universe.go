package sign

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
)

func NewSignUniverse() *SignUniverse {
	return &SignUniverse{}
}

type SignUniverse struct{}

func (a *SignUniverse) GetUniverseId() string {
	return "admission_default_sign_universe"
}

func (a *SignUniverse) GetUniverseDescription() string {
	return "default universe for sign handling"
}

func (a *SignUniverse) ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error) {
	//TODO implement me
	panic("implement me")
}

func (a *SignUniverse) ExecuteObserver(ctx context.Context, universeContext any, accumulatorStatistics experimental.AccumulatorStatistics, event experimental.Event, observer theoretical.ObserverModel) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a *SignUniverse) ExecuteAction(ctx context.Context, universeContext any, event experimental.Event, action theoretical.ActionModel) error {
	//TODO implement me
	panic("implement me")
}

func (a *SignUniverse) ExecuteInvoke(ctx context.Context, universeContext any, event experimental.Event, invoke theoretical.InvokeModel) {
	//TODO implement me
	panic("implement me")
}

func (a *SignUniverse) ExecuteCondition(ctx context.Context, conditionName string, args map[string]any, universeContext any, event experimental.Event) (bool, error) {
	//TODO implement me
	panic("implement me")
}
