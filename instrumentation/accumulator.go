package instrumentation

// Accumulator allows to accumulate events in different realities for a given universe
type Accumulator interface {
	// Accumulate accumulates an Event in the given reality
	// Only events that exist in the RealityModel.On will be accumulated
	// Parameters:
	// 	- RealityName: the reality name
	// 	- Event: the Event to accumulate
	Accumulate(realityName string, event Event)

	// GetStatistics returns the Event accumulator statistics
	GetStatistics() AccumulatorStatistics

	// GetActiveRealities returns the realities that have accumulated events
	GetActiveRealities() []string
}

// AccumulatorStatistics allows to get statistics from an Event accumulator
type AccumulatorStatistics interface {
	// GetRealitiesEvents returns the accumulated events for each reality
	// Events can be repeated if they are accumulated in more than one reality
	// The map key is the reality name and the value is the accumulated events
	GetRealitiesEvents() map[string][]Event

	// GetRealityEvents returns the accumulated events for the given reality
	// Events can be repeated if they were received more than once
	GetRealityEvents(realityName string) []Event

	// GetAllEventsNames returns the accumulated events names for all realities (without repetitions)
	GetAllEventsNames() []string

	// CountAllEventsNames returns the number of accumulated events names for all realities (without repetitions)
	CountAllEventsNames() int
}
