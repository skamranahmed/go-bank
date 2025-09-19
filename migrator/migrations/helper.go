package migrations

import (
	"log"
	"path/filepath"
	"runtime"
)

func logMigrationStatus(message string) {
	_, file, _, ok := runtime.Caller(1) // caller of logMigrationStatus
	filename := "unknown"
	if ok {
		filename = filepath.Base(file)
	}
	log.Printf("%s: %s", message, filename)
}
