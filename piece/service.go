package piece

import "fmt"

type TInvocation[ContextType any] func(ContextType, Event)

type InvocationResponse struct {
	Target *string
	Event  Event
	Err    error
}

type GService[ContextType any] struct {
	Id      *string // Mandatory
	Src     *string // Mandatory
	Inv     TInvocation[ContextType]
	OnDone  *GTransition[ContextType]
	OnError *GTransition[ContextType]
}

func (s *GService[ContextType]) invoke(c *ContextType, e Event) {
	go s.Inv(*c, e)
}

func CastToSrv[ContextType any](i any) (TInvocation[ContextType], error) {
	if f, ok := i.(func(ContextType, Event)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("service '%s' with wrong type. Expected: func(ContextType, Event)", i)
}
