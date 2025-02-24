package models

import (
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
	"os"
)
type Testcase struct {
	ID 						string		`gorm:"primaryKey"`
	Type      				string		`gorm:"not null;index:idx,composite"`
	/* here we have used composite key because we will first of all always search like give me all tc of qs which are 
	"s=>system || p=>pretests ||  g=>given" so the space overhead for making composite index is just 16 bytes uuid + 1 
	for Type string so mere 12kb for 10000 records of table so 1 record table would be 120 kb ~can be negligible considering
	100 bytes of data for a row is average with 2 uuid + 2 urls .
	*/
	QuestionID 				string		`gorm:"not null;index:idx,composite"`
	Question				Question	`gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (tc *Testcase) BeforeCreate(tx *gorm.DB) (err error) {
	tc.ID, err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}
