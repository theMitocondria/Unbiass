package models

import (
	"time"
)
type Student struct {
	ID				string 		`gorm:"primaryKey" json:"id"`
	Name 			string		`gorm:"size:50" json:"name"`
	Email 			string		`gorm:"not null;size:50" json:"email"`
	// here "M" for male , "F" for female , "O" for others
	Gender 			string 		`gorm:"size:1" json:"gender"`
	Score 			uint32		`gorm:"default:0" json:"score"`
	Token 			string  	`gorm:"not null;size:500"`
	//because of composite key now 70k students being traversed now in 1.5ms almost 10ms to 1.5 ms, especially making this first index and hence only 4 bytes trade of but time almost cut to 4 times minimum
	TestDate		time.Time	`gorm:"type:date;not null;index:idx,composite" json:"date" time_format:"2006-01-02"`
	// if only making a student with contestID then 20k traversal takes 2.5 ms 
	ContestID 		string		`gorm:"not null;index:idx,composite" json:"contest_id"`
	Contest 		Contest     `gorm:"foreignKey:ContestID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}
/*
Current Scenario => as each student would have answered x/y mcq questions in a format like =>( 1=>a , 2=> b....... ) uncertain
					number that we dont know as company can have different number of mcq type type question .

Options =>	1. We make a new submissions object , then in that we have two keys , mcq or code and then in mcq based we would have
				answers to mcq type and in codes we would have an array of submmisons
*/

