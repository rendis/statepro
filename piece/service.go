package piece

import "fmt"

type ServiceResponse struct {
	Response any
	Err      error
}

type TInvocation[T any] func(T, Event) ServiceResponse

type InvocationResponse struct {
	Target *string
	Event  Event
	Err    error
}

type GService[T any] struct {
	Id      *string // Mandatory
	Src     *string // Mandatory
	Inv     TInvocation[T]
	OnDone  *GTransition[T]
	OnError *GTransition[T]
}

func (s *GService[T]) invoke(c T, e Event, respCh chan<- *InvocationResponse) {
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
		target, err = s.error(c, evt)
	} else {
		evt.data = r.Response
		target, err = s.done(c, evt)
	}

	// TODO: Kill the service if respCh was closed
	select {
	case respCh <- &InvocationResponse{Target: target, Event: evt, Err: err}:
	default:
	}
}

func (s *GService[T]) done(c T, e Event) (*string, error) {
	if s.OnDone != nil {
		return s.OnDone.resolve(c, e)
	}
	return nil, nil
}

func (s *GService[T]) error(c T, e Event) (*string, error) {
	if s.OnError != nil {
		return s.OnError.resolve(c, e)
	}
	return nil, nil
}

func CastToSrv[T any](i any) (TInvocation[T], error) {
	if f, ok := i.(func(T, Event) ServiceResponse); ok {
		return f, nil
	}
	return nil, fmt.Errorf("service '%s' with wrong type", i)
}
