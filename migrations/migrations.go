package main

import (
	"log"

	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/models"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	log.Println("Running migrations...")
	err = db.AutoMigrate(
		&models.Company{},
		&models.Question{},
		&models.Contest{},
		&models.McqAnswer{},
		&models.Submission{},
		&models.Student{},
		&models.Testcase{},
		&models.RankedStudent{},
		&models.Developer{},
		&models.Feedback{},
		&models.Bug{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations complete ✅")
}
