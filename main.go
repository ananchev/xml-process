package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ananchev/processxml/processor"
)

// Args command-line parameters
type Args struct {
	targetXML string
	logfile   string
}

var args Args

func main() {

	requiredFlags := []string{"xml"}

	args.targetXML = *flag.String("xml", "", "Absolute filepath to the XML to process")
	args.logfile = *flag.String("log", "", "Logging disabled if skipped. Allowed values stdout OR filepath")

	flag.Parse()
	suppliedFlags := make(map[string]bool)
	flag.Visit(func(fl *flag.Flag) { suppliedFlags[fl.Name] = true })

	for _, req := range requiredFlags {
		if !suppliedFlags[req] {
			fmt.Fprintf(os.Stderr, "Missing required -%s flag.\nRun %s -h for usage.\n", req, os.Args[0])
			os.Exit(2) // the same exit code flag.Parse uses
		}
	}
	flag.VisitAll(setArgs)

	processor.TransformXML(args.logfile, args.targetXML)
}

func setArgs(fl *flag.Flag) {
	switch fl.Name {
	case "xml":
		args.targetXML = fl.Value.String()

	case "log":
		args.logfile = fl.Value.String()
	}
}
