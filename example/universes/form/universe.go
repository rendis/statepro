package form

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
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

func (a *FormUniverse) ExecuteObserver(ctx context.Context, args experimental.ObserverExecutorArgs) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteAction(ctx context.Context, args experimental.ActionExecutorArgs) error {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteInvoke(ctx context.Context, args experimental.InvokeExecutorArgs) {
	//TODO implement me
	panic("implement me")
}

func (a *FormUniverse) ExecuteCondition(ctx context.Context, args experimental.ConditionExecutorArgs) (bool, error) {
	//TODO implement me
	panic("implement me")
}
