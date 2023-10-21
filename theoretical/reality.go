package theoretical

// RealityType is the json representation of a reality type.
type RealityType string

const (
	RealityTypeFinal  RealityType = "final"
	RealityTypeNormal RealityType = "normal"
)

// RealityModel is the json representation of a state of a universe.
type RealityModel struct {
	// ID is the id of the reality.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, underscore (_) and dash (-)
	// * must start with a letter
	// * min length: 1
	ID string `json:"id,omitempty" bson:"id,omitempty" xml:"id,omitempty" yaml:"id,omitempty"`

	// Observers is the list of observers that are always executed when the reality receives an event.
	// Observers are executed only when the superposition state is active.
	// Validations:
	// * optional
	// * if not nil, each ObserverModel must be valid.
	// Observers execution:
	//   * are executed in parallel and synchronously.
	//   * first observer that returns true, stops the execution of the other observers.
	//   * if an observer fails and other observer returns false, the error is returned.
	//   * if an observer fails and other observer returns true, the error is ignored.
	Observers []*ObserverModel `json:"observers,omitempty" bson:"observers,omitempty" xml:"observers,omitempty" yaml:"observers,omitempty"`

	// Always is the list of transitions that are always executed when the reality is established.
	// If one of the transitions fails, the reality is not established and the error is returned.
	// Transitions are executed in the order they are defined.
	// First transition that returns true, stops the execution of the other transitions.
	// Validations:
	// * optional
	// * if not nil, each TransitionModel must be valid.
	Always []*TransitionModel `json:"always,omitempty" bson:"always,omitempty" xml:"always,omitempty" yaml:"always,omitempty"`

	// Type is the type of the reality.
	// Validations:
	// * required
	Type RealityType `json:"type,omitempty" bson:"type,omitempty" xml:"type,omitempty" yaml:"type,omitempty"`

	// On is the list of transitions that are executed when an event is received and the reality is established.
	// Validations:
	// * required if Type is RealityTypeNormal
	// * ignored if Type is RealityTypeFinal
	// * if not nil, each TransitionModel must be valid.
	// * over each group of transitions:
	//	- must have only one transition to a reality of the same universe (format: 'RealityModel.ID').
	//	- can have one or more transitions that point to other universes (format: 'UniverseModel.ID@UniverseModel.Version').
	//	- can have one or more transitions that point to realities from other universes (format: 'UniverseModel.ID@UniverseModel.Version:RealityModel.ID').
	On map[string][]*TransitionModel `json:"on,omitempty" bson:"on,omitempty" xml:"on,omitempty" yaml:"on,omitempty"`

	// EntryInvokes is the list of invokes that are executed when the reality is established (Asynchronously).
	// * EntryInvokes are executed after the EntryActions.
	// * Invokes are executed in parallel and asynchronously.
	// * There is no way to get the result of an invoke (triggered and forget).
	// * Nothing is guaranteed about the order of execution and the order of completion.
	// * The reality can be exited before the invokes are completed or started.
	// * If an invoke fails that does not affect the reality or the other invokes.
	// Validations:
	// * optional
	// * if not nil, each InvokeModel must be valid.
	EntryInvokes []*InvokeModel `json:"entryInvokes,omitempty" bson:"entryInvokes,omitempty" xml:"entryInvokes,omitempty" yaml:"entryInvokes,omitempty"`

	// ExitInvokes is the list of invokes that are executed when the reality is exited (Asynchronously).
	// * ExitInvokes are executed after the ExitActions.
	// * Invokes are executed in parallel and asynchronously.
	// * There is no way to get the result of an invoke (triggered and forget).
	// * Nothing is guaranteed about the order of execution and the order of completion.
	// * The reality can be exited before the invokes are completed or started.
	// * If an invoke fails that does not affect the reality or the other invokes.
	// Validations:
	// * optional
	// * if not nil, each InvokeModel must be valid.
	ExitInvokes []*InvokeModel `json:"exitInvokes,omitempty" bson:"exitInvokes,omitempty" xml:"exitInvokes,omitempty" yaml:"exitInvokes,omitempty"`

	// EntryActions is the list of actions that are executed when the reality is established (Synchronously).
	// * Actions are executed in the order they are defined.
	// * If an action fails:
	//	- the last reality is restored
	//	- the error is returned
	// Validations:
	// * optional
	// * if not nil, each ActionModel must be valid.
	EntryActions []*ActionModel `json:"entryActions,omitempty" bson:"entryActions,omitempty" xml:"entryActions,omitempty" yaml:"entryActions,omitempty"`

	// ExitActions is the list of actions that reality are executed when the reality is exited (Synchronously).
	// * Actions are executed in the order they are defined.
	// * If an action fails:
	//	- current reality is not changed
	//	- the error is returned
	// Validations:
	// * optional
	// * if not nil, each ActionModel must be valid.
	ExitActions []*ActionModel `json:"exitActions,omitempty" bson:"exitActions,omitempty" xml:"exitActions,omitempty" yaml:"exitActions,omitempty"`

	// Description is the description of the reality.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`
}
