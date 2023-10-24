package theoretical

// RealityType is the json representation of a reality type.
type RealityType string

const (
	// RealityTypeTransition represents a transitional or intermediate state.
	// This is a state through which the machine passes on its way to a final state.
	// It is used for processing or making decisions that lead to another state.
	RealityTypeTransition RealityType = "transition"

	// RealityTypeFinal represents a successful final state.
	// Once the machine reaches this state, it generally does not transition to any other state.
	// This state is used to indicate that an operation or task has successfully completed.
	RealityTypeFinal RealityType = "final"

	// RealityTypeUnsuccessfulFinal represents an unsuccessful final state.
	// The machine stops upon reaching this state, but the operation is not considered to have been successful.
	// This state is used to capture conditions where the task or operation fails to achieve the desired outcome.
	RealityTypeUnsuccessfulFinal RealityType = "unsuccessfulFinal"
)

func IsFinalState(realityType RealityType) bool {
	return realityType == RealityTypeFinal || realityType == RealityTypeUnsuccessfulFinal
}

// RealityModel is the json representation of a state of a universe.
type RealityModel struct {
	// ID is the id of the reality.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, underscore (_) and dash (-)
	// * must start with a letter
	// * min length: 1
	ID string `json:"id" bson:"id" xml:"id" yaml:"id"`

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
	Type RealityType `json:"type" bson:"type" xml:"type" yaml:"type"`

	// On is the list of transitions that are executed when an event is received and the reality is established.
	// Validations:
	// * required if Type is RealityTypeTransition
	// * ignored if Type belongs to final states (RealityTypeFinal or RealityTypeUnsuccessfulFinal)
	// * if not nil, each TransitionModel must be valid.
	// * over each group of transitions:
	//	- must have only one transition to a reality of the same universe (format: 'RealityModel.ID').
	//	- can have one or more transitions that point to other universes (format: 'UniverseModel.ID').
	//	- can have one or more transitions that point to realities from other universes (format: 'UniverseModel.ID:RealityModel.ID').
	On map[string][]*TransitionModel `json:"on" bson:"on" xml:"on" yaml:"on"`

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
	//	- an error will be returned
	// Validations:
	// * optional
	// * if not nil, each ActionModel must be valid.
	ExitActions []*ActionModel `json:"exitActions,omitempty" bson:"exitActions,omitempty" xml:"exitActions,omitempty" yaml:"exitActions,omitempty"`

	// Description is the description of the reality.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`
}
