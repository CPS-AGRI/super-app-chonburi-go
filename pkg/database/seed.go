package database

import (
	"log"
)

// Seed handles initial data population. 
// Now emptied as requested because data management is handled via API.
func Seed() {
	if DB == nil {
		return
	}

	log.Println("✅ Database is ready (Skipping auto-seeding as requested)")
}
