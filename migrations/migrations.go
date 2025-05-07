package main

import (
	"github.com/theMitocondria/Unbiass/inits" 
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
)

func init(){
	inits.LoadEnv()
	database.DBInit()
}

func main(){
	database.DB.AutoMigrate(&models.Company{})
	database.DB.AutoMigrate(&models.Question{})
	database.DB.AutoMigrate(&models.Contest{})
	database.DB.AutoMigrate(&models.McqAnswer{})
	database.DB.AutoMigrate(&models.Submission{})
	database.DB.AutoMigrate(&models.Student{})
	database.DB.AutoMigrate(&models.Testcase{})
	database.DB.AutoMigrate(&models.RankedStudent{})
	database.DB.AutoMigrate(&models.Developer{})
	database.DB.AutoMigrate(&models.Feedback{})
	database.DB.AutoMigrate(&models.Bug{})
}