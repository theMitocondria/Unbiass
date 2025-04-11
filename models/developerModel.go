package models

import (
	"os"
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
)
type Developer struct {
	ID	string 	`json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
}

// ( type of the object which can call ) (parameters kya lega) (return kya krta h)
func (dev *Developer)BeforeCreate(tx *gorm.DB)(err error){
	dev.ID  , err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return nil
}