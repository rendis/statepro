package theoretical

// UniverseModel represents an individual universe within a quantum machine (QuantumMachineModel).
// A universe has its own list of realities and can possess universal constants.
// It also has initial states that can be a mixture of realities and other universes.
// In addition, each universe has an optional version and description.
type UniverseModel struct {
	// ID is the id of the universe.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, underscore (_) and dash (-)
	// * must start with a letter
	// * min length: 1
	ID string `json:"id,omitempty" bson:"id,omitempty" xml:"id,omitempty" yaml:"id,omitempty"`

	// Initial is the initial reality of the universe.
	// Validations:
	// * required
	// * must be a key of the Realities map.
	Initial string `json:"initial,omitempty" bson:"initial,omitempty" xml:"initial,omitempty" yaml:"initial,omitempty"`

	// Realities is the list of Realities of the universe.
	// Validations:
	// * required
	// * size > 0
	// * keys must be the RealityModel.ID value.
	// * values can't be nil.
	// * each RealityModel must be valid.
	Realities map[string]*RealityModel `json:"realities,omitempty" bson:"realities,omitempty" xml:"realities,omitempty" yaml:"realities,omitempty"`

	// UniversalConstants is the list of universal constants of the universe.
	// Validations:
	// * optional
	// * if not nil, must be valid.
	UniversalConstants *UniversalConstantsModel `json:"universalConstants,omitempty" bson:"universalConstants,omitempty" xml:"universalConstants,omitempty" yaml:"universalConstants,omitempty"`

	// Description is the description of the universe.
	// Validations:
	// * optional
	Description *string `json:"description,omitempty" bson:"description,omitempty" xml:"description,omitempty" yaml:"description,omitempty"`

	// Version is the version of the universe.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, dot (.), underscore (_) and dash (-). Example: 1.0.0, v1.0.2, 1.0, v1 or 1.0.0-alpha.1
	Version string `json:"version,omitempty" bson:"version,omitempty" xml:"version,omitempty" yaml:"version,omitempty"`
}

// GetReality returns the reality with the given id.
func (u *UniverseModel) GetReality(id string) *RealityModel {
	return u.Realities[id]
}
