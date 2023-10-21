package statepro

import (
	"encoding/json"
	"github.com/rendis/statepro/v3/theoretical"
)

// ----- theoretical.QuantumMachineModel Serializers/Deserializers -----

func DesFromMap(source map[string]any) (*theoretical.QuantumMachineModel, error) {
	jsonStr, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var resp theoretical.QuantumMachineModel
	if err := json.Unmarshal(jsonStr, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func DesFromBinary(b []byte) (*theoretical.QuantumMachineModel, error) {
	var resp theoretical.QuantumMachineModel
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
