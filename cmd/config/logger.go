package config

import (
	"log"
	"os"
)

type Logger struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func InitLogger() Logger {
	logger := Logger{}

	logger.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	logger.ErrorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return logger
}
