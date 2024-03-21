package main

import (
	"flag"
	"os"

	"github.com/ananchev/xml-process/internal/logger"
	"github.com/ananchev/xml-process/internal/processor"
)

// Args command-line parameters
type Args struct {
	targetXML string
	logFile   string
}

func main() {
	args := ProcessArgs()
	logger.InitLogger(args.logFile)

	processor.Tranform(args.targetXML)

}

func ProcessArgs() Args {
	var a Args

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&a.targetXML, "xml", "", "Absolute filepath to the XML to process")
	f.StringVar(&a.logFile, "log", "", "If empty no logging. Allowed values stdout OR log filepath")

	f.Parse(os.Args[1:])

	return a

}
