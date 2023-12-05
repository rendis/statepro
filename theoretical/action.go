package theoretical

// ActionModel is the json representation of an action.
type ActionModel struct {
	// Src is the name of the action to be executed.
	// Validations:
	// * required
	// * no white space
	// * only letters and numbers
	// * must start with a letter
	// * min length: 1
	Src string `json:"src" bson:"src" xml:"src" yaml:"src"`

	// Args is the map of arguments to be passed to the action.
	// Validations:
	// * optional
	// * keys:
	//	- no white space
	//	- only letters and numbers
	//	- must start with a letter
	// * values:
	//	- can be string, number or boolean
	Args map[string]any `json:"args,omitempty" bson:"args,omitempty" xml:"args,omitempty" yaml:"args,omitempty"`

	// Description is the description of the action.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`

	// Metadata is the map of metadata to be passed to the action.
	// Validations:
	// * optional
	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty" xml:"metadata,omitempty" yaml:"metadata,omitempty"`
}
