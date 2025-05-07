package controllers

import(
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
)

func CreateBug(ctx *gin.Context){
	var body struct {
		Description string `json:"desc" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&body) ; err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":err.Error()})
		return
	}

	bug := models.Bug{
		Description : body.Description,
	}
	if err := database.DB.Model(&models.Bug{}).Create(&bug).Error;err!=nil{
		ctx.JSON(http.StatusInternalServerError , gin.H{"error":err.Error()})
		return
	}

	ctx.JSON(http.StatusOK,gin.H{"message":bug.ID})
}