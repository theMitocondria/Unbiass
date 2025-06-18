package controllers

import (
    "os"
    "fmt"
    "sync"
	"time"
    "net/http"
    "context"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/theMitocondria/slimuuid"
    "github.com/theMitocondria/Unbiass/models" 
    "github.com/theMitocondria/Unbiass/inits" 
	"github.com/theMitocondria/Unbiass/database"
    "github.com/theMitocondria/Unbiass/utils"
	"github.com/theMitocondria/Unbiass/awsHandler"

)


type ErrChanResponse struct {
    Status int
    Error error
    Message int
}
func CompileSubmission(ctx *gin.Context) {

    // Create request with raw body HandleCompile
    var body struct {
        Code string `json:"code" binding:"required"`
        Lang string `json:"lang" binding:"required"`
        Input string `json:"input" `
    }

    if err := ctx.ShouldBindJSON(&body); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    fmt.Println(body)

    resp , err := utils.HandleCompile(body.Code , body.Input , body.Lang)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, resp)
}

func SubmitSubmission(ctx *gin.Context){
	/* inputs : lang , code , stid , qid  
       steps : 1. aws se questionId GetTestcaseByQuestionIDAndType se testcase joki json fromat me honge array jisme 10 inputs or outputs honge 
                2. Now us inputs ke liye line by line run krdenge compile function ko call krke
                3. Compare krege output wanted or got output ko
    */
    //step1 calling aws to get a file by questionid and type
    var Body struct {
        QuestionID string `json:"question_id" binding:"required"`
        Code  string `json:"code" binding:"required"`
        Type string `json:"type" binding:"required"`
        Lang  string `json:"lang" binding:"required"`
        StudentID string `json:"student_id" binding:"required"`
    }

    if err:= ctx.ShouldBindJSON(&Body) ; err != nil {
        fmt.Println(err.Error())
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    transaction_id, err := slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK , gin.H{"message":transaction_id})


    //starting a transaction's redis record
    
    go func (){
        bgCtx, cancel := context.WithCancel(context.Background())
        defer cancel()

        inits.RedisClient.Set(bgCtx,"progress"+transaction_id,"Started",time.Hour * 24)
        content , err := GetTestcaseData(Body.Type , Body.QuestionID) 
        if err != nil {
            inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, err.Error(), time.Hour * 24) 
            cancel()
            return 
        }

        var wg sync.WaitGroup
        errChan := make (chan ErrChanResponse, 1 )
        // totaltestCase := len(content)
        // completedTestCase := 0
        var mu sync.Mutex
        var stopProcessing bool
        completedTestCase := 0
        totaltestCase := len(content)
        // step2 each time make a call to compile with command
        for _ , curr := range content {
            wg.Add(1)
            go func(curr interface{}){
                defer wg.Done()
                
                mu.Lock()
                if stopProcessing {
                    mu.Unlock()
                    return
                }
                mu.Unlock()
                var response utils.CompileResponse
                response, err := utils.HandleCompile(Body.Code, curr.(map[string]interface{})["input"].(string), Body.Lang)
                if err != nil {
                    errChan <- ErrChanResponse{
                        Status : 0 ,
                        Error : err ,
                    }
                    return
                }
            
                if utils.CleanString(response.Output.CodeOutput) != utils.CleanString(curr.(map[string]interface{})["output"].(string)) {
                    mu.Lock()
                    if !stopProcessing {
                        stopProcessing = true
                        errChan <- ErrChanResponse{
                            Status: 1,
                            Message: 0,
                        }
                    }
                    mu.Unlock()
                    return
                }

                mu.Lock()
                if !stopProcessing {
                    completedTestCase++
                    progress := float64(completedTestCase) / float64(totaltestCase) * 100
                    inits.RedisClient.Set(ctx, "progress:"+transaction_id, 
                        strconv.FormatFloat(progress, 'f', 2, 64), time.Hour * 24)
                }
                mu.Unlock()
            }(curr)

        }
        // Create a done channel to signal completion
        done := make(chan struct{})
        go func() {
            wg.Wait()
            close(done)
        }()

        select {
        case err := <-errChan:
            if err.Status == 1 {
                if errors := CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.WrongAnswer, Body.Code); errors != nil {
                    inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, errors.Error(), time.Hour * 24)
                    return
                }
                inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Wrong Answer", time.Hour * 24)
            } else if err.Status == 2 {
                if errors := CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.CompilationError, Body.Code); errors != nil {
                    inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, errors.Error(), time.Hour * 24)
                    return
                }
                inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Compilation Error", time.Hour * 24)
            }
            return
        case <-done:
            // All tests passed
           
            //check if this student already has passed this question then dont increase his score otherwise increase it
            Scored  := CheckStudentQuestionPassedOrNot(Body.StudentID, Body.QuestionID) 

            fmt.Println(Scored)
            if err := CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.PretestsPassed, Body.Code); err != nil {
                inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, err.Error(), time.Hour * 24)
                return
            }
            if !Scored {
                mu.Lock()
                var Student models.Student 

                if err := database.DB.Model(&models.Student{}).
                            Where("id=?",Body.StudentID).
                            Select("score").
                            First(&Student).Error ;err != nil {
                    inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Retry", time.Hour * 24)
                    return
                }

                fmt.Println(Student.Score)

                database.DB.Model(&models.Student{}).Where("id = ?", Body.StudentID).Update("score", Student.Score + 10)
                mu.Unlock()
            }

            inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Pretests Passed", time.Hour * 24)
        } 
    }()
    
}



