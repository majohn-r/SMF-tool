package main

import (
	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"os"
	"smf-tool/internal/commands"
)

var (
	// the following variables are set by the build process; the variable names
	// are known to that process, so do not casually change them
	version  = "unknown version!" // semantic version
	creation string               // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	appName  string               // the name of the application
	// defaultCommand string               // the name of the default command
	firstYear string // the year when development of this application began
	// these are variables in order to allow unit testing to inject
	// test-friendly functions
	exitFunc = os.Exit
	bus      = output.NewDefaultBus(tools.ProductionLogger)
)

func main() {
	commands.Load()
	exitCode := 0
	exitFunc(exitCode)
}
