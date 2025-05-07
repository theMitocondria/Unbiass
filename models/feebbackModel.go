package models

import(
	"os"
	"gorm.io/gorm"
	"github.com/theMitocondria/slimuuid"
)
type Response struct {
	Question string `json:"question"`
    Answer   string `json:"answer"`
}

type Feedback struct {
	ID        string     `gorm:"primaryKey"`
	Responses ResponseSlice `gorm:"type:jsonb"`
}
func (c *Feedback) BeforeCreate (tx *gorm.DB) (err error) {
	c.ID , err = slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		return err
	}
	return
}
