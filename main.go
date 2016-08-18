package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

// pluginCommand defines the name of the command that this plugin creates.
// "oslo" is Object Storage Large Object
const pluginCommand string = "oslo"
const pluginName string = "LargeObjectsPlugin"

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
