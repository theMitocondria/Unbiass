package utils

import (
    "github.com/gin-gonic/gin"
	"log"
)

// Success sends a JSON response with a success message and data
func Success(ctx *gin.Context, statusCode int, data interface{}) {
    ctx.JSON(statusCode, gin.H{
        "data": data,
    })
}

// Error sends a JSON response with an error message
func Error(ctx *gin.Context, statusCode int, err error) {
    ctx.JSON(statusCode, gin.H{
        "error": err.Error(),
    })
}

// CheckError checks if an error is not nil and sends an error response if true
func CheckError(ctx *gin.Context, err error) bool {
    if err != nil {
        Error(ctx, 500, err)
        return true
    }
    return false
}

// LogError logs an error with an optional message
func LogError(err error, message ...string) {
    if err != nil {
        if len(message) > 0 {
            log.Fatal(message[0])
        } else {
            log.Fatal(err.Error())
        }
    }
}