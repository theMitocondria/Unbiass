package controllers

import(
	"context"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/inits" 
)
func GetProgress(ctx *gin.Context) {
	
    jobID := ctx.Param("transaction_id")
    progress, err := inits.RedisClient.Get(context.Background(), "progress:"+jobID).Result()
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Job ID not found"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"job_id": jobID, "progress": progress})
}
