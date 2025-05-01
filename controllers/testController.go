package controllers

import(
	"os"
	"os/exec"
	"fmt"
	"time"
	"strings"
	"strconv"
	"context"
	"net/http"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/theMitocondria/Unbiass/models"
	"github.com/theMitocondria/slimuuid"
	"github.com/theMitocondria/Unbiass/inits"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/awsHandler"
)

type CompileCode struct {
	Code string
	Lang string 
	Input string
	QuestionID string `json:"question_id"`
}

type SubmitCode struct {
	QuestionID string `json:"question_id" `
	Code  string `json:"code"`
	Type string `json:"type" `
	Lang  string `json:"lang" `
	StudentID string `json:"student_id"`
}

type fileData struct{
	ID string 
	Code string 
}
func frontendMonitoring(data int)(bool,error){
	return true , nil
}

func cleanOutput(str string) []string {
	
	// Remove curly braces and single quotes
	cleaned := strings.ReplaceAll(str, "{", "")
	cleaned = strings.ReplaceAll(cleaned, "}", "")
	cleaned = strings.ReplaceAll(cleaned, "'", "")

	// Split by comma and trim any whitespace
	parts := strings.Split(cleaned, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
func EndTest(ctx *gin.Context){

	var body struct{
		StudentID string `json:"student_id" binding:"required"`
		Answers []answer `json:"answers" binding:"required"`
		FrontendScore int `json:"fs" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&body) ; err != nil {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":err})
		return
	}

	transaction_id , err := slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err})
		return
	}

	inits.RedisClient.Set(context.Background(),"progress:"+transaction_id,0,0)
	ctx.JSON(http.StatusOK,gin.H{"message":transaction_id})

	go func (){
			//tab switch face detection screen left time resize event check for monitoring 
		fair , err := frontendMonitoring(body.FrontendScore);
		if err != nil {
			ctx.JSON(http.StatusInternalServerError , gin.H{"error":err})
			return
		}

		//cheating ki h to mcq to bnao hi mt or codes jo bhi submit kre h unhe delete krdo
		if !fair {
			var Submissions []models.Submission
			if err  := 	database.DB.Model(&models.Submission{}).
						Where("student_id=?").	
						Find(&Submissions).Error ; err != nil {
							inits.RedisClient.Set(context.Background(),"progress:"+transaction_id,"100",0)
						}
			
			for _,sub := range Submissions {
				if err := awsHandler.DeleteContentFromS3("unbiasss",sub.ID) ; err!=nil{
					inits.RedisClient.Set(context.Background(),"progress:"+transaction_id,"100",0)
				}
				database.DB.Model(&models.Submission{}).Delete(sub.ID)
			}
			//student ko delete krdo
			return	
		}

		//ab hume uske mcqs ko creation ke liye bhjna h
		if err := MCQSubmissionCreation(body.StudentID , body.Answers , transaction_id) ; err != nil {
			inits.RedisClient.Set(context.Background(),"progress:"+transaction_id,fmt.Sprintf("Error %v",err.Error()),time.Hour*24)
		}
	}()	
}

func PlagTesting(ctx *gin.Context){
	ContestID := ctx.Param("contestid")
	if ContestID == "" {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":"No Contestid provided"})
		return 
	}

	var contest models.Contest
	if err := database.DB.Model(&models.Contest{}).
			Where("ID=?",ContestID).
			Select("id , students_required , start_time").
			First(&contest).Error ; err != nil {
				ctx.JSON(http.StatusInternalServerError , gin.H{"error":err})
				return
			}

	transaction_id , err := slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err})
		return
	}
	ctx.JSON(http.StatusOK , gin.H{"message":transaction_id})
	
	go func(){
		bgCtx , cancel := context.WithCancel(context.Background())
		defer cancel()
		
		inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , 0 , time.Hour*24)
		var students []models.Student
		if err := database.DB.Model(&models.Student{}).
					Select("id").
					Where("test_date=? AND contest_id=?",contest.StartTime.Format("2006-01-02"),contest.ID).
					Order("score DESC").
					Find(&students).Error ; err != nil {
						inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , err.Error() , time.Hour*24)
						return
					}
		
		var questions []models.Question
		if err := database.DB.Model(&models.Question{}).
					Select("id").
					Where("contest_id=?",contest.ID).
					Find(&questions).Error ; err != nil {
						inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , err.Error() , time.Hour*24)
						return
					}
		
		indexing := map[string]int{}
		for index , question := range questions {
			indexing[question.ID]=index
		}
		
		var submissions [][]fileData
		for _ , student := range students {
			var curr []models.Submission
			subquery := database.DB.Model(&models.Submission{}).
						Select("MAX(time)").
						Where("student_id = ? AND verdict = ?", student.ID , models.SystemTestsPassed).
						Group("question_id")

			if err := database.DB.Model(&models.Submission{}).
						Select("id , question_id").
						Where("student_id=? AND time IN (?)",student.ID , subquery).
						Find(&curr).Error ; err != nil {
							inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , fmt.Sprintf("error while submission DB fetching ", err.Error()) , time.Hour*24)
							return
					}
			
		
			// time consuming 200 * 2 seconds minimum
			filedata := make([]fileData, len(questions))
			for _ , submission := range curr {
				code , err := awsHandler.DownloadS3Object("unbiasss",fmt.Sprintf("%s.txt",submission.ID))
				if err != nil {
					inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , fmt.Sprintf("error while aws fetching ", err.Error()), time.Hour*24)
					return
				}
				content := fileData{
					ID : student.ID,
					Code : string(code) ,
				}
				filedata[indexing[submission.QuestionID]] =  content
			}
			
			if len(curr) == 0 {
				continue
			} 
			for index := range len(questions) {
				if (filedata[index] == fileData{}) { // if empty
					content := fileData{
						ID:   student.ID,
						Code: "",
					}
					filedata[index] = content
				}
			}
			submissions = append(submissions,filedata)
		}
		submissionsData := struct{
			No_of_questions int `json:"no_of_questions"`
			Submissions [][]fileData `json:"submissions"`
		}{
			No_of_questions : len(questions) ,
			Submissions : submissions ,
		}
		jsonData, err := json.Marshal(submissionsData)
		if err != nil {
			inits.RedisClient.Set(bgCtx , "progress:"+transaction_id ,fmt.Sprintf("error while marshaling data ", err.Error()), time.Hour*24)
			return
			return
		}
		file , err := os.Create(fmt.Sprintf("Plag/%s.json",contest.ID))
		if err != nil {
			inits.RedisClient.Set(bgCtx , "progress:"+transaction_id ,fmt.Sprintf("error while creating ", err.Error()), time.Hour*24)
			return
		}
		defer file.Close()
		defer os.Remove(fmt.Sprintf("Plag/%s.json",contest.ID))
		_ , err = file.Write(jsonData)
		if err != nil {
			inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , fmt.Sprintf("error while writing ", err.Error()), time.Hour*24)
			return
		}
		arguements := [5]string{"python3" , "Plag/script.py" , fmt.Sprintf("Plag/%s.json",contest.ID), strconv.FormatUint(uint64(contest.StudentsRequired), 10), transaction_id}
		cmd := exec.Command(arguements[0],arguements[1],arguements[2],arguements[3],arguements[4])
		output , err := cmd.CombinedOutput()
		if err != nil {
			inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , fmt.Sprintf("error while executing python ", err.Error()), time.Hour*24)
			return
		}
		fmt.Println(string(output))
		
		studentIds := cleanOutput(string(output))
		if err := CreateRanklist(studentIds , contest.ID) ; err != nil {
			inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , fmt.Sprintf("error while creating ranklist%s",err.Error()) , time.Hour*24)
			return
		}
	}()
}





/* student ne submit button dabaya / contest end hua  =>
end points =>   end test by student ek bache ka data aaega ek request lgi (frontend monitoering ka check hoga + mcqsanswer creation  ) not time consuming
                finish automatic call at end time + 1 minute (frontend monitoering ka check hoga + mcqsanswer creation  + codes check honge) time consuming
                MCQ Testing (mcq answer comparison) time consuming
                system testing (all students without plag system testing hogi) time consuming
                plag testing (thode thode krke student mngao or plag run kro) time consuming
                ranklist generation (make rnaklist db) not time consuming
                ranklist getting (get) not
                contest ke students ko thanku for submitting mail (students mail sending) time consuming
                contest data deletion (contest , question(used) , submissions except in ranklist , students , mcq answers saare , testcases ) immediate not time consuming

bhut saare students ke {MCQ[],Codes[]} => humare paas aaye 
code hume rkhna ek question last right answer , mcq ka ek hi hoga 
mcq ko unke id , given asnswers ke saath mcq submission me bhej Do
codes ko check krlo agr right submit h to thik h werna agr koi submission nhi h to compile kro fir submit kro dono pass ho jaye to score dedo is point tak ab humne mcqs or simple
score dedia h bas ab request khtm
fir student ko frontent monitoring score ke basis pr htana H
fir system testing ko call kro
fir plag testing ke liye call kro some students only remember
fir ranklist bnao 
*/