package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	InfoLogger  *log.Logger
	DebugLogger *log.Logger
	ErrorLogger *log.Logger
)

func InitLogger(logfile string) {
	var multi_writer io.Writer
	if logfile == "" {
		multi_writer = io.MultiWriter(os.Stdout)
	} else {
		file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal(err)
		}
		multi_writer = io.MultiWriter(os.Stdout, file)

	}
	InfoLogger = log.New(multi_writer, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(multi_writer, "ERROR: ", log.Ldate|log.Ltime)
	DebugLogger = log.New(multi_writer, "DEBUG: ", log.Ldate|log.Ltime)
}

func format_string(format string, args ...interface{}) string {
	args2 := make([]string, len(args))
	for i, v := range args {
		if i%2 == 0 {
			args2[i] = fmt.Sprintf("{%v}", v)
		} else {
			args2[i] = fmt.Sprint(v)
		}
	}
	r := strings.NewReplacer(args2...)
	return r.Replace(format)
}

func write_to_log(loggerType int, format string, args ...interface{}) {
	_, fn, line, _ := runtime.Caller(1)
	// below adds caller info to the string to be logged
	format = filepath.Base(fn) + ":" + strconv.Itoa(line) + ": " + format
	log_msg := format_string(format, args...)
	switch loggerType {
	case 1:
		ErrorLogger.Println(log_msg)
	case 2:
		InfoLogger.Println(log_msg)
	}
}

func Error(format string, args ...interface{}) {
	write_to_log(1, format, args...)
}

func Info(format string, args ...interface{}) {
	write_to_log(2, format, args...)
}
