package payment

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
)

func NewPaymentUniverse() *PaymentUniverse {
	return &PaymentUniverse{}
}

type PaymentUniverse struct{}

func (a *PaymentUniverse) GetUniverseId() string {
	return "admission_default_payment_universe"
}

func (a *PaymentUniverse) GetUniverseDescription() string {
	return "default universe for payment handling"
}

func (a *PaymentUniverse) ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error) {
	return quantumMachineContext, nil
}

func (a *PaymentUniverse) ExecuteObserver(ctx context.Context, args experimental.ObserverExecutorArgs) (bool, error) {
	return false, nil
}

func (a *PaymentUniverse) ExecuteAction(ctx context.Context, args experimental.ActionExecutorArgs) error {
	return nil
}

func (a *PaymentUniverse) ExecuteInvoke(ctx context.Context, args experimental.InvokeExecutorArgs) {
}

func (a *PaymentUniverse) ExecuteCondition(ctx context.Context, args experimental.ConditionExecutorArgs) (bool, error) {
	return false, nil
}
