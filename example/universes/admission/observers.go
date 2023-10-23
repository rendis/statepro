package admission

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"
)

func isAdmissionCompleted(ctx context.Context, realityName string, accumulator instrumentation.AccumulatorStatistics) (bool, error) {
	log.Printf("checking if admission on reality '%s' is completed with accumulator: %s\n", realityName, accumulator)

	events := accumulator.GetRealityEvents(realityName)

	return len(events) > 1, nil
}
