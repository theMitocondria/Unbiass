package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"github.com/gin-gonic/gin"
 	"github.com/golang-jwt/jwt/v5"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
)

/*
	now till now we have checked if token is right and not expired .
	now as company can not initialize testing and other tasks like plag check , ranklist creation .
	user can only give submit compile and run .  so each of them will now have different middleware .
*/
func AuthCompany(ctx *gin.Context){
	// Get the cookie off req
	tokenString, err := ctx.Cookie("Authorization")

	// Check if the cookie is present
	if err != nil {
		ctx.JSON(http.StatusUnauthorized , gin.H{"error" : err.Error()})
		// after below line no further module / middleware will be called
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token , err := jwt.Parse(tokenString , func (token *jwt.Token)(interface{} , error){
		// jwt.SigningMethodHMAC is a type of jwt.SigningMethod which is told to the jwt.Parse function , ok will be false if the token is not a jwt.SigningMethodHMAC
		if _ , ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil , fmt.Errorf("unexpected sigining method %v" , token.Header["alg"])
		}

		return []byte(os.Getenv("COMPANY_SECRET")) , nil
	})
	
	// claims is a map of the claims in the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if float64(time.Now().Unix()) > exp {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// now we will take out sub from claims and search for it in company table
		var company models.Company
		database.DB.Model(&models.Company{}).Where("id = ?", claims["sub"]).First(&company)

		if company.ID == "" {
			ctx.JSON(http.StatusUnauthorized , gin.H{"error": "unauthorized"})
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// now we will set the company in the context
		ctx.Set("company", company)
		ctx.Next()
	}else {
		ctx.JSON(http.StatusUnauthorized , gin.H{"error": "unauthorized"})
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}		
}


func AuthAdmin(ctx *gin.Context){
	// Get the cookie off req
	tokenString, err := ctx.Cookie("Authorization")
	// Check if the cookie is present
	fmt.Println(tokenString)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized , gin.H{"error" : err.Error()})
		// after below line no further module / middleware will be called
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token , err := jwt.Parse(tokenString , func (token *jwt.Token)(interface{} , error){
		// jwt.SigningMethodHMAC is a type of jwt.SigningMethod which is told to the jwt.Parse function , ok will be false if the token is not a jwt.SigningMethodHMAC
		if _ , ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil , fmt.Errorf("unexpected sigining method %v" , token.Header["alg"])
		}

		return []byte(os.Getenv("ADMIN_SECRET")) , nil
	})
	fmt.Println(token)
	// claims is a map of the claims in the token
	if err != nil || token == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, ok := claims["exp"].(float64)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if float64(time.Now().Unix()) > exp {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// now we will take out sub from claims and search for it in company table
		var dev models.Developer
		database.DB.Model(&models.Developer{}).Where("id = ?", claims["sub"]).First(&dev)

		if dev.ID == "" {
			ctx.JSON(http.StatusUnauthorized , gin.H{"error": "unauthorized"})
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// now we will set the company in the context
		ctx.Set("admin", dev)
		ctx.Next()
	}else {
		ctx.JSON(http.StatusUnauthorized , gin.H{"error": "unauthorized"})
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}		
}
