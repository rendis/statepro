package instrumentation

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/rendis/statepro/v3/internal/util"
)

// SnapshotProvider is an optional capability for callers that need capture errors.
// QuantumMachine and executor interfaces remain compatible with existing implementations.
type SnapshotProvider interface {
	GetSnapshotWithError() (*MachineSnapshot, error)
}

// GetSnapshotWithError captures from a machine or synchronous action args exposing
// SnapshotProvider. It returns an error for providers without that capability.
func GetSnapshotWithError(source interface{ GetSnapshot() *MachineSnapshot }) (*MachineSnapshot, error) {
	if provider, ok := source.(SnapshotProvider); ok {
		return provider.GetSnapshotWithError()
	}
	return nil, errors.New("snapshot provider does not support capture errors")
}

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

// UnmarshalJSON preserves numeric values that cannot round-trip through float64.
func (ms *MachineSnapshot) UnmarshalJSON(data []byte) error {
	type snapshotJSON MachineSnapshot
	var decoded snapshotJSON
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	for _, snapshot := range decoded.Snapshots {
		util.NormalizeJSONNumbers(map[string]any(snapshot))
	}
	*ms = MachineSnapshot(decoded)
	return nil
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

	// SuperpositionUniversesFinalized is the map of the superposition universes that have been finalized
	// key: universe id, value: universe final reality
	// a universe is in superposition if:
	// - has been initialized
	// - in superposition
	// - finalized
	SuperpositionUniversesFinalized map[string]string `json:"superpositionUniversesFinalized,omitempty" bson:"superpositionUniversesFinalized,omitempty" xml:"superpositionUniversesFinalized,omitempty"`
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

func (ms *MachineSnapshot) AddSuperpositionUniverseFinalized(universeCanonicalName string, reality string) {
	if ms.Resume.SuperpositionUniversesFinalized == nil {
		ms.Resume.SuperpositionUniversesFinalized = make(map[string]string)
	}
	ms.Resume.SuperpositionUniversesFinalized[universeCanonicalName] = reality
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
