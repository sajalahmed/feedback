package main

import (
	"feedback-app/config"
	"feedback-app/db"
	"flag"
	"log"
)

func main() {
	force := flag.Int("force", -1, "Force migration version to clean dirty state")
	down := flag.Bool("down", false, "Revert all migrations (tear down database)")
	flag.Parse()

	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// 2. Handle Force Flag
	if *force > -1 {
		log.Printf("Forcing migration version to %d...", *force)
		if err := db.ForceVersion(cfg.DatabaseDSN, *force); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		log.Println("Force complete. You can now run the migration again.")
		return
	}

	// 3. Handle Down Flag
	if *down {
		log.Println("Reverting database migrations...")
		if err := db.RevertMigrations(cfg.DatabaseDSN); err != nil {
			log.Fatalf("Revert failed: %v", err)
		}
		log.Println("Revert finished successfully.")
		return
	}

	// 4. Run Migrations (Up)
	log.Println("Starting database migration...")
	if err := db.RunMigrations(cfg.DatabaseDSN); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration finished successfully.")
}