func SystemTesting(ctx *gin.Context){
	contestID := ctx.Param("contestid")
	if contestID == "" {
		ctx.JSON(http.StatusBadRequest , gin.H{"error":"ContestID not found"})
		return
	}

	var Questions []models.Question
	if err := 	database.DB.Model(&models.Question{}).
				Where("contest_id=?",contestID).
				Find(&Questions).Error ; err != nil {
					ctx.JSON(http.StatusBadRequest , gin.H{"error":err.Error()})
					return
				}

	transaction_id , err := slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError , gin.H{"error":err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": transaction_id})

	go func(){
		bgCtx, cancel := context.WithCancel(context.Background())
        defer cancel()

		inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Started", time.Hour*24)

		for index,question := range Questions {
			var submissions []models.Submission
			subQuery := database.DB.
			Model(&models.Submission{}).
			Select("student_id, MAX(time) as max_submission_time").
			Where("question_id = ? AND verdict = ?", question.ID, models.PretestsPassed).
			Group("student_id")

			// Use the subquery in the main query with an inner join.
			err := database.DB.
			Model(&models.Submission{}).
			Joins("INNER JOIN (?) AS s2 ON submissions.student_id = s2.student_id AND submissions.time = s2.max_submission_time", subQuery).
			Where("submissions.question_id = ? AND submissions.verdict = ?", question.ID, models.PretestsPassed).
			Find(&submissions).Error

			if err != nil {
				inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Error Fetching submissions", time.Hour*24)
                cancel()
				return
			}  
			
			testCases , err := GetTestcaseData("t" ,question.ID) 
			if err != nil {
				inits.RedisClient.Set(bgCtx,"progress:"+transaction_id,"Error Fetching testcase",time.Hour*24)
				cancel()
                return
			}  
			
			//workers bnade 5 
			var wg sync.WaitGroup
            errChan := make(chan error, 1)
            workers := make(chan struct{}, 5)
            completedCount := 0
            totalCount := len(submissions)
            var mu sync.Mutex
            var stopProcessing bool

            for _, submission := range submissions {
                select {
                case <-bgCtx.Done():
                    inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Operation cancelled", time.Hour*24)
                    return
                case workers <- struct{}{}:
                    wg.Add(1)
                    go func(submission models.Submission) {
                        defer wg.Done()
                        defer func() { <-workers }()

                        mu.Lock()
                        if stopProcessing {
                            mu.Unlock()
                            return
                        }
                        mu.Unlock()

                        body, err := awsHandler.DownloadS3Object("unbiasss", fmt.Sprintf(submission.ID+".txt"))
                        if err != nil {
                            mu.Lock()
                            if !stopProcessing {
                                stopProcessing = true
                                errChan <- err
                                cancel()
                            }
                            mu.Unlock()
                            return
                        }

                        for _, curr := range testCases {
                            var response utils.CompileResponse
                            response, err := utils.HandleCompile(string(body), curr.(map[string]interface{})["input"].(string), submission.Language)
                            if err != nil {
                                mu.Lock()
                                if !stopProcessing {
                                    stopProcessing = true
                                    errChan <- err
                                    cancel()
                                }
                                mu.Unlock()
                                return
                            }
	
							if utils.CleanString(response.Output.CodeOutput) != utils.CleanString(curr.(map[string]interface{})["output"].(string)) {
								continue 
							}


							database.DB.Model(&models.Student{}).
								Where("id=?",submission.StudentID).
								Update("Score",20)

							database.DB.Model(&models.Submission{}).
								Where("id=?",submission.ID).
								Update("Verdict",models.SystemTestsPassed)
							// Update progress only if not stopped
                            mu.Lock()
                            if !stopProcessing {
                                completedCount++
                                progress := (float64(index)/float64(len(Questions)) + 
                                    (1.0/float64(len(Questions))) * (float64(completedCount)/float64(totalCount))) * 10
                                inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, 
                                    fmt.Sprintf("%.2f", progress), time.Hour*24)
                            }
                            mu.Unlock()
                        }
                    }(submission)
                }
            }

            done := make(chan struct{})
            go func() {
                wg.Wait()
                close(done)
            }()

            select {
            case err := <-errChan:
                progress, _ := inits.RedisClient.Get(bgCtx, "progress:"+transaction_id).Result()
                inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, 
                    fmt.Sprintf("Progress stopped at %s: %v", progress, err), time.Hour*24)
                return
            case <-done:
                continue
            case <-bgCtx.Done():
                inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Operation cancelled", time.Hour*24)
                return
            }
        }

        inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, "Completed", time.Hour*24)
    }()
}



