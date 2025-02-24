package models

import (
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
	"os"
)

type RankedStudent struct {
	ID				string		`gorm:"primaryKey" json:"id"`
	StudentID		string		`gorm:"not null;" json:"student_id"`//yha se unksname submissions wagera fetch ho jaege
	ContestID		string		`gorm:"not null" json:"contest_id"`
	Contest 		Contest		`gorm:"foreignKey:ContestID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (rs *RankedStudent) BeforeCreate(tx *gorm.DB) (err error) {
	rs.ID, err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}