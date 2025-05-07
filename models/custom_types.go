package models
import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

func (res *Response) Value()(driver.Value , error){
	return json.Marshal(res)
}

func (res *Response) Scan(value interface{})error {
	bytes , ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes , res)
}

type ResponseSlice []Response

func (responses ResponseSlice) Value()(driver.Value , error){
	return json.Marshal(responses)
}

func (responses *ResponseSlice) Scan(value interface{})error {
	bytes , ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed")
	}
	return json.Unmarshal(bytes , responses)
}