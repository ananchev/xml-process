package processxml

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ananchev/xml-process/processor"
)

// Args command-line parameters
type Args struct {
	targetXML string
	logFile   string
}

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func main() {
	args := ProcessArgs()

	var writer io.Writer
	if args.logFile == "" {
		writer = io.Discard
	} else if args.logFile == "stdout" {
		writer = os.Stdout
	} else {
		f, err := os.OpenFile(args.logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		writer = f
		defer f.Close()
	}
	InfoLogger = log.New(writer, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(writer, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	processor.TransformXML(args.targetXML)

}

func ProcessArgs() Args {
	var a Args

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&a.targetXML, "xml", "", "Absolute filepath to the XML to process")
	f.StringVar(&a.logFile, "log", "", "If empty no logging. Allowed values stdout OR log filepath")

	f.Parse(os.Args[1:])

	return a

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
		if i%2 == 0 {
			args2[i] = fmt.Sprintf("{%v}", v)
		} else {
			args2[i] = fmt.Sprint(v)
		}
	}
	r := strings.NewReplacer(args2...)
	return r.Replace(format)
}
