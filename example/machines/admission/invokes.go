package admission

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"
)

func notifyStatusChanged(ctx context.Context, quantumMachineContext any, event instrumentation.Event) {
	log.Println("Notify admission changed")
}
