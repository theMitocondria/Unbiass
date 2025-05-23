package controllers

import (
    "os"
    "fmt"
    "time"
    "context"
	"sync"
	"net/http"
	"github.com/gin-gonic/gin"
    "github.com/theMitocondria/Unbiass/inits"
	"github.com/theMitocondria/Unbiass/database"
	"github.com/theMitocondria/Unbiass/models"
    "github.com/theMitocondria/slimuuid"
)

type answer struct {
    Answer string 
    QuestionID string
}

func MCQSubmissionCreation(studentID string , answers []answer , transaction_id string) error {
	/*
        student ID ek single hogi , answers array h jisme question ID or given option hoga : steps :
            1. student id  => saare answers or options wale objects bna or sabko ek saath create krdio
            2. then create krna
    */

    var McqAnswers []models.McqAnswer
    for index,answer := range answers {
        McqAnswer := models.McqAnswer{
            Answer : answer.Answer ,
            StudentID : studentID ,
            QuestionID : answer.QuestionID ,
        }
        McqAnswers = append(McqAnswers,McqAnswer)
        // progress := float64(index)/float64(len(McqAnswers)) * 100

        // if int(progress) % 10 == 0 {
        //     inits.RedisClient.Set(context.Background(),"progress:"+transaction_id,progress,0)
        // }
    } 

    if err := database.DB.Model(&models.McqAnswer{}).Create(&McqAnswers).Error ; err != nil {
        return err
    }

    return nil 
}

func MCQSubmissionTesting(ctx *gin.Context){
    
    ContestID := ctx.Param("contestid")
    if ContestID == "" {
        ctx.JSON(http.StatusOK , gin.H{"error":"empty contest id"})
        return
    }
    
    var contest models.Contest 
    if err := database.DB.Model(&models.Contest{}).
                Where("id = ?",ContestID).
                Select("mcq").
                First(&contest).Error ; err != nil {
                    ctx.JSON(http.StatusBadRequest , gin.H{"error" : err.Error()})
                    return 
                }

    transaction_id , err := slimuuid.GenerateBest(os.Getenv("MAC_ADDRESS"))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError,gin.H{"error":err})
        return
    }

    ctx.JSON(http.StatusOK,gin.H{"message":transaction_id})

    go func(){
        bgCtx , cancel := context.WithCancel(context.Background())
        defer cancel() //does cleanup

        inits.RedisClient.Set(bgCtx,"progress:"+transaction_id,0,time.Hour * 24)
    
        var questions []models.Question
        if err:= database.DB.Model(&models.Question{}).
                    Where("mcq_or_code = ? AND contest_id=?","M",ContestID).
                    Select("options , id ").
                    Find(&questions).Error ; err != nil {
                        inits.RedisClient.Set(context.Background(),"progress:"+transaction_id,err.Error(),time.Hour*24)
                        cancel()
                        return
                    }

        questionAnswers := make(map[string]string)
        for _,question := range questions {
            questionAnswers[question.ID]=question.MCQ_Options[4]
        }

        var students []models.Student 
        if err := database.DB.Model(&models.Student{}).
                    Where("contest_id = ?",ContestID).
                    Select("id , score").
                    Find(&students).Error ; err != nil {
                        inits.RedisClient.Set(bgCtx,"progress:"+transaction_id,err.Error(),time.Hour*24)
                        return
                    }

        workers := make(chan struct{},5)
        var wg sync.WaitGroup
        completedCount := 0
        totalCount := len(students)
        var mu sync.Mutex
       
        for _ , student := range students {
            select{
                // if cancel called this will handel
            case <-bgCtx.Done():
                inits.RedisClient.Set(bgCtx , "progress:"+transaction_id,"Cancelled",time.Hour * 24)
                return

            case workers <-struct{}{}:
                wg.Add(1)
                go func(student models.Student){
                    defer wg.Done()
                    defer func(){
                        <-workers
                    }()

                    var mcqAnswers []models.McqAnswer
                    var score uint32 = 0
                    database.DB.Model(&models.McqAnswer{}).Where("student_id = ?",student.ID).Select("question_id , answer").Find(&mcqAnswers)
                    
                    for _ , answer := range mcqAnswers {
                        if answer.Answer == questionAnswers[answer.QuestionID] {
                            score++ 
                        }else if score > 0{
                            score--
                        }
                    }
                    mu.Lock()
                    database.DB.Model(&models.Student{}).Where("id=?",student.ID).Update("score" , student.Score + score)
                    completedCount++
                    progress := float64(completedCount)/float64(totalCount) * 100
                    if int(progress) %10 == 0 {
                        inits.RedisClient.Set(bgCtx , "progress:"+transaction_id , progress , time.Hour * 24)
                    }
                    mu.Unlock()
                }(student)
            }
          
        }  
        
        done := make(chan struct{})
        go func(){
            wg.Wait()
            close(done)
        }()

        select {
        case <-done :
            inits.RedisClient.Set(bgCtx , "progress:"+transaction_id,"Completed",time.Hour * 24)
            return
        case <-bgCtx.Done() :
            progress, _ := inits.RedisClient.Get(bgCtx, "progress:"+transaction_id).Result()
            inits.RedisClient.Set(bgCtx, "progress:"+transaction_id, 
            fmt.Sprintf("Progress stopped at %s due to context cancellation", progress), time.Hour*24)
            return
        }
    }()
}

