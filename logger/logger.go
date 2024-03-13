package logger

import (
	"log"
	"os"
	"sync"
)

var (
	logger Logger
	once   sync.Once
)

type Logger struct {
	Err, Info, Deb *log.Logger
}

func Get() Logger {
	once.Do(func() {
		logger = Logger{
			Err:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime),
			Info: log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
			Deb:  log.New(os.Stdout, "DEBUG\t", log.Ldate|log.Ltime|log.Lshortfile),
		}
	})

	return logger
}
