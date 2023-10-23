package instrumentation

type SerializedUniverseSnapshot map[string]any

type MachineSnapshot struct {
	Resume    UniversesResume                       `json:"resume" bson:"resume" xml:"resume"`
	Snapshots map[string]SerializedUniverseSnapshot `json:"snapshots,omitempty" bson:"snapshots,omitempty" xml:"snapshots,omitempty"`
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

func (ms *MachineSnapshot) AddActiveUniverse(universeId string, reality string) {
	if ms.Resume.ActiveUniverses == nil {
		ms.Resume.ActiveUniverses = make(map[string]string)
	}
	ms.Resume.ActiveUniverses[universeId] = reality
}

func (ms *MachineSnapshot) AddFinalizedUniverse(universeId string, reality string) {
	if ms.Resume.FinalizedUniverses == nil {
		ms.Resume.FinalizedUniverses = make(map[string]string)
	}
	ms.Resume.FinalizedUniverses[universeId] = reality
}

func (ms *MachineSnapshot) AddSuperpositionUniverse(universeId string, reality string) {
	if ms.Resume.SuperpositionUniverses == nil {
		ms.Resume.SuperpositionUniverses = make(map[string]string)
	}
	ms.Resume.SuperpositionUniverses[universeId] = reality
}

func (ms *MachineSnapshot) AddUniverseSnapshot(universeId string, snapshot SerializedUniverseSnapshot) {
	if ms.Snapshots == nil {
		ms.Snapshots = make(map[string]SerializedUniverseSnapshot)
	}
	ms.Snapshots[universeId] = snapshot
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
