package models

import(
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
	"time"
	"os"
)

type Submission struct {
	ID 					string 		`gorm:"primaryKey"`
	Time 				time.Time	`gorm:"not null" json:"time"`
	Verdict 			string		`gorm:"not null" json:"verdict"`
	Language 			string		`gorm:"not null" json:"language"`
	CodeFileLink		string		`gorm:"not null" json:"code_file_link"`
	/*
	Here we are using single indexed because space overhead is lot like 16 bytes for single indexed
	vs 32 bytes for composite and also submissions would require to be fast update and insert based and even
	in 10k records almost doubled size of index table and almost 0.03ms difference in response time.
	*/
	StudentID 			string		`gorm:"not null"`
	QuestionID 			string		`gorm:"not null;"`
	Student 			Student		`gorm:"foreignKey:StudentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (s *Submission) BeforeCreate(tx *gorm.DB) (err error) {	
	
	s.ID, err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}

/*
Approachs => 1. make a composite key with questionId and studentId and index it or not ?
			Getting response in 0.085 ms for 5000 records. without index => time is 1.85 ms
		   	2. make a studentId index and max 2 questions would be there and after getting studentis's sol there will be only max 10 submissions .
			Approx 0.125 ms for 5000 records
*/




//pehla goal => btree vs hash dekho