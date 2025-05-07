package models

import (
	"os"	
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"

)

type Company struct {
	ID          string      `gorm:"primaryKey" json:"id"`
	Name 	  	string		`gorm:"not null;size:50" json:"name"`
	Email 		string		`gorm:"not null;size:50;unique;index:idx_email" json:"email"`
	Password 	string		`gorm:"not null;size:100" json:"password"`
}

// creates unique id before creating the company
func (c *Company) BeforeCreate (tx *gorm.DB) (err error) {
	// sends mac_address stored in env to generate best uuid
	c.ID, err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return 
}
