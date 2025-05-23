package models

import (
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
	"github.com/lib/pq"
	"os"	
)

type Question struct {
	ID					string			`gorm:"primaryKey" json:"id"`
	Name				string			`gorm:"not null;" json:"name"`
	Description 		string			`gorm:"not null;type:text" json:"description"`
	// "M" for MCQ , "C" for Code
	MCQ_OR_Code			string			`gorm:"not null;size:1" json:"mcq_or_code"`
	Tags 				pq.StringArray	`gorm:"not null;type:text[]"`
	Difficulty_rating 	uint32			`gorm:"not null;default:0;index" json:"difficulty_rating"`
	Language 			string			`gorm:"size:50" json:"language"`
	// if mcq then only here 4-5 options whatever and last option for answer of that (4 options , 5 one answer)
	MCQ_Options			pq.StringArray	`gorm:"type:text[]" json:"options"`
	ContestID 			string			`gorm:"index"  json:"contest_id"`
} 	

func (qs *Question) BeforeCreate (tx *gorm.DB) (err error) {
	qs.ID , err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}