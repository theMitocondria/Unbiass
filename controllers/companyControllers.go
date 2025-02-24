package controllers

import (
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/Unbiass/database"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func CreateCompany(ctx *gin.Context){
	var body struct {
		Name string `json:"name" binding:"required"`
		// this required check if field is present and when we hae added email with that to it is check if the givn is just ext or proper email
		Email string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// whatever json data came from request it tries to bind it to body struct
	if ctx.BindJSON(&body) != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	// average response time 170ms with 10 cost , 7ms with 4 cost , 12ms with 6 cost , 14ms with 7 cost so 7 can be choosen as it will be faster than 10 and also secure
	hash , err := bcrypt.GenerateFromPassword([]byte(body.Password), 7)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Generate a hash password
	company := models.Company{
		Name: body.Name,
		Email: body.Email,
		Password: string(hash),
	}

	// creare a db entry
	results := database.DB.Model(&models.Company{}).Create(&company)

	if results.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to create company"})
		return
	}
	// here we are not returning any body but a headers only with status code so we can use 201 status code 
	ctx.Writer.WriteHeader(http.StatusCreated)
}

func LoginCompany(ctx *gin.Context){
	var body struct {
		Email string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if ctx.BindJSON(&body) != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	var company models.Company 
	database.DB.Model(&models.Company{}).Where("email = ?", body.Email).First(&company)

	if company.ID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email or password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(company.Password), []byte(body.Password))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email or password"})
		return
	}	

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": company.ID,
		"exp": time.Now().Add(time.Hour * 30 *24).Unix(),//expires after 30 days
		"rol":"company",
	})
	
	tokenString, err := token.SignedString([]byte(os.Getenv("COMPANY_SECRET")))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	
	// here we are just writing it to headers only
	ctx.SetCookie("Authorization", "Bearer "+tokenString, 60*60*24*30, "/", "", false, true) 
	ctx.Writer.WriteHeader(http.StatusOK)
}

func LogoutCompany(ctx *gin.Context){
	ctx.SetCookie("Authorization", "", -1, "/", "", false, true)
	ctx.Writer.WriteHeader(http.StatusOK)
}

type CompanyFilters struct {
	Name      string `form:"name"`
	Email     string `form:"email"`
	SortBy    string `form:"sort"`
	Limit     int    `form:"limit,default=10"`
	Page      int    `form:"page,default=1"`
	CreatedAt string `form:"created_at"`
}

func GetCompaniesWithStruct(ctx *gin.Context) {
	var filters CompanyFilters
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := database.DB.Model(&models.Company{})

	// Apply filters
	if filters.Name != "" {
		query = query.Where("name LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Email != "" {
		query = query.Where("email LIKE ?", "%"+filters.Email+"%")
	}
	if filters.CreatedAt != "" {
		query = query.Where("created_at >= ?", filters.CreatedAt)
	}

	// Apply sorting
	if filters.SortBy != "" {
		query = query.Order(filters.SortBy)
	}

	// Calculate offset
	offset := (filters.Page - 1) * filters.Limit

	var companies []models.Company
	var total int64
	query.Count(&total)
	result := query.Limit(filters.Limit).Offset(offset).Find(&companies)

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch companies"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": companies,
		"total": total,
		"page": filters.Page,
		"limit": filters.Limit,
	})
}

