package admission

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
	"log"
)

func notifyStatusChanged(ctx context.Context, quantumMachineContext any, event experimental.Event) {
	log.Println("Notify admission changed")
}
