package logger

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

func Init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(msg string) {
	InfoLogger.Println(msg)
}

func Infof(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

func Error(msg string) {
	ErrorLogger.Println(msg)
}

func Errorf(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

func Debug(msg string) {
	DebugLogger.Println(msg)
}

func Debugf(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}

func Fatal(msg string) {
	ErrorLogger.Fatal(msg)
}

func Fatalf(format string, v ...interface{}) {
	ErrorLogger.Fatalf(format, v...)
}