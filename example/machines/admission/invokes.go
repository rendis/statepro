package admission

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"
)

func notifyStatusChanged(ctx context.Context, quantumMachineContext any, event instrumentation.Event) {
	log.Println("Notify admission changed")
}

func logTransition(ctx context.Context, args instrumentation.InvokeExecutorArgs) {
	log.Printf("log transition from reality '%s'", args.GetRealityName())
}
