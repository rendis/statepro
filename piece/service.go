package piece

import "fmt"

type ServiceResponse struct {
	Response any
	Err      error
}

type TInvocation[ContextType any] func(ContextType, Event) ServiceResponse

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

func (s *GService[ContextType]) invoke(c ContextType, e Event, at ActionTool[ContextType], respCh chan<- *InvocationResponse) {
	evt := &GEvent{
		name:    "OnDone",
		err:     nil,
		evtType: EventTypeOnDone,
	}

	var target *string
	var err error
	r := s.Inv(c, e)
	if r.Err != nil {
		evt.name = "OnError"
		evt.err = r.Err
		evt.evtType = EventTypeOnError
		target, err = s.error(c, evt, at)
	} else {
		evt.data = r.Response
		target, err = s.done(c, evt, at)
	}

	// TODO: Kill the service if respCh was closed
	select {
	case respCh <- &InvocationResponse{Target: target, Event: evt, Err: err}:
	default:
	}
}

func (s *GService[ContextType]) done(c ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	if s.OnDone != nil {
		return s.OnDone.resolve(c, e, at)
	}
	return nil, nil
}

func (s *GService[ContextType]) error(c ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	if s.OnError != nil {
		return s.OnError.resolve(c, e, at)
	}
	return nil, nil
}

func CastToSrv[ContextType any](i any) (TInvocation[ContextType], error) {
	if f, ok := i.(func(ContextType, Event) ServiceResponse); ok {
		return f, nil
	}
	return nil, fmt.Errorf("service '%s' with wrong type", i)
}
