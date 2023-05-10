package statepro

import (
	"fmt"
	"github.com/rendis/statepro/piece"
	"log"
)

func GetMachineById[ContextType any](machineId string, context *ContextType) (piece.ProMachine[ContextType], error) {
	if ptrPM, ok := proMachines[machineId]; ok {
		if pm, ok := ptrPM.(*piece.GMachine[ContextType]); ok {
			return piece.NewProMachine[ContextType](pm, context), nil
		}
	}
	return nil, fmt.Errorf("machine '%s' does not exist", machineId)
}

func InitMachines() {
	loadPropOnce.Do(func() {
		log.Print("Loading statepro properties")
		isPropLoaded = true
		loadXMachines()
		notifyXMachines()
		log.Print("Statepro properties loaded")
	})
}
