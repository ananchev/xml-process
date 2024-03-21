package main

import (
	"flag"
	"os"

	"github.com/ananchev/processxml/logger"
	"github.com/ananchev/processxml/processor"
)

// Args command-line parameters
type Args struct {
	targetXML     string
	loggingOption string
}

func main() {
	args := ProcessArgs()
	logger.InitLogger(args.loggingOption)
	processor.TransformXML(args.targetXML)
}

func ProcessArgs() Args {
	var a Args

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&a.targetXML, "xml", "", "Absolute filepath to the XML to process")
	f.StringVar(&a.loggingOption, "log", "", "If empty no logging. Allowed values stdout OR log filepath")

	f.Parse(os.Args[1:])

	return a

}
