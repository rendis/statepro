package instrumentation

type EventType string

const (
	EventTypeStart   EventType = "Start"   // Event triggered when the universe starts
	EventTypeStartOn EventType = "StartOn" // Event triggered when the universe starts on a reality
	EventTypeAlways  EventType = "Always"  // Event triggered from reality from its "always transitions"
	EventTypeOn      EventType = "On"      // Event triggered from reality from its "on transitions"
	EventTypeOnEntry EventType = "OnEntry" // Event used to force the current reality to execute logic on entry
)

type Event interface {
	// GetEventName returns the Event name
	GetEventName() string

	// GetData returns the Event data
	GetData() map[string]any

	// DataContainsKey returns true if the Event data contains the given key
	DataContainsKey(key string) bool

	// GetEvtType returns the Event type
	GetEvtType() EventType

	// GetFlags returns the Event flags
	GetFlags() EventFlags
}

type EventBuilder interface {
	// SetData sets the Event data
	SetData(data map[string]any) EventBuilder

	// SetEvtType sets the Event type
	SetEvtType(evtType EventType) EventBuilder

	// SetFlags sets the Event flags
	SetFlags(flags EventFlags) EventBuilder

	// Build returns the Event
	Build() Event
}

type EventFlags struct {
	ReplayOnEntry bool
}
