package commands

// this variable and the functions that manipulate it provide a mechanism for
// main to punch into this package, which causes the package to be loaded, and
// the various command init functions to be executed. Without this, there are no
// commands to run.

const defaultCommand = "read"

// IsDefault is called by the init function of each command, so it can state
// whether it's the default command - and without hardcoding that fact into one
// of the commands. Instead, it's chosen by main calling DeclareDefault().
func IsDefault(commandName string) bool {
	return commandName == defaultCommand
}

// Load is meant to be called by main(), to load the commands package
func Load() {
	// does nothing
}
