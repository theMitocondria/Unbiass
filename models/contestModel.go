package models

import (	
	"gorm.io/gorm"
	"time"
	"github.com/theMitocondria/slimuuid"
	"os"
)
// Contest represents a contest model
type Contest struct {
	ID          string      `gorm:"primaryKey"`
	Name 	  	string  	`gorm:"not null;size:50" json:"name"`
	Description string		`gorm:"not null" json:"description"`
	MCQ         bool 		`gorm:"default:false" json:"mcq"`
	Code 		bool 		`gorm:"default:false" json:"code"`
	StudentsRequired uint32	`gorm:"not null" json:"students_required"`
	StartTime 	time.Time	`gorm:"not null"json:"start_time"`
	EndTime 	time.Time	`gorm:"not null"json:"end_time"`
	// Duration in minutes
	Duration 	uint32		`gorm:""not null" json:"duration"`
	CompanyID 	string		`gorm:"not null" json:"company_id"`	
	Company 	Company     `gorm:"foreignKey:CompanyID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate is a gorm hook that is called before creating a new record
func (c *Contest) BeforeCreate (tx *gorm.DB) (err error) {
	c.ID , err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}

