package builtin

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
	"sync"
	"testing"
)

func TestRevision_RegistroConcurrenteDeTodosLosExecutors(t *testing.T) {
	observer := func(context.Context, instrumentation.ObserverExecutorArgs) (bool, error) { return true, nil }
	action := func(context.Context, instrumentation.ActionExecutorArgs) error { return nil }
	invoke := func(context.Context, instrumentation.InvokeExecutorArgs) {}
	condition := func(context.Context, instrumentation.ConditionExecutorArgs) (bool, error) { return true, nil }
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				if err := RegisterObserver("test:review-observer", observer); err != nil {
					t.Error(err)
				}
				if err := RegisterAction("test:review-action", action); err != nil {
					t.Error(err)
				}
				if err := RegisterInvoke("test:review-invoke", invoke); err != nil {
					t.Error(err)
				}
				if err := RegisterCondition("test:review-condition", condition); err != nil {
					t.Error(err)
				}
				if GetObserver("test:review-observer") == nil || GetAction("test:review-action") == nil || GetInvoke("test:review-invoke") == nil || GetCondition("test:review-condition") == nil {
					t.Error("executor registrado no encontrado")
				}
			}
		}()
	}
	wg.Wait()
}
