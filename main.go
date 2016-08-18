package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

// LargeObjectsPlugin is the struct implementing the interface defined by the core CLI.
type LargeObjectsPlugin struct{}

// Run handles each invocation of the CLI plugin.
func (c *LargeObjectsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command basic-plugin-command
	if args[0] == "lo" {
		fmt.Println("Running the large objects plugin")
	}
	if args[1] == "command" {
		fmt.Println("Invoking the large objects plugin command.")
	}
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
func (c *LargeObjectsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "LargeObjectsPlugin",
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
				Name:     "lo command",
				HelpText: "LargeObjects plugin command's help text",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "command\n   cf lo command",
				},
			},
		},
	}
}

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "github.com/cloudfoundry/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(LargeObjectsPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
