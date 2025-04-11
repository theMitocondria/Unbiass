package controllers
import (
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
)
func CreateRanklist(studentIds []string , contestID string)error{
	//yha ab ranklist aaagi studentids aagye
	var ranklistStudents []models.RankedStudent
	for _ , studentId := range studentIds {
		student := models.RankedStudent{
			StudentID : studentId ,
			ContestID : contestID ,
		}
		ranklistStudents = append(ranklistStudents , student)
	}

	if err := database.DB.Model(&models.RankedStudent{}).Create(&ranklistStudents).Error ; err != nil {
		return err
	}

	return nil

}