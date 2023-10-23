package payment

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
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
	//TODO implement me
	panic("implement me")
}

func (a *PaymentUniverse) ExecuteObserver(ctx context.Context, universeContext any, accumulatorStatistics experimental.AccumulatorStatistics, event experimental.Event, observer theoretical.ObserverModel) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a *PaymentUniverse) ExecuteAction(ctx context.Context, universeContext any, event experimental.Event, action theoretical.ActionModel) error {
	//TODO implement me
	panic("implement me")
}

func (a *PaymentUniverse) ExecuteInvoke(ctx context.Context, universeContext any, event experimental.Event, invoke theoretical.InvokeModel) {
	//TODO implement me
	panic("implement me")
}

func (a *PaymentUniverse) ExecuteCondition(ctx context.Context, conditionName string, args map[string]any, universeContext any, event experimental.Event) (bool, error) {
	//TODO implement me
	panic("implement me")
}
