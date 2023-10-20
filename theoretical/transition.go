package theoretical

// TransitionModel is the json representation of a transition.
type TransitionModel struct {
	// Condition is the condition that allows the transition to be executed.
	// If nil, the transition is always executed.
	// Validations:
	// * optional
	// * if not nil, must be valid.
	Condition *ConditionModel `json:"condition,omitempty" bson:"condition,omitempty" xml:"condition,omitempty" yaml:"condition,omitempty"`

	// Targets is the list of targets of the transition.
	// Validations:
	// * required
	// * size > 0
	// * each target can be:
	// 	- one reality of the same universe: 'RealityModel.ID'.
	//  - mix of universe and realities from other universes.
	//    format:
	//   	+ UniverseModel.ID@UniverseModel.Version
	//  	+ UniverseModel.ID@UniverseModel.Version:RealityModel.ID
	Targets []string `json:"targets,omitempty" bson:"targets,omitempty" xml:"targets,omitempty" yaml:"targets,omitempty"`

	// Actions is the list of actions that are executed when the transition is executed (when the condition is true).
	// * Actions are executed in the order they are defined and synchronously.
	// * If an action fails:
	//	- transition is not executed
	//	- the error is returned
	// Validations:
	// * optional
	// * if not nil, each ActionModel must be valid.
	Actions []*ActionModel `json:"actions,omitempty" bson:"actions,omitempty" xml:"actions,omitempty" yaml:"actions,omitempty"`

	// Description is the description of the transition. Optional.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`
}
