package contract

import (
	"context"
	"fmt"
	"github.com/rendis/statepro/v3/example/domain"
	"github.com/rendis/statepro/v3/experimental"
)

func NewContractUniverse() experimental.UniverseLaws {
	return &ContractUniverse{}
}

type ContractUniverse struct{}

func (u *ContractUniverse) GetUniverseId() string {
	return "1"
}

func (u *ContractUniverse) GetUniverseVersion() string {
	return "0.1"
}

func (u *ContractUniverse) GetUniverseDescription() string {
	return "ContractUniverse"
}

func (u *ContractUniverse) ExtractObservableKnowledge(machineContext any) (any, error) {
	switch c := machineContext.(type) {
	case *domain.AdmissionQMContext:
		return c.Contract, nil
	default:
		return nil, fmt.Errorf("unknown machine context type: %T", machineContext)
	}
}

func (u *ContractUniverse) ExecuteObserver(
	ctx context.Context,
	observerName string,
	args map[string]any,
	universeContext any,
	event experimental.Event,
	accumulatorStatistics experimental.AccumulatorStatistics,
) (bool, error) {
	return false, nil
}

func (u *ContractUniverse) ExecuteAction(
	ctx context.Context,
	actionName string,
	args map[string]any,
	universeContext any,
	event experimental.Event,
) error {
	return nil
}

func (u *ContractUniverse) ExecuteInvoke(
	ctx context.Context,
	invokeName string,
	args map[string]any,
	universeContext any,
	event experimental.Event,
) {
}

func (u *ContractUniverse) ExecuteCondition(
	ctx context.Context,
	conditionName string,
	args map[string]any,
	universeContext any,
	event experimental.Event,
) (bool, error) {
	return false, nil
}
