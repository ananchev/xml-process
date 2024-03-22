package processor

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

func InitLogger(logfile string) {
	var writer io.Writer
	if logfile == "" {
		writer = io.Discard
	} else if logfile == "stdout" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal(err)
		}
		writer = file

	}
	InfoLogger = log.New(writer, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(writer, "ERROR: ", log.Ldate|log.Ltime)
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
