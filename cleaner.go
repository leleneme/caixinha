package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func cleanerService(c *AppContext) {
	log.Println("[Cleaner] Started cleaner service")
	for {
		time.Sleep(15 * time.Minute)
		log.Println("[Cleaner] Looking for files to delete...")

		toDelete, err := c.getFilesToRemove()
		if err != nil {
			log.Printf("[Cleaner] Error while looking files to remove: %s\n", err)
			continue
		}

		if len(toDelete) == 0 {
			continue
		}

		log.Printf("[Cleaner] %d candidates\n", len(toDelete))

		for _, record := range toDelete {
			path := fmt.Sprintf("%s/%s", *c.StoragePath, record.Id)
			// Don't care if the file is already deleted from the filesystem
			_ = os.Remove(path)
			// and removeFileRecord shouldn't fail while a connection to the db is stabilized
			_ = c.removeFileRecord(record.Id)
		}
	}
}
