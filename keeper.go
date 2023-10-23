package statepro

import (
	"errors"
	"github.com/rendis/statepro/v3/experimental"
	"log"
	"sync"
)

// errors
var (
	ErrQuantumMachineLawsAlreadyRegistered = errors.New("quantum machine laws already registered")
	ErrUniverseLawsAlreadyRegistered       = errors.New("universe laws already registered")
)

var (
	quantumMachinesKeeperMtx sync.Mutex
	universeKeeperMtx        sync.Mutex
	quantumMachinesLaws      = map[string]experimental.QuantumMachineLaws{}
	universesLaws            = map[string]experimental.UniverseLaws{}
)

// RegisterQuantumMachineLaws registers a new quantum machine laws.
func RegisterQuantumMachineLaws(laws experimental.QuantumMachineLaws) error {
	quantumMachinesKeeperMtx.Lock()
	defer quantumMachinesKeeperMtx.Unlock()

	// check if the quantum machine laws is already registered.
	mId := laws.GetQuantumMachineId()
	if _, ok := quantumMachinesLaws[mId]; ok {
		log.Printf("[ERROR] quantum machine laws '%s' already registered", mId)
		return ErrQuantumMachineLawsAlreadyRegistered
	}

	// register the quantum machine laws.
	quantumMachinesLaws[mId] = laws
	return nil
}

// RegisterUniverseLaws registers a new universe laws.
func RegisterUniverseLaws(laws experimental.UniverseLaws) error {
	universeKeeperMtx.Lock()
	defer universeKeeperMtx.Unlock()

	// check if the universe laws is already registered.
	uId := laws.GetUniverseId()
	if _, ok := universesLaws[uId]; ok {
		log.Printf("[ERROR] universe laws '%s'already registered", uId)
		return ErrUniverseLawsAlreadyRegistered
	}

	// register the universe laws.
	universesLaws[uId] = laws
	return nil
}

// GetQuantumMachineLaws returns the quantum machine laws given the id and the version.
func GetQuantumMachineLaws(uniqueIdentifier string) experimental.QuantumMachineLaws {
	return quantumMachinesLaws[uniqueIdentifier]
}

// GetUniverseLaws returns the universe laws given the id and the version.
func GetUniverseLaws(uniqueIdentifier string) experimental.UniverseLaws {
	return universesLaws[uniqueIdentifier]
}
