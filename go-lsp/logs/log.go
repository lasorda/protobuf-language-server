package logs

import (
	"log"
	"os"
	"path"
)

const (
	logFileName        = ".protobuf-language-server.log"
	logBackUpExtension = ".bak"
)

var logger *log.Logger

// Init initializes logger with a logPath
//
// If logPath is nil then log file is created in os.UserHomeDir
func Init(logPath *string) {
	if logPath == nil || *logPath == "" {
		logger = stdErrLogger()
		return
	}
	l, err := fileLogger(*logPath)
	if err != nil {
		panic("error initializing file logger at " + *logPath)
	}
	logger = l
}

func stdErrLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

func fileLogger(path string) (*log.Logger, error) {
	if _, err := os.Stat(path); err == nil {
		os.Rename(path, path+logBackUpExtension)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	logger = log.New(f, "", 0)
	return logger, nil
}

func DefaultLogFilePath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, logFileName)
}

func Println(v ...interface{}) {
	logger.Println(v...)
}

func Printf(fmt string, v ...interface{}) {
	logger.Printf(fmt, v...)
}
