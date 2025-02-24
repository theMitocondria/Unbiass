package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/lib/pq"
	"net/http"		
	"fmt"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/go-playground/validator/v10"
)

func CreateQuestion(ctx *gin.Context){
	var body struct {
		Name				string			`json:"name" binding:"required"`
	    Description 		string			`json:"description" binding:"required"`
	    // "M" for MCQ , "C" for Code
	    MCQ_OR_Code			string			`json:"mcq_or_code" binding:"required,oneof=M C"`
	    Tags 				pq.StringArray	`json:"tags"`
	    Difficulty_rating 	uint32			`json:"difficulty_rating"`
	    Language 			string			`json:"language"`
		// if mcq then only here 4-5 options whatever and last option for answer of that (4 options , 5 one answer)
		MCQ_Options			pq.StringArray	`json:"mcq_options"`
		ContestID 			string			`json:"contest_id"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		if _, ok := err.(validator.ValidationErrors); ok {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation failed",
				"details": err.Error(),
			})
			return
		}
		
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
		})
		return
	}

	question := models.Question{
		Name: body.Name,
		Description: body.Description,
		MCQ_OR_Code: body.MCQ_OR_Code,
		Tags: body.Tags,
		Difficulty_rating: body.Difficulty_rating,
		Language: body.Language,
		MCQ_Options: body.MCQ_Options,
	}

	result := database.DB.Model(&models.Question{}).Create(&question)

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": len(question.ID)})
}


/*
	Get all questions paginated
	Get question by difficulty rating paginated
	Get question by tags paginated
	Get question by mcq or code paginated
	Get question by language paginated
	Get question by contest id (more than 1 question)
	Get question by questionid , jab student mange (1 question)
	Delete question by contest id (more than 1 question) 
	Edit question by questionid (1 question) 

	So we will have 3 controllers for this
	1. find with questionid
	2. find with contestid
	3. all others paginated resources
*/


func GetQuestionByQuestionID(ctx *gin.Context){
	questionid := ctx.Param("questionid")

	var question models.Question
	database.DB.Model(&models.Question{}).Where("id = ?", questionid).First(&question)
	ctx.JSON(http.StatusOK, gin.H{"message": question})
}

func GetQuestionByContestID(ctx *gin.Context){
	contestid := ctx.Param("contestid")

	var questions []models.Question
	database.DB.Model(&models.Question{}).Where("contest_id = ?", contestid).Find(&questions)
	ctx.JSON(http.StatusOK, gin.H{"message": questions})
}

func GetQuestionsByPaginatedFilters(ctx *gin.Context){
	// we will have these filters as query params
	type QuestionFilters struct {
		Difficulty_rating uint32 `form:"difficulty_rating"`
		Tags 				[]string `form:"tags"`
		MCQ_OR_Code			string `form:"mcq_or_code"`
		Language 			string `form:"language"`
		ContestID 			string `form:"contest_id"`
		Page 				int `form:"page"`
		Limit 				int `form:"limit"`
	}

	var filters QuestionFilters
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filters"})
		return
	}
	
	query := database.DB.Model(&models.Question{})

	if filters.Difficulty_rating != 0 {
		query = query.Where("difficulty_rating = ?", filters.Difficulty_rating)
	}

	if len(filters.Tags) > 0 {
		tags := pq.StringArray(filters.Tags)
		fmt.Println(tags)
		query = query.Where("tags @> ?", tags)
	}

	if filters.MCQ_OR_Code != "" {
		query = query.Where("mcq_or_code = ?", filters.MCQ_OR_Code)
	}

	if filters.Language != "" {
		query = query.Where("language = ?", filters.Language)
	}

	if filters.Page != 0 {
		query = query.Offset((filters.Page - 1) * filters.Limit)
	}

	if filters.Limit != 0 {
		query = query.Limit(filters.Limit)
	}

	var questions []models.Question
	query.Find(&questions)
	ctx.JSON(http.StatusOK, gin.H{"message": questions})
}

func DeleteQuestionByContestID(ctx *gin.Context){
	contestid := ctx.Param("contestid")

	var questions []models.Question
	database.DB.Model(&models.Question{}).Where("contest_id = ?", contestid).Find(&questions)
	for _, question := range questions {
		if err := DeleteTestcaseByQuestionID(question.ID) ; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	database.DB.Model(&models.Question{}).Where("contest_id = ?", contestid).Delete(&questions)
	ctx.JSON(http.StatusOK, gin.H{"message": "questions deleted"})
}

func DeleteQuestionByQuestionID(ctx *gin.Context){
	questionid := ctx.Param("questionid")

	var question models.Question
	database.DB.Model(&models.Question{}).Where("id = ?", questionid).Delete(&question)

	if err := DeleteTestcaseByQuestionID(questionid) ; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "question deleted"})
}

