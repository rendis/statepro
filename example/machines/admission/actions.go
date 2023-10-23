package admission

import (
	"context"
	"log"
)

func logEntryToStatus(ctx context.Context, realityName, universeName string) error {
	log.Printf("entry to reality '%s' from universe '%s'", realityName, universeName)
	return nil
}

func logExitFromStatus(ctx context.Context, realityName, universeName string) error {
	log.Printf("exit from reality '%s' from universe '%s'", realityName, universeName)
	return nil
}
