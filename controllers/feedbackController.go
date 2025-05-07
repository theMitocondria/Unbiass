package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
)

func CreateFeedback(ctx *gin.Context) {
	var body struct {
		Responses []models.Response `json:"responses" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert []Response to ResponseSlice
	responses := models.ResponseSlice(body.Responses)

	feedback := models.Feedback{
		Responses: responses,
	}

	if err := database.DB.Create(&feedback).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Feedback created successfully",
		"id":      feedback.ID,
	})
}
