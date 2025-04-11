package utils 

import (
	"io"
	"fmt"
	"bytes"
	"encoding/json"
)

type CompileRequest struct {
    Code string  `json:"code" binding:"required"`
    Lang string  `json:"lang" binding:"required"`
    Input string  `json:"input" binding:"required"`
}


func CompileCode(code string , lang string , input string) ([]byte , error) {
    compileReq := CompileRequest{
        Code : code ,
        Lang : lang ,
        Input : input ,
    }

    jsonBody , err := json.Marshal(compileReq) 
    if err != nil {
        return []byte(""), fmt.Errorf("failed to marshal request: %v", err)
    }

    resp , err := SendRequest("POST" ,"http://localhost:3000/api/v1/compile",bytes.NewReader(jsonBody))
    if err!=nil {
        return []byte(""), fmt.Errorf("failed to read response: %v", err)
    }

    respBody , err  := io.ReadAll(resp.Body)
    if err!=nil {
        return []byte(""), fmt.Errorf("failed to read response: %v", err)
    }
    defer resp.Body.Close()

    return respBody, nil

}