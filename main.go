package main

import (
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/inits"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/controllers"
	"github.com/theMitocondria/Unbiass/awsHandler"
)

func init(){
	//used for loading envs in the code
	inits.LoadEnv()
	database.DBInit()
	awsHandler.InitializeAWS()
	awsHandler.CreateBucket("unbiass")
	awsHandler.InitializeSESService()
}
func main(){

	r:= gin.Default()

	// company routes
	r.POST("/api/v1/companies" , controllers.CreateCompany)	
	r.POST("/api/v1/companies/auth" , controllers.LoginCompany)
	r.POST("/api/v1/companies/logout" , controllers.LogoutCompany)
	
	// question routes
	r.POST("/api/v1/questions" , controllers.CreateQuestion)
	r.GET("/api/v1/questions/:questionid" , controllers.GetQuestionByQuestionID)
	r.GET("/api/v1/contests/:contestid/questions" , controllers.GetQuestionByContestID)
	r.GET("/api/v1/questions" , controllers.GetQuestionsByPaginatedFilters) // in these i need to add query parameters for filters
	r.DELETE("/api/v1/questions/:questionid" , controllers.DeleteQuestionByQuestionID)
	r.DELETE("/api/v1/contests/:contestid/questions" , controllers.DeleteQuestionByContestID)

	// testcase routes 
	r.POST("/api/v1/testcases" , controllers.CreateTestcase)
	r.GET("/api/v1/testcases/questions/:questionid/type/:type" , controllers.GetTestcaseByQuestionIDAndType)

	// contest routes
	r.POST("/api/v1/contests" , controllers.CreateContest)

	// student routes
	r.POST("/api/v1/contests/:contestid/students" , controllers.CreateStudentsFromCSVFromContestIDAndSendMail)
	r.POST("/api/v1/students/:studentid/auth",controllers.AuthStudentInfo)

	//templates
	r.POST("/api/v1/templates" , controllers.CreateEmailTemplate)
	r.Run()
}