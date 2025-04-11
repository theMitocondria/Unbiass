package models

import(
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
	"os"
)

type McqAnswer struct {
	ID 					string 		`gorm:"primaryKey"`
	Answer				string		`gorm:"not null" json:"answer"`
	QuestionID			string		`gorm:"not null" json:"question_id"`
	StudentID 			string		`gorm:"not null;index" json:"student_id"`
	Student 			Student		`gorm:"foreignKey:StudentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}


func (m *McqAnswer ) BeforeCreate(tx *gorm.DB) (err error) {	
	m.ID , err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}