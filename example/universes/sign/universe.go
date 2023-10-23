package sign

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
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
	return quantumMachineContext, nil
}

func (a *SignUniverse) ExecuteObserver(ctx context.Context, args experimental.ObserverExecutorArgs) (bool, error) {
	return false, nil
}

func (a *SignUniverse) ExecuteAction(ctx context.Context, args experimental.ActionExecutorArgs) error {
	return nil
}

func (a *SignUniverse) ExecuteInvoke(ctx context.Context, args experimental.InvokeExecutorArgs) {

}

func (a *SignUniverse) ExecuteCondition(ctx context.Context, args experimental.ConditionExecutorArgs) (bool, error) {
	return false, nil
}
