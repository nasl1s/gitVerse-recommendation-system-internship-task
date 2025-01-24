package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp   string `json:"@timestamp"`
	LogLevel    string `json:"log_level"`
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Application string `json:"application"`
	Environment string `json:"environment"`
	Message     string `json:"message"`
}

type Logger struct {
	fileLogger  *log.Logger
	serviceName string
	application string
	environment string
	mu          sync.Mutex
}

func NewLogger(logFile, serviceName, application, environment string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open or create log file: %v", err)
	}

	return &Logger{
		fileLogger:  log.New(file, "", log.LstdFlags),
		serviceName: serviceName,
		application: application,
		environment: environment,
	}, nil
}


func (l *Logger) logMessage(level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		LogLevel:    level,
		ServiceName: l.serviceName,
		Host:        getHostname(),
		Application: l.application,
		Environment: l.environment,
		Message:     message,
	}

	logBytes, err := json.Marshal(logEntry)
	if err != nil {
		l.fileLogger.Printf(`{"error":"failed to marshal log entry: %v"}`, err)
		return
	}

	l.fileLogger.Println(string(logBytes))
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func (l *Logger) Print(v ...interface{}) {
	message := fmt.Sprint(v...)
	l.logMessage("INFO", message)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.logMessage("INFO", message)
}

func (l *Logger) Println(v ...interface{}) {
	message := fmt.Sprintln(v...)
	l.logMessage("INFO", message)
}

func (l *Logger) Fatal(v ...interface{}) {
	message := fmt.Sprint(v...)
	l.logMessage("FATAL", message)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.logMessage("FATAL", message)
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	message := fmt.Sprintln(v...)
	l.logMessage("FATAL", message)
	os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
	message := fmt.Sprint(v...)
	l.logMessage("PANIC", message)
	panic(message)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.logMessage("PANIC", message)
	panic(message)
}

func (l *Logger) Panicln(v ...interface{}) {
	message := fmt.Sprintln(v...)
	l.logMessage("PANIC", message)
	panic(message)
}
