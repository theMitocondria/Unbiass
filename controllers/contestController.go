package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/models"
	"fmt"
	"io"
	"github.com/theMitocondria/Unbiass/awsHandler"
	"time"
)

func CreateContest(ctx *gin.Context){
	var body struct {
		Name             string    `form:"name" binding:"required"`
		Description      string    `form:"description" binding:"required"`
		Code             bool      `form:"code"`
		MCQ              bool      `form:"mcq"`
		StudentsRequired uint32    `form:"students_required" binding:"required,min=1"`
		StartTime        time.Time `form:"start_time" binding:"required"`
		Duration         uint32    `form:"duration" binding:"required,min=1"`
		CompanyID        string    `form:"company_id" binding:"required"`
	}

	if err := ctx.ShouldBind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error" : err.Error()})
		return
	}

	file , err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error" : "file not found"})
		return
	}

    src, err := file.Open()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError , gin.H{"error": "Failed to open file"})
        return
    }
    defer src.Close() // Ensure the file is closed after use

    // Read the file data into a byte slice
    data, err := io.ReadAll(src)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError , gin.H{"error": "Failed to read file"})
        return
    }

    // Convert byte slice to string
    content := string(data)

	contest := models.Contest{
		Name: body.Name,
		Description: body.Description,
		MCQ: body.MCQ,
		Code: body.Code,
		StudentsRequired: body.StudentsRequired,
		StartTime: body.StartTime,
		EndTime: body.StartTime.Add(time.Duration(body.Duration) * time.Minute),
		Duration: body.Duration,
		CompanyID: body.CompanyID,
	}

	if err := database.DB.Model(&models.Contest{}).Create(&contest).Error ; err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	if err := awsHandler.UploadContentToS3(fmt.Sprintf(contest.ID + ".csv") , content , "unbiass") ; err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK , gin.H{"message" : "contest created"})
}
