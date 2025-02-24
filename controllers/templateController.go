package controllers 

import (
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/awsHandler"
	"net/http"
)

func CreateEmailTemplate(ctx *gin.Context){
	var body struct {
		HTML string `binding:"required" json:"html"`
		Name string `binding:"required" json:"name"`
		Subject string `binding:"required" json:"subject"`
		Text string `binding:"required" json:"text"`
	}

	if ctx.ShouldBindJSON(&body) != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error" :"Bad request"})
		return
	}

	//create a template
	if err := awsHandler.CreateSESEmailTemplate(body.Name, body.Subject , body.Text , body.HTML ) ; err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK , gin.H{"message":"template created"})
}