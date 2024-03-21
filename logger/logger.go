package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func InitLogger(loggingOption string) {
	var writer io.Writer

	if loggingOption == "" {
		writer = io.Discard
	} else if loggingOption == "stdout" {
		writer = os.Stdout
	} else {
		f, err := os.OpenFile(loggingOption, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		writer = f
		defer f.Close()
	}
	InfoLogger = log.New(writer, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(writer, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func LogError(format string, args ...interface{}) {
	write_to_log(1, format, args...)
}

func LogInfo(format string, args ...interface{}) {
	write_to_log(2, format, args...)
}

func write_to_log(loggerType int, format string, args ...interface{}) {
	log_msg := format_string(format, args...)
	switch loggerType {
	case 1:
		ErrorLogger.Println(log_msg)
	case 2:
		InfoLogger.Println(log_msg)
	}
}

func format_string(format string, args ...interface{}) string {
	args2 := make([]string, len(args))
	for i, v := range args {
		if i%2 != 0 {
			args2[i] = fmt.Sprint(v)
		} else {
			args2[i] = fmt.Sprintf("{%v}", v)
		}
	}
	r := strings.NewReplacer(args2...)
	return r.Replace(format)
}
