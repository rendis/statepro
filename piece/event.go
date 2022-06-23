package piece

type EventType string

const (
	EventTypeTransitionalEvent EventType = "TransitionalEvent"
	EventTypeOnEntry           EventType = "OnEntry"
	EventTypeDoAfter           EventType = "DoAfter"
	EventTypeDoAlways          EventType = "DoAlways"
	EventTypeOnDone            EventType = "OnDone"
	EventTypeOnError           EventType = "OnError"
)

type GEvent struct {
	Name    string
	Data    any
	Err     error
	EvtType EventType
}
