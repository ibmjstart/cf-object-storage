package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// pluginCommand defines the name of the command that this plugin creates.
// "oslo" is Object Storage Large Object
const pluginCommand string = "oslo"
const pluginName string = "LargeObjectsPlugin"

// LargeObjectsPlugin is the struct implementing the interface defined by the core CLI.
type LargeObjectsPlugin struct{}

// checkErr panics if given an error and otherwise does nothing
func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

// Run handles each invocation of the CLI plugin.
func (c *LargeObjectsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Catch any panic statements from here onward
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Fatal Error: ", r)
			os.Exit(1)
		}
	}()

	// Ensure that we called the command basic-plugin-command
	if args[0] != pluginCommand {
		panic(errors.New("Invocation error!\n" + pluginName + " called with args: " + strings.Join(args, " ")))
	}
	fmt.Println("Running large objects plugin")
	if len(args) < 2 {
		panic(errors.New("Incorrect Usage: " + c.GetMetadata().Commands[0].UsageDetails.Usage))
	}

	// Handle Command Line arguments
	targetService := args[1]

	// Ensure that user is logged in
	if logged_in, err := cliConnection.IsLoggedIn(); !logged_in {
		panic(errors.New("You are not logged in. Run `cf login` and then rerun this command."))
	} else {
		checkErr(err)
	}
	if args[1] == "command" {
		fmt.Println("Invoking the large objects plugin command.")
	}
}

// GetMetadata() returns a PluginMetadata struct with information
// about the current version of this plugin and how to use it. This
// information is used to build the CF CLI helptext for this plugin's
// commands.
func (c *LargeObjectsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: pluginName,
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     pluginCommand,
				HelpText: "LargeObjects plugin command's help text",
				UsageDetails: plugin.Usage{
					Usage: "command\n   cf " + pluginCommand + " [args]",
				},
			},
		},
	}
}

// main initializes a plugin on install, but is not invoked when that plugin
// is run from the CLI. See Run() function for logic invoked when CLI plugin
// is actually used.
func main() {
	// Any initialization for your plugin can be handled here
	plugin.Start(new(LargeObjectsPlugin))
}
