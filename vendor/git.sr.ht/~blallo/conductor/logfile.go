package conductor

import (
	"os"
	"sync"
)

var (
	logFilePath *string
	logFile     *os.File
	mu          sync.Mutex
)

func initLogFile() *os.File {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		return logFile
	}

	filePath := os.DevNull
	if logFilePath != nil {
		filePath = *logFilePath
	}

	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	logFile = f

	return f
}

// SetLogFile explicitly sets the file where to log the inner workings of the library.
// Panics if called more than once.
func SetLogFile(path string) {
	mu.Lock()
	defer mu.Unlock()

	if logFilePath == nil {
		logFilePath = &path
	} else {
		panic("calling SetLogFile more than once")
	}
}
