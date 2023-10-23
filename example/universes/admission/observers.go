package admission

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
	"log"
)

func isAdmissionCompleted(ctx context.Context, accumulator experimental.AccumulatorStatistics) (bool, error) {
	log.Printf("checking if admission is completed with accumulator: %s\n", accumulator)
	return false, nil
}
