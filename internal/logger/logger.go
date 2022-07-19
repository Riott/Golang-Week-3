package logger

import (
	"log"
	"os"
)

func InitLogger() *log.Logger {
	file, err := os.OpenFile("./logs/server.log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644)
	if err != nil {
		log.Fatal(err)
	}
	serverLogger := log.New(file, "SERVER: ", log.Ldate|log.Ltime|log.Lshortfile)
	return serverLogger
}