func SubmitSubmissionWithoutRedis(ctx *gin.Context) {
    var Body struct {
        QuestionID string `json:"question_id" binding:"required"`
        Code       string `json:"code" binding:"required"`
        Type       string `json:"type" binding:"required"`
        Lang       string `json:"lang" binding:"required"`
        StudentID  string `json:"student_id" binding:"required"`
    }

    if err := ctx.ShouldBindJSON(&Body); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    testcases, err := GetTestcaseData(Body.Type, Body.QuestionID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch testcases"})
        return
    }

    var wg sync.WaitGroup
    resultChan := make(chan string, 1) // will return only one final result
    done := make(chan struct{})
    var once sync.Once

    for _, tc := range testcases {
        wg.Add(1)
        go func(tc interface{}) {
            defer wg.Done()

            input := tc.(map[string]interface{})["input"].(string)
            expected := tc.(map[string]interface{})["output"].(string)
            var response utils.CompileResponse
            response , err := utils.HandleCompile(Body.Code, input, Body.Lang)
            if err != nil {
                once.Do(func() {
                    CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.CompilationError, Body.Code)
                    resultChan <- "Compilation Error"
                })
                return
            }

            
            if response.Error != "" {
                once.Do(func() {
                    CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.CompilationError, Body.Code)
                    resultChan <- "Compilation Error"
                })
                return
            }

            if utils.CleanString(response.Output.CodeOutput) != utils.CleanString(expected) {
                once.Do(func() {
                    CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.WrongAnswer, Body.Code)
                    resultChan <- "Wrong Answer"
                })
                return
            }
        }(tc)
    }

    // Close the done channel after all goroutines complete
    go func() {
        wg.Wait()
        close(done)
    }()

    // Wait for either an error from a test or all passing
    select {
    case status := <-resultChan:
        ctx.JSON(http.StatusOK, gin.H{"status": status})
        return
    case <-done:
        // All tests passed
        _ = CreateSubmission(Body.Lang, Body.StudentID, Body.QuestionID, models.PretestsPassed, Body.Code)

        if !CheckStudentQuestionPassedOrNot(Body.StudentID, Body.QuestionID) {
            var student models.Student
            if err := database.DB.Model(&models.Student{}).Where("id=?", Body.StudentID).First(&student).Error; err == nil {
                database.DB.Model(&models.Student{}).Where("id=?", Body.StudentID).Update("score", student.Score+10)
            }
        }

        ctx.JSON(http.StatusOK, gin.H{"status": "Pretests Passed"})
        return
    }
}
