package database

import (
    "gorm.io/driver/postgres" // or any other driver you're using
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "log"
    "os"
    "time"	
)


//DB is the database connection object
var DB *gorm.DB

func DBInit(){
	//get the database url from the environment variable
	dsn := os.Getenv("DB_URL")

	// Initialize the database with logging
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // Output to stdout
        logger.Config{
            SlowThreshold:             time.Second, // Threshold for slow queries
            LogLevel:                  logger.Info, // Log level
            IgnoreRecordNotFoundError: false,
            Colorful:                  true,
        },
    )

	//connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
        PrepareStmt: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	DB = db
}