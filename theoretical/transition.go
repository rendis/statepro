package theoretical

type TransitionType string

const (
	// TransitionTypeDefault is the default transition type. It is used when no type is specified.
	// If the condition is true, the reality that triggers it is abandoned
	// This type of transition can target all types of realities (internal and external).
	TransitionTypeDefault TransitionType = "default"

	// TransitionTypeNotify is the notify transition type. It is used to notify the target realities.
	// If the condition is true, the reality that triggers it is NOT abandoned.
	// This type of transition can only target external universes or realities.
	TransitionTypeNotify TransitionType = "notify"
)

// TransitionModel is the json representation of a transition.
type TransitionModel struct {
	// Condition is the condition that allows the transition to be executed.
	// If nil, the transition is always executed.
	// Validations:
	// * optional
	// * if not nil, must be valid.
	Condition *ConditionModel `json:"condition,omitempty" bson:"condition,omitempty" xml:"condition,omitempty" yaml:"condition,omitempty"`

	// Type is the type of the transition.
	// Validations:
	// * optional
	// * if not nil, must be valid.
	// * if nil, TransitionTypeDefault is used.
	Type *TransitionType `json:"type,omitempty" bson:"type,omitempty" xml:"type,omitempty" yaml:"type,omitempty"`

	// Targets is the list of targets of the transition.
	// Validations:
	// * required
	// * size > 0
	// * each target can be:
	// 	- one reality of the same universe: 'RealityModel.ID'.
	//  - mix of universe and realities from other universes.
	//    format:
	//   	+ UniverseModel.ID
	//  	+ UniverseModel.ID:RealityModel.ID
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

	// Invokes is the list of invocations that are executed when the transition is executed (when the condition is true).
	// * Invocations are executed asynchronously.
	// Validations:
	// * optional
	// * if not nil, each InvokeModel must be valid.
	Invokes []*InvokeModel `json:"invokes,omitempty" bson:"invokes,omitempty" xml:"invokes,omitempty" yaml:"invokes,omitempty"`

	// Description is the description of the transition. Optional.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`
}

func (t *TransitionModel) IsNotification() bool {
	return t.Type != nil && *t.Type == TransitionTypeNotify
}
