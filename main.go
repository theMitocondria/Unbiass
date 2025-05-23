package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/theMitocondria/Unbiass/inits"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/controllers"
	"github.com/theMitocondria/Unbiass/awsHandler"
	"github.com/theMitocondria/Unbiass/middlewares"
)

func init(){
	//used for loading envs in the code
	// inits.LoadEnv()
	inits.InitHTTPClient()
	inits.Redis()
	database.DBInit()
	awsHandler.InitializeAWS()
	awsHandler.CreateBucket("unbiass")
	awsHandler.InitializeSESService()
}
func main(){

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://unbiass.com"}, // adjust for your frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))


	// dev routes
	r.POST("/api/v1/admin",controllers.CreateAdmin)
	r.POST("/api/v1/admin/auth",controllers.LoginAdmin)
	r.HEAD("/api/v1/admin/logout",middlewares.AuthAdmin , controllers.LogoutAdmin)
	// company routes
	r.POST("/api/v1/companies" , controllers.CreateCompany)	
	r.POST("/api/v1/companies/auth" , controllers.LoginCompany)
	r.HEAD("/api/v1/companies/logout" , middlewares.AuthCompany , controllers.LogoutCompany)
	
	// question routes
	r.POST("/api/v1/questions" ,middlewares.AuthAdmin , controllers.CreateQuestion)
	r.GET("/api/v1/questions/:questionid" , controllers.GetQuestionByQuestionID)
	r.GET("/api/v1/contests/:contestid/questions" , controllers.GetQuestionByContestID)
	r.GET("/api/v1/questions" , controllers.GetQuestionsByPaginatedFilters) // in these i need to add query parameters for filters
	r.DELETE("/api/v1/questions/:questionid" ,middlewares.AuthAdmin , controllers.DeleteQuestionByQuestionID)
	r.DELETE("/api/v1/contests/:contestid/questions" , middlewares.AuthAdmin , controllers.DeleteQuestionByContestID)

	// testcase routes 
	r.POST("/api/v1/testcases" ,middlewares.AuthAdmin , controllers.CreateTestcase)
	r.GET("/api/v1/testcases/questions/:questionid/type/:type" , controllers.GetTestcaseByQuestionIDAndType)

	// contest routes
	r.POST("/api/v1/contests" , controllers.CreateContest)

	// student routes
	r.POST("/api/v1/contests/:contestid/students" ,middlewares.AuthAdmin, controllers.CreateStudentsFromCSVFromContestIDAndSendMail)
	r.POST("/api/v1/students/:studentid/auth",controllers.AuthStudentInfo)

	//templates
	r.POST("/api/v1/templates" , middlewares.AuthAdmin , controllers.CreateEmailTemplate)

	//submission routes
	r.POST("/api/v1/submissions/compile" , controllers.CompileSubmission)
	r.POST("/api/v1/submissions/submit" , controllers.SubmitSubmissionWithoutRedis)
	r.POST("/api/v1/submissions/system/:contestid" ,middlewares.AuthAdmin, controllers.SystemTesting)
	r.POST("/api/v1/submissions/mcq/:contestid" , middlewares.AuthAdmin ,controllers.MCQSubmissionTesting)
	r.POST("/api/v1/submissions/plag/:contestid" ,middlewares.AuthAdmin, controllers.PlagTesting)
	r.GET("/api/v1/submissions/students/:studentid/questions/:questionid",controllers.GetSubmissionsBYQuestionIDAndStudentID)
	r.GET("/api/v1/submissions/:submissionid",controllers.GetSubmissionCodeByID)
	//miscellaneous
	r.POST("/api/v1/end",controllers.EndTest)
	r.POST("/api/v1/create-template",controllers.CreateTemplate)
	
	//trnasaction
	r.GET("/api/v1/transactions/:transaction_id",controllers.GetProgress)

	//bugs and feedback
	r.POST("/api/v1/feedback",controllers.CreateFeedback)
	r.POST("/api/v1/bug",controllers.CreateBug)
	r.Run()
}