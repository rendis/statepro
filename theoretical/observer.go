package theoretical

// ObserverModel is the json representation of an observer.
type ObserverModel struct {
	// Src is the name of the observer to be executed.
	// Validations:
	// * required
	// * no white space
	// * only letters and numbers
	// * must start with a letter
	// * min length: 1
	Src string `json:"src,omitempty" bson:"src,omitempty" xml:"src,omitempty" yaml:"src,omitempty"`

	// Args is the map of arguments to be passed to the observer.
	// Validations:
	// * optional
	// * keys:
	//	- no white space
	//	- only letters and numbers
	//	- must start with a letter
	// * values:
	//	- can be string, number or boolean
	Args map[string]any `json:"args,omitempty" bson:"args,omitempty" xml:"args,omitempty" yaml:"args,omitempty"`

	// Description is the description of the observer.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`
}