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
type slimQuestion struct {
	ID       string `gorm:"column:id"`
	McqOrCode string `gorm:"column:mcq_or_code"`
}
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
		case "Male":
			gender = "M"
		case "Female":
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
				"link" : fmt.Sprintf("https://unbiass.com/app/companies/%s-%s/students/%s%c%c%c%c/auth",result.CompanyName,position,student.ID,IntToChar(hour),IntToChar(minute),IntToChar(endHour),IntToChar(endMinute)),
				"duration" : result.Duration ,
				"companyName" : result.CompanyName,
				"startTime" : result.StartTime,
				"contestName" : result.ContestName,
			},
		}
		fmt.Println("recipient", recipient)
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
	database.DB.Model(&models.Student{}).Select("token , id, contest_id").Where("id=? and has_logged_in_once=false",studentID).First(&student)

	if student.ID == "" {
		ctx.JSON(http.StatusBadRequest,gin.H{"error":"Not a valid candidate or you have loggedin before"})
		return
	}

	// if err := database.DB.Model(&models.Student{}).Where("id=?",student.ID).Update("has_logged_in_once",true).Error; err != nil {
	// 	ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	// 	return
	// }// now check if the conntest has started and not ended yet , then fetch questions from the contest in a sequence first all mcq then all coding questions
	var contest models.Contest
	if err := database.DB.Model(&models.Contest{}).Select("start_time, end_time , duration").Where("id=?", student.ContestID).First(&contest).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error in searching the contests %s" , err.Error())})
		return
	}

	if time.Now().Before(contest.StartTime) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Contest has not started yet"})
		return
	}
	if time.Now().After(contest.EndTime) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Contest has already ended"})
		return
	}

	var questions []slimQuestion 
	if err := database.DB.Model(&models.Question{}).Select("id" , "mcq_or_code").Where("contest_id=?", student.ContestID).Order("mcq_or_code DESC").Find(&questions).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error in searching the questions %s" , err.Error())})
		return
	}

	// now make a object with two arrays one for code and second fro mcq and this araays will contain id only
	var question models.Question
	var testcase []interface{}
	var questionIDS []string
	
	if questions[0].McqOrCode == "M" {
		if err := database.DB.Model(&models.Question{}).Select("id", "name", "description", "time" , "mcq_options","mcq_or_code").Where("id=?", questions[0].ID).First(&question).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error in searching the questions %s" , err.Error())})
			return
		}
	} else {
		if err := database.DB.Model(&models.Question{}).Select("id", "name", "description", "time","mcq_or_code").Where("id=?", questions[0].ID).First(&question).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error in searching the questions %s" , err.Error())})
			return
		}
		// GetTestcaseData is function is used to get the testcases for the code question
		test , err := GetTestcaseData("g" , question.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error in getting the testcases %s" , err.Error())})
			return
		}
		testcase = test
	}

	
	for _, q := range questions {
		questionIDS = append(questionIDS, q.ID)
	}

	
	// update the student to has_logged_in_once true
	if err := database.DB.Model(&models.Student{}).Where("id=?", student.ID).Update("has_logged_in_once", true).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error in updating the student %s" , err.Error())})
		return
	}
	// Set the cookie for authorization
	ctx.SetCookie("Authorization", "Bearer "+ student.Token, int(60*int(contest.EndTime.Sub(time.Now()).Minutes())), "/", "", false, true)  // for 2 hours
	ctx.JSON(http.StatusOK, gin.H{
		"message": gin.H{
			"questionIds": questionIDS,
			"question": question,
			"testcase": testcase,
			"contestID": student.ContestID,
			"contestStartTime": contest.StartTime,
			"contestEndTime": contest.EndTime,
			"remainingMinutes": int(contest.EndTime.Sub(time.Now()).Minutes()),
		},
	})
}

// l

