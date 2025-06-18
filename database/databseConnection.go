package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
    "github.com/theMitocondria/Unbiass/inits"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
    if err := inits.LoadEnv(); err != nil {
		return nil, err
	}
    
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		return nil, ErrMissingDSN
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}

var ErrMissingDSN = &ConfigError{"missing DB_URL environment variable"}

type ConfigError struct {
	s string
}

func (e *ConfigError) Error() string {
	return e.s
}
