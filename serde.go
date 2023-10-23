package statepro

import (
	"encoding/json"
	"github.com/rendis/statepro/v3/theoretical"
)

// ----- theoretical.QuantumMachineModel Serializers/Deserializers -----

func DeserializeQuantumMachineFromMap(source map[string]any) (*theoretical.QuantumMachineModel, error) {
	if source == nil {
		return nil, nil
	}

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

func DeserializeQuantumMachineFromBinary(b []byte) (*theoretical.QuantumMachineModel, error) {
	var resp theoretical.QuantumMachineModel
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func SerializeQuantumMachineToMap(source *theoretical.QuantumMachineModel) (map[string]any, error) {
	if source == nil {
		return nil, nil
	}

	jsonStr, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var resp map[string]any
	if err = json.Unmarshal(jsonStr, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func SerializeQuantumMachineToBinary(source *theoretical.QuantumMachineModel) ([]byte, error) {
	if source == nil {
		return nil, nil
	}
	return json.Marshal(source)
}
