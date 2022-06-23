package piece

import "fmt"

type TInvocation[CTX any] func(CTX, GEvent) (any, error)

type GInvocationResponse struct {
	Target string
	Event  GEvent
}

type Service[CTX any] struct {
	Id         string
	Src        string
	Invocation *TInvocation[CTX]
	OnDone     *GTransition[CTX]
	OnError    *GTransition[CTX]
}

func (s *Service[CTX]) invoke(c CTX, e GEvent, respCh chan<- *GInvocationResponse) {
	evt := GEvent{
		Name:    "OnDone",
		Data:    e.Data,
		Err:     nil,
		EvtType: EventTypeOnDone,
	}

	var target string
	if s.Invocation != nil {
		resp, err := (*s.Invocation)(c, e)
		if err != nil {
			evt.Name = "OnError"
			evt.Err = err
			evt.EvtType = EventTypeOnError
			target = s.done(c, evt)
		} else {
			evt.Data = resp
			target = s.error(c, evt)
		}
	}

	// TODO: Kill the service if respCh was closed
	select {
	case respCh <- &GInvocationResponse{Target: target, Event: evt}:
		fmt.Printf("Service %s invoked\n", s.Id)
	default:
		fmt.Printf("Service %s Invocation response channel is full or closed\n", s.Id)
	}
}

func (s *Service[CTX]) done(c CTX, e GEvent) string {
	if s.OnDone != nil {
		return s.OnDone.resolve(c, e)
	}
	return ""
}

func (s *Service[CTX]) error(c CTX, e GEvent) string {
	if s.OnError != nil {
		return s.OnError.resolve(c, e)
	}
	return ""
}
