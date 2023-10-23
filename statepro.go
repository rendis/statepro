package statepro

import (
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/instrumentation"
)

func NewEvent(eventName string, data map[string]any, evtType instrumentation.EventType) instrumentation.Event {
	return experimental.NewEvent(eventName, data, evtType)
}

func NewEventBuilder(eventName string) instrumentation.EventBuilder {
	return experimental.NewEventBuilder(eventName)
}
