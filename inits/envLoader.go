package inits

import (
	"github.com/joho/godotenv"
	"github.com/theMitocondria/Unbiass/utils"
)

func LoadEnv(){
	err := godotenv.Load()
	utils.LogError(err,"Error connecting to ENV")
}