package log

import (
	"fmt"
	"github.com/go-errors/errors"
	"io"
	"os"
	"strings"
	"time"
)

//go:generate stringer -type=Level
type Level int8

const (
	Debug Level = iota
	Info
	Warning
	Error
)

type Logger struct {
	file       string
	fileWriter io.Writer
}

type ErrorStackProvider interface {
	Error() string
	ErrorStack() string
}

func New(fileName string) *Logger {
	fileNameExpanded := os.ExpandEnv(fileName)
	return &Logger{file: fileNameExpanded, fileWriter: noOp}
}

func (logger *Logger) fout() (io.Writer, error) {
	if logger.file != "" && logger.fileWriter == noOp {
		if writer, err := os.OpenFile(logger.file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644); err != nil {
			return nil, err
		} else {
			logger.fileWriter = writer
		}
	}
	return logger.fileWriter, nil
}

var noOp = &fakeWriter{}

type fakeWriter struct{}

func (w fakeWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (logger *Logger) Debugf(format string, v ...interface{}) {
	writer, err := logger.fout()
	if err != nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	level := strings.ToUpper(Debug.String())
	message := fmt.Sprintf(format, v...)
	fmt.Fprintf(writer, "%s %s %s", timestamp, level, message)
}

func (logger *Logger) Errorf(format string, v ...interface{}) {
	writer, errGetLogger := logger.fout()
	if errGetLogger != nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	level := strings.ToUpper(Error.String())
	var err ErrorStackProvider = nil
	for _, elem := range v {
		if errElem, ok := elem.(error); ok {
			if errors.As(errElem, &err) {
				break
			}
		}
	}
	message := fmt.Sprintf(format, v...)
	fmt.Fprintf(writer, "%s %s %s", timestamp, level, message)
	if err != nil {
		fmt.Fprintf(writer, "\n")
		fmt.Fprintf(writer, "=== Stack Trace Start ===\n")
		fmt.Fprintf(writer, "%s", err.ErrorStack())
		fmt.Fprintf(writer, "=== Stack Trace End ===\n")
	} else {
		fmt.Fprintf(writer, "\n")
	}
}
