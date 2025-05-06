package controllers

import (
	"net/http"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/slimuuid"
	"github.com/theMitocondria/Unbiass/awsHandler"
	"time"
	"strings"
	"fmt"
	"os"
	"github.com/golang-jwt/jwt/v5"
)

type result struct {
	StartTime time.Time `gorm:"column:start_time"`
	ContestName string `gorm:"column:contest_name"`
	EndTime time.Time `gorm:"column:end_time"`
	Duration uint32 `gorm:"column:duration"`
	CompanyName string `gorm:"column:company_name"`
}

func createStudent(fields []string , contestID string ,  StartTime time.Time , EndTime time.Time)(models.Student , error) {

	MAC_ADDRESS := os.Getenv("MAC_ADDRESS")
	STUDENT_SECRET := os.Getenv("STUDENT_SECRET")

	studentID , err := slimuuid.GenerateBest(MAC_ADDRESS)
	if err != nil {
		return models.Student{} , err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": studentID,
		"exp": EndTime,
		"rol":"student",
	})

	tokenString, err := token.SignedString([]byte(STUDENT_SECRET))
	
	if err != nil {
		return models.Student{}, err
	}
	var gender string
	if len(fields[2]) > 1 {
		switch strings.ToLower(fields[2]) {
		case "male":
			gender = "M"
		case "female":
			gender = "F"
		default:
			gender = "O"
		}
	} else {
		gender = fields[2]
	}

	student := models.Student{
		ID: studentID,
		Name: fields[0],
		Email: fields[1],
		Gender: gender,
		TestDate: StartTime.Truncate(24 * time.Hour),
		ContestID: contestID,
		Token: tokenString,
	}
	return student , nil
}

func IntToChar(n int) rune {
	switch {
	case n >= 0 && n < 26:
		return rune('A' + n)
	case n >= 26 && n < 52:
		return rune('a' + (n - 26))
	default:
		return rune('a')
	}
}

func sendMails(students []models.Student , result result , contestID string)[]error{
	recipients := []awsHandler.EmailRecipient{}
	position := strings.ReplaceAll(result.ContestName, " ", "-")
	hour := result.StartTime.Hour()
	minute:= result.StartTime.Minute()
	endHour := result.EndTime.Hour()
	endMinute:= result.EndTime.Minute()
	for _, student := range students {
		recipient := awsHandler.EmailRecipient{
			Email : student.Email,
			TemplateData: map[string]interface{}{
				"name": student.Name,
				"link" : fmt.Sprintf("www.unbiass.com/companies/%s-%s/students/%s%c%c/auth",result.CompanyName,position,student.ID,IntToChar(hour),IntToChar(minute),IntToChar(endHour),IntToChar(endMinute)),
				"duration" : result.Duration ,
				"companyName" : result.CompanyName,
				"startTime" : result.StartTime,
				"contestName" : result.ContestName,
			},
		}
		recipients = append(recipients,recipient)
	}

	err := awsHandler.SendBulk("online-assessment", "dhruvmehta382@gmail.com", recipients) 
	if len(err) > 0 {
		return err
	}

	return nil
}
// create students from csv file and add them to the database and send them mail from SES 
func CreateStudentsFromCSVFromContestIDAndSendMail(ctx *gin.Context) {
	// contest id milegi
	contestID := ctx.Param("contestid")

	var result result 
	if err := database.DB.Model(&models.Contest{}).Select("contests.start_time as start_time ,contests.end_time as end_time , contests.name as contest_name , contests.duration as duration , companies.name as company_name").Joins("INNER JOIN companies on contests.company_id = companies.id").Where("contests.id=?" , contestID).Scan(&result).Error ; err != nil {
			ctx.JSON(http.StatusInternalServerError , gin.H{"error":fmt.Sprintf("error in searching the contests %s" , err.Error())})
			return
		}
    // result.start_time ke hour or minute ko nikalna 
	// check if contest has started
	if time.Now().After(result.StartTime) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Contest is already started"})
		return
	}

	data , err := awsHandler.DownloadS3Object("unbiasss",contestID + ".csv")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var students []models.Student
	for _,line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, ",")
		if len(fields) > 2 {
			student , err := createStudent(fields, contestID , result.StartTime , result.EndTime)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError , gin.H{"error":err.Error()})
				return
			}		
			students = append(students, student)
		}
	}

	if err := database.DB.Model(models.Student{}).Create(&students).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// mails hav been sent but not recieved idk why
	errors := sendMails(students,result,contestID) 
	if len(errors) > 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errors})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Students created successfully"})
}

func AuthStudentInfo(ctx *gin.Context){
	studentID := ctx.Param("studentid")

	//check if there is a student like that
	var student models.Student 
	database.DB.Model(&models.Student{}).Select("token , id, contest_id").Where("id=?",studentID).First(&student)

	if student.ID == "" {
		ctx.JSON(http.StatusBadRequest,gin.H{"error":"Not a valid candidate"})
		return
	}

	ctx.SetCookie("Authorization", "Bearer "+ student.Token, 60*120, "/", "", false, true)  // for 2 hours 
	ctx.JSON(http.StatusOK , gin.H{
		"message":student.ContestID,
	})
}

// l

