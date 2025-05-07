package models

import(
	"os"
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
)

type Bug struct {
	ID string `gorm:"primaryKey"`
	Description string `json:"desc"`
}

func (bug *Bug) BeforeCreate(tx *gorm.DB)(err error){
	bug.ID , err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return 
}