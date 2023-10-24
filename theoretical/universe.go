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
	ID string `json:"id" bson:"id" xml:"id" yaml:"id"`

	// CanonicalName serves as the unchanging identifier used to group multiple versions of this individual universe under a unified concept.
	// Different versions of the universe may have varying sets of realities, universal constants, or other attributes.
	// However, all versions share the same CanonicalName to signify that they are variations of the same foundational concept.
	// Validations:
	// * required
	// * no white space
	// * only letters, numbers, underscore (_) and dash (-)
	// * must start with a letter
	// * min length: 1
	CanonicalName string `json:"canonicalName" bson:"canonicalName" xml:"canonicalName" yaml:"canonicalName"`

	// Initial is the initial reality of the universe.
	// Validations:
	// * required
	// * must be a key of the Realities map.
	Initial string `json:"initial" bson:"initial" xml:"initial" yaml:"initial"`

	// Realities is the list of Realities of the universe.
	// Validations:
	// * required
	// * size > 0
	// * keys must be the RealityModel.ID value.
	// * values can't be nil.
	// * each RealityModel must be valid.
	Realities map[string]*RealityModel `json:"realities" bson:"realities" xml:"realities" yaml:"realities"`

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
	Version string `json:"version" bson:"version" xml:"version" yaml:"version"`
}

// GetReality returns the reality with the given id.
func (u *UniverseModel) GetReality(id string) *RealityModel {
	return u.Realities[id]
}
