package theoretical

import (
	"encoding/json"
)

// BuildQuantumMachineModelFromJSONDefinition builds a QuantumMachineModel from a json definition.
func BuildQuantumMachineModelFromJSONDefinition(byteArr []byte) (*QuantumMachineModel, error) {
	var tm = &QuantumMachineModel{}
	err := json.Unmarshal(byteArr, tm)
	return tm, err
}

//// FlowStepType is the json representation of a flow step type.
//type FlowStepType string
//
//const (
//	FlowStepTypeState    FlowStepType = "state"    // Means that the step is a specific state.
//	FlowStepTypeWildcard FlowStepType = "wildcard" // Means that the step can be any state or states defined in the machine.
//)

//// Flow is the json representation of a flow.
//type Flow struct {
//	// Steps is the list of steps of the flow.
//	Steps []*FlowStep `json:"steps" bson:"steps" xml:"steps" yaml:"steps"`
//
//	// Description is the description of the flow.
//	Description *string `json:"description" bson:"description" xml:"description" yaml:"description"`
//}
//
//// FlowStep is the json representation of a flow step.
//type FlowStep struct {
//	// StateName is the name of the state.
//	StateName string `json:"stateName" bson:"stateName" xml:"stateName" yaml:"stateName"`
//
//	// Type is the type of the step. Default is state.
//	Type FlowStepType `json:"type" bson:"type" xml:"type" yaml:"type"`
//
//	// Required is a flag that indicates if the step is required or not. Default is true.
//	Required *bool `json:"required" bson:"required" xml:"required" yaml:"required"`
//}
