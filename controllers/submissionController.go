package controllers

import(
	"os"
	"fmt"
	"time"
	"net/http"
	"github.com/theMitocondria/slimuuid"
	"github.com/theMitocondria/Unbiass/awsHandler"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/gin-gonic/gin"
)
func CreateSubmission(Lang string , StudentID string , QuestionID string , Verdict models.Verdict , Code string)error{
    MAC_ADDRESS := os.Getenv("MAC_ADDRESS")
	id , err := slimuuid.GenerateBest(MAC_ADDRESS)
	if err != nil {
		return err
	}

    submission := models.Submission{
		ID : id,
        Time : time.Now(),
        Language:Lang,
        StudentID:StudentID,
        QuestionID:QuestionID,
        Verdict:Verdict,
    }

    results := database.DB.Model(&models.Submission{}).Create(&submission)
	
	if results.Error != nil {
		return results.Error
	}

	if err := awsHandler.UploadContentToS3(fmt.Sprintf(submission.ID + ".txt") , Code, "unbiasss") ; err != nil {
		return err
	}

    return nil
}

func CheckStudentQuestionPassedOrNot(StudentID string , QuestionID string )bool {
	var submission []models.Submission 
	if err := database.DB.Model(&models.Submission{}).
				Where("student_id=? AND question_id=? AND verdict=? ",StudentID,QuestionID,models.PretestsPassed).
				First(&submission).Error ; err != nil {
					return false 
				}

	return true 
}

func GetSubmissionByQuestionIDAndLIMIT(questionID string ,StudentsRequired uint32 , index uint32 )([]models.Submission , error){
	var submissions[]models.Submission
	if err := database.DB.Model(&models.Submission{}).
				Where("question_id=? AND Verdict=?",questionID , 1).
				Select("ID , student_id").
				Order("time ASC").
				Offset(int(index)*int(StudentsRequired)).
				Limit(int(StudentsRequired)).
				Find(&submissions).Error ; err != nil {
					return submissions , err 
				}

	return submissions , nil
}


func GetSubmissionsBYQuestionIDAndStudentID(ctx *gin.Context) {
	var questionid = ctx.Param("questionid")
	var studentid = ctx.Param("studentid")

	if questionid == "" || studentid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "fields empty"})
		return
	}

	var Submissions []models.Submission
	if err := database.DB.Model(&models.Submission{}).Where("student_id=? AND question_id=?", studentid, questionid).Find(&Submissions).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": Submissions})
}

func GetSubmissionCodeByID(ctx *gin.Context){
	id := ctx.Param("submissionid")
	if(id == ""){
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "fields empty"})
		return
	}	

	var Submission models.Submission
	if err := database.DB.Model(&models.Submission{}).Where("id=?", id).Find(&Submission).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	body, err := awsHandler.DownloadS3Object("unbiasss", fmt.Sprintf(id+".txt"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": string(body) , "submission":Submission})
		
}

