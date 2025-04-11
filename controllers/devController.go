package controllers

import(
	"os"
	"time"
	"net/http"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
)

func CreateAdmin(ctx *gin.Context){
	var body struct {
		Email string `json:"email"`
		Secret string `json:"secret"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&body) ; err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":err.Error()})
		return
	}

	if body.Secret != os.Getenv("ADMIN_SECRET") {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":"Wrong Secret to become admin"})
		return
	}

	hash , err := bcrypt.GenerateFromPassword([]byte(body.Password), 7)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	Dev := models.Developer{
		Password : string(hash),
		Email : body.Email,
	}

	if err := database.DB.Model(&models.Developer{}).Create(&Dev).Error ; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK , gin.H{"message":1})	
}

func LoginAdmin(ctx *gin.Context){
	var body struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&body) ; err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":err.Error()})
		return
	}

	var Dev models.Developer
	database.DB.Model(&models.Developer{}).Where("email=?",body.Email).Find(&Dev)
	if Dev.ID == "" {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":"Wrong Email or Password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(Dev.Password) , []byte(body.Password))
	if err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":"Wrong Email or Password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": Dev.ID,
        "exp": time.Now().Add(time.Hour * 24).Unix(), // Use Unix timestamp
        "rol": "admin",
    })
	
	tokenString , err := token.SignedString([]byte(os.Getenv("ADMIN_SECRET")))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	// here we are just writing it to headers only
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // or LaxMode, NoneMode
	})	
	ctx.Writer.WriteHeader(http.StatusOK)
}

func LogoutAdmin(ctx *gin.Context){
	ctx.SetCookie("Authorization", "", -1, "/", "", false, true)
	ctx.Writer.WriteHeader(http.StatusOK)
}