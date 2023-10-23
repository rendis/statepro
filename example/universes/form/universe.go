package form

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
)

func NewFormUniverse() *FormUniverse {
	return &FormUniverse{}
}

type FormUniverse struct{}

func (a *FormUniverse) GetUniverseId() string {
	return "admission_default_form_universe"
}

func (a *FormUniverse) GetUniverseDescription() string {
	return "default universe for form handling"
}

func (a *FormUniverse) ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error) {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteObserver(ctx context.Context, universeContext any, accumulatorStatistics experimental.AccumulatorStatistics, event experimental.Event, observer theoretical.ObserverModel) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteAction(ctx context.Context, universeContext any, event experimental.Event, action theoretical.ActionModel) error {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteInvoke(ctx context.Context, universeContext any, event experimental.Event, invoke theoretical.InvokeModel) {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteCondition(ctx context.Context, conditionName string, args map[string]any, universeContext any, event experimental.Event) (bool, error) {
	//TODO implement me
	panic("implement me")
}
