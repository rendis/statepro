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
	qmId := laws.GetQuantumMachineId()
	qmVersion := laws.GetQuantumMachineVersion()
	uniqueId := buildUniqueId(qmId, qmVersion)

	// check if the quantum machine laws is already registered.
	if _, ok := quantumMachinesLaws[uniqueId]; ok {
		log.Printf("[ERROR] quantum machine laws already registered. Id: '%s', Version: '%s'", qmId, qmVersion)
		return ErrQuantumMachineLawsAlreadyRegistered
	}

	// register the quantum machine laws.
	quantumMachinesLaws[uniqueId] = laws
	return nil
}

// RegisterUniverseLaws registers a new universe laws.
func RegisterUniverseLaws(laws experimental.UniverseLaws) error {
	universeKeeperMtx.Lock()
	defer universeKeeperMtx.Unlock()

	// check if the universe laws is already registered.
	universeId := laws.GetUniverseId()
	universeVersion := laws.GetUniverseVersion()
	uniqueId := buildUniqueId(universeId, universeVersion)

	// check if the universe laws is already registered.
	if _, ok := universesLaws[uniqueId]; ok {
		log.Printf("[ERROR] universe laws already registered. Id: '%s', Version: '%s'", universeId, universeVersion)
		return ErrUniverseLawsAlreadyRegistered
	}

	// register the universe laws.
	universesLaws[uniqueId] = laws
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
