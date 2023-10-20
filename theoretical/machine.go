package theoretical

// QuantumMachineModel represents a quantum machine that serves as a container for multiple universes (UniverseModel).
// It defines its own initial states and universal constants, and may have an optional version and description.
type QuantumMachineModel struct {
	// ID is the id of the machine.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, underscore (_) and dash (-)
	// * must start with a letter
	// * min length: 1
	ID string `json:"id,omitempty" bson:"id,omitempty" xml:"id,omitempty" yaml:"id,omitempty"`

	// Universes is the list of universes of the machine.
	// Validations:
	// * required
	// * size > 0
	// * keys must be the UniverseModel.ID value.
	// * values can't be nil.
	// * each UniverseModel must be valid.
	Universes map[string]*UniverseModel `json:"universes,omitempty" bson:"universes,omitempty" xml:"universes,omitempty" yaml:"universes,omitempty"`

	// Initials is the list of initials realities or universes of the machine.
	// Validations:
	// * required
	// * size > 0
	// * initial can be:
	//	- mix of realities and universes
	//	- only one reality per universe. To reference a reality from another universe
	//    the format is 'UniverseModel.ID@UniverseModel.Version:RealityModel.ID'.
	//	- one or more universes. To reference a universe the format is 'UniverseModel.ID@UniverseModel.Version'.
	Initials []string `json:"initials,omitempty" bson:"initials,omitempty" xml:"initials,omitempty" yaml:"initials,omitempty"`

	// UniversalConstants is the list of universal constants of the machine.
	// Validations:
	// * optional
	// * if not nil, UniversalConstants must be valid.
	UniversalConstants *UniversalConstantsModel `json:"universalConstants,omitempty" bson:"universalConstants,omitempty" xml:"universalConstants,omitempty" yaml:"universalConstants,omitempty"`

	// Description is the description of the machine.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`

	// Version is the version of the machine.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, dot (.), underscore (_) and dash (-). Example: 1.0.0, v1.0.2, 1.0, v1 or 1.0.0-alpha.1
	Version string `json:"version,omitempty" bson:"version,omitempty" xml:"version,omitempty" yaml:"version,omitempty"`
}
