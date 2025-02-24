package controllers 

import(
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
	"net/http"
	"github.com/theMitocondria/Unbiass/awsHandler"
	"fmt"
)

func CreateTestcase(ctx *gin.Context){
	var body struct {
		Type string `json:"type" binding:"required,oneof=s p g"`
		QuestionID string `json:"question_id" binding:"required"`
		Body string `json:"body" binding:"required"`
	}

	if err:= ctx.ShouldBindJSON(&body) ; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	testcase := models.Testcase{
		Type : body.Type ,
		QuestionID : body.QuestionID,
	}

	if err := database.DB.Model(&models.Testcase{}).Create(&testcase).Error ; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err := awsHandler.UploadContentToS3(fmt.Sprintf(testcase.ID + ".txt") , body.Body , "unbiass") 
	if err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK , gin.H{"message": testcase.ID})
}

func GetTestcaseByQuestionIDAndType(ctx *gin.Context){
	QuestionID := ctx.Param("questionid")
	Type := ctx.Param("type")
	var testcases[] models.Testcase

	if err := database.DB.Model(&models.Testcase{}).Where("type = ? AND question_id = ?" ,Type , QuestionID).Select("id").Find(&testcases).Error ; err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	var testcaseBodies []string
	for _,testcase := range testcases {
		body , err := awsHandler.DownloadS3Object("unbiass",fmt.Sprintf(testcase.ID + ".txt")) 
		if err != nil {
			ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
			return
		}

		testcaseBodies = append(testcaseBodies , string(body)) 
	}

	ctx.JSON(http.StatusOK , gin.H{"message" : testcaseBodies})
}

func DeleteTestcaseByQuestionID(questionID string) error {
	var testcases []models.Testcase
	database.DB.Model(&models.Testcase{}).Where("question_id = ?", questionID).Find(&testcases)
	for _, testcase := range testcases {
		if err := awsHandler.DeleteContentFromS3("unbiass", fmt.Sprintf(testcase.ID + ".txt")) ; err != nil {
			return err
		}
	}
	return nil
}