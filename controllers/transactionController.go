package controllers

import(
    "fmt"
	"context"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/inits" 
)
func GetProgress(ctx *gin.Context) {
	
    jobID := ctx.Param("transaction_id")
   
    progress, err := inits.RedisClient.Get(context.Background(), "progress:"+jobID).Result()
    if err != nil {
        fmt.Println(err)
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, progress)
}
