package main

import (
	"log"
	"os"
)

// Create customs logs: errorFile and infoLog
func initLogs() (*log.Logger, *log.Logger) {

	errorFile, err := os.OpenFile("logs/webError.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	infoFile, err := os.OpenFile("logs/webInfo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	errorLog := log.New(errorFile, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(infoFile, "INFO\t", log.Ldate|log.Ltime)

	return errorLog, infoLog
}
