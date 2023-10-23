package admission

import (
	"context"
	"log"
)

func logEntryToStatus(ctx context.Context) error {
	log.Println("entry to status")
	return nil
}

func logExitFromStatus(ctx context.Context) error {
	log.Println("exit from status")
	return nil
}
