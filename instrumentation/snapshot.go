package instrumentation

import "encoding/json"

type SerializedUniverseSnapshot map[string]any

type MachineSnapshot struct {
	// Resume is the resume of the machine
	Resume UniversesResume `json:"resume" bson:"resume" xml:"resume"`

	// Snapshots is the map of the universe snapshots
	// key: universe id, value: universe snapshot
	Snapshots map[string]SerializedUniverseSnapshot `json:"snapshots,omitempty" bson:"snapshots,omitempty" xml:"snapshots,omitempty"`

	// Tracking is the map of the universe status tracking
	// key: universe id, value: list of states the universe has been through
	Tracking map[string][]string `json:"tracking,omitempty" bson:"tracking,omitempty" xml:"tracking,omitempty"`
}

type UniversesResume struct {
	// ActiveUniverses is the map of the active universes
	// key: universe id, value: universe current reality
	// a universe is active if:
	// - has been initialized
	// - not in superposition
	// - not finalized
	ActiveUniverses map[string]string `json:"activeUniverses,omitempty" bson:"activeUniverses,omitempty" xml:"activeUniverses,omitempty"`

	// FinalizedUniverses is the map of the finalized universes
	// key: universe id, value: universe final reality
	// a universe is finalized if:
	// - has been initialized
	// - not in superposition
	// - finalized
	FinalizedUniverses map[string]string `json:"finalizedUniverses,omitempty" bson:"finalizedUniverses,omitempty" xml:"finalizedUniverses,omitempty"`

	// SuperpositionUniverses is the map of the superposition universes
	// key: universe id, value: last reality before superposition
	// a universe is in superposition if:
	// - has been initialized
	// - in superposition
	SuperpositionUniverses map[string]string `json:"superpositionUniverses,omitempty" bson:"superpositionUniverses,omitempty" xml:"superpositionUniverses,omitempty"`
}

func (ms *MachineSnapshot) AddActiveUniverse(universeCanonicalName string, reality string) {
	if ms.Resume.ActiveUniverses == nil {
		ms.Resume.ActiveUniverses = make(map[string]string)
	}
	ms.Resume.ActiveUniverses[universeCanonicalName] = reality
}

func (ms *MachineSnapshot) AddFinalizedUniverse(universeCanonicalName string, reality string) {
	if ms.Resume.FinalizedUniverses == nil {
		ms.Resume.FinalizedUniverses = make(map[string]string)
	}
	ms.Resume.FinalizedUniverses[universeCanonicalName] = reality
}

func (ms *MachineSnapshot) AddSuperpositionUniverse(universeCanonicalName string, reality string) {
	if ms.Resume.SuperpositionUniverses == nil {
		ms.Resume.SuperpositionUniverses = make(map[string]string)
	}
	ms.Resume.SuperpositionUniverses[universeCanonicalName] = reality
}

func (ms *MachineSnapshot) AddUniverseSnapshot(universeId string, snapshot SerializedUniverseSnapshot) {
	if ms.Snapshots == nil {
		ms.Snapshots = make(map[string]SerializedUniverseSnapshot)
	}
	ms.Snapshots[universeId] = snapshot
}

func (ms *MachineSnapshot) AddTracking(universeId string, tracking []string) {
	if ms.Tracking == nil {
		ms.Tracking = make(map[string][]string)
	}
	ms.Tracking[universeId] = tracking
}

func (ms *MachineSnapshot) GetResume() UniversesResume {
	return ms.Resume
}

func (ms *MachineSnapshot) GetActiveUniverses() map[string]string {
	return ms.Resume.ActiveUniverses
}

func (ms *MachineSnapshot) GetFinalizedUniverses() map[string]string {
	return ms.Resume.FinalizedUniverses
}

func (ms *MachineSnapshot) GetSuperpositionUniverses() map[string]string {
	return ms.Resume.SuperpositionUniverses
}

func (ms *MachineSnapshot) GetTracking() map[string][]string {
	return ms.Tracking
}

func (ms *MachineSnapshot) ToJson() (string, error) {
	b, err := json.Marshal(ms)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
