package main

import (
	"flag"
	"os"
)

// Args command-line parameters
type Args struct {
	targetXML string
	logFile   string
}

func main() {
	args := ProcessArgs()

}

func ProcessArgs() Args {
	var a Args

	f := flag.NewFlagSet("Default", 1)
	f.StringVar(&a.targetXML, "xml", "", "Absolute filepath to the XML to process")
	f.StringVar(&a.logFile, "log", "none", "path to logfile. No logging takes place if not set")

	f.Parse(os.Args[1:])
	return a
}
