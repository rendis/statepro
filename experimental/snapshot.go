package experimental

type ExUniverseSnapshot map[string]any

type ExQuantumMachineSnapshot struct {
	Resume    UniversesResume               `json:"resume" bson:"resume" xml:"resume"`
	Snapshots map[string]ExUniverseSnapshot `json:"snapshots,omitempty" bson:"snapshots,omitempty" xml:"snapshots,omitempty"`
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

func (s *ExQuantumMachineSnapshot) processUniverse(u *ExUniverse) {
	// add snapshot
	if s.Snapshots == nil {
		s.Snapshots = make(map[string]ExUniverseSnapshot)
	}
	s.Snapshots[u.id] = u.GetSnapshot()

	// just process initialized universes
	if !u.initialized {
		return
	}

	// add active universe
	if !u.inSuperposition && !u.isFinalReality {
		if s.Resume.ActiveUniverses == nil {
			s.Resume.ActiveUniverses = make(map[string]string)
		}
		s.Resume.ActiveUniverses[u.id] = *u.currentReality
	}

	// add finalized universe
	if !u.inSuperposition && u.isFinalReality {
		if s.Resume.FinalizedUniverses == nil {
			s.Resume.FinalizedUniverses = make(map[string]string)
		}
		s.Resume.FinalizedUniverses[u.id] = *u.currentReality
	}

	// add superposition universe
	if u.inSuperposition {
		if s.Resume.SuperpositionUniverses == nil {
			s.Resume.SuperpositionUniverses = make(map[string]string)
		}
		s.Resume.SuperpositionUniverses[u.id] = *u.realityBeforeSuperposition
	}
}
