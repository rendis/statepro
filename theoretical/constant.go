package theoretical

// UniversalConstantsModel is the json representation of universal constants.
// Universal constants are operations that are always executed over any reality and will be executed before any other operation.
type UniversalConstantsModel struct {
	// EntryInvokes is the list of invokes that are always executed when the reality is established (Asynchronously).
	// Validations:
	// * optional
	// * if not nil, each InvokeModel must be valid.
	EntryInvokes []*InvokeModel `json:"entryInvokes,omitempty" bson:"entryInvokes,omitempty" xml:"entryInvokes,omitempty" yaml:"entryInvokes,omitempty"`

	// ExitInvokes is the list of invokes that are always executed when the reality is exited (Asynchronously).
	// Validations:
	// * optional
	// * if not nil, each InvokeModel must be valid.
	ExitInvokes []*InvokeModel `json:"exitInvokes,omitempty" bson:"exitInvokes,omitempty" xml:"exitInvokes,omitempty" yaml:"exitInvokes,omitempty"`

	// EntryActions is the list of actions that are always executed when the reality is established (Synchronously).
	// Validations:
	// * optional
	// * if not nil, each ActionModel must be valid.
	EntryActions []*ActionModel `json:"entryActions,omitempty" bson:"entryActions,omitempty" xml:"entryActions,omitempty" yaml:"entryActions,omitempty"`

	// ExitActions is the list of actions that are always executed when the reality is exited (Synchronously).
	// Validations:
	// * optional
	// * if not nil, each ActionModel must be valid.
	ExitActions []*ActionModel `json:"exitActions,omitempty" bson:"exitActions,omitempty" xml:"exitActions,omitempty" yaml:"exitActions,omitempty"`
}
