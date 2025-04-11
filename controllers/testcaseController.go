package controllers 

import(
	"fmt"
	"errors"
	"net/http"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/awsHandler"
)

func CreateTestcase(ctx *gin.Context){
	var body struct {
		Type string `json:"type" binding:"required,oneof=s p t"`
		QuestionID string `json:"question_id" binding:"required"`
		Body []interface{} `json:"body" binding:"required"`
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

	bytes, err := json.Marshal(body.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	if err := awsHandler.UploadContentToS3(fmt.Sprintf(testcase.ID + ".json") , string(bytes) , "unbiass") ; err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK , gin.H{"message": testcase.ID})
}

func GetTestcaseByQuestionIDAndType(ctx *gin.Context){
	QuestionID := ctx.Param("questionid")
	Type := ctx.Param("type")
	var testcase models.Testcase

	if err := database.DB.Model(&models.Testcase{}).Where("type = ? AND question_id = ?" ,Type , QuestionID).Select("id").First(&testcase).Error ; err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}


    if testcase.ID == "" {
        ctx.JSON(http.StatusInternalServerError , gin.H{"error" : "No testcase for the given question og particular type "})
		return
    }

	
	body , err := awsHandler.DownloadS3Object("unbiass",fmt.Sprintf(testcase.ID + ".json")) 
	if err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}	
	
	var testcases []interface{}
    if err := json.Unmarshal(body, &testcases); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON content"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"data": testcases})

}

func DeleteTestcaseByQuestionID(questionID string) error {
	var testcases []models.Testcase
	database.DB.Model(&models.Testcase{}).Where("question_id = ?", questionID).Find(&testcases)
	for _, testcase := range testcases {
		if err := awsHandler.DeleteContentFromS3("unbiass", fmt.Sprintf(testcase.ID + ".json")) ; err != nil {
			return err
		}
	}
	return nil
}

func GetTestcaseData(Type string , QuestionID string)([]interface{} , error){
	var testcase models.Testcase
	if err := database.DB.Model(&models.Testcase{}).Where("type = ? AND question_id = ?" , Type , QuestionID).Select("id").Find(&testcase).Error ; err != nil {
		return nil ,  err
	}

    if testcase.ID == "" {
        return nil ,  errors.New("No testcase for the given question og particular type ")
    }

	// fmt.Println(testcase.ID)
	body , err := awsHandler.DownloadS3Object("unbiass",fmt.Sprintf(testcase.ID + ".json")) 
	if err != nil {
		return nil ,  err
	}	
    
    var content []interface{}
    if err := json.Unmarshal(body , &content) ; err != nil {
		return nil ,  err
    }

	return  content , nil

}