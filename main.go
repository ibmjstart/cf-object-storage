package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/plugin"
	"github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	"github.ibm.com/ckwaldon/cf-large-objects/x_auth"
)

// getXAuthCommand defines the name of the command that fetches X-Auth Tokens.
const getAuthInfoCommand string = "get-auth-info"

// makeDLOCommand defines the name of the command that creates DLOs in
// object storage.
const makeDLOCommand string = "make-dlo"

// makeDLOCommand defines the name of the command that creates SLOs in
// object storage.
const makeSLOCommand string = "make-slo"

// pluginName defines the name of this plugin for use installing and
// uninstalling it.
const pluginName string = "ObjectStorageLargeObjects"

// LargeObjectsPlugin is the struct implementing the plugin interface.
// It has no public members.
type LargeObjectsPlugin struct {
	subcommands map[string](func(plugin.CliConnection, []string) error)
}

// Run handles each invocation of the CLI plugin.
func (c *LargeObjectsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Associate each subcommand with a handler function
	c.subcommands = map[string](func(plugin.CliConnection, []string) error){
		getAuthInfoCommand: c.getAuthInfo,
		makeDLOCommand:     c.makeDLO,
		makeSLOCommand:     c.makeSLO,
	}

	// Dispatch the subcommand that the user wanted, if it exists
	subcommandFunc := c.subcommands[args[0]]
	err := subcommandFunc(cliConnection, args)

	// Check for an error
	if err != nil {
		fmt.Printf("\r\033[2K\n%s\n%s\n", console_writer.Red("FAILED"), err)
		os.Exit(1)
	}
}

// getXAuthToken executes the logic to fetch the X-Auth token for an object storage instance.
func (c *LargeObjectsPlugin) getAuthInfo(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 2 {
		return fmt.Errorf("Missing service name\nUsage: %s", c.GetMetadata().Commands[0].UsageDetails.Usage)
	}

	// Parse flags
	flags, err := x_auth.ParseFlags(args[2:])
	if err != nil {
		return err
	}

	quiet := flags.Url_flag || flags.X_auth_flag
	writer := console_writer.NewConsoleWriter()

	// Start console writer if not in quiet mode
	if !quiet {
		err := x_auth.DisplayUserInfo(cliConnection)
		if err != nil {
			return err
		}

		go writer.Write()
	}

	// Get authorization info
	authUrl, xAuth, err := x_auth.GetAuthInfo(cliConnection, writer, args[1])
	if err != nil {
		return err
	}

	// Print requested attributes
	if flags.Url_flag {
		fmt.Println(authUrl)
	}
	if flags.X_auth_flag {
		fmt.Println(xAuth)
	}

	// Kill console writer if not in quiet mode
	if !quiet {
		writer.Quit()

		fmt.Printf("\r\033[2K%s\n\n%s\n%s %s\n%s %s\n", console_writer.Green("OK"), console_writer.Cyan(args[1]), console_writer.White("auth url:"), authUrl, console_writer.White("x-auth:  "), xAuth)
	}

	return nil
}

// makeDLO executes the logic to create a Dynamic Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeDLO(cliConnection plugin.CliConnection, args []string) error {
	fmt.Println("making dlo")
	return nil
}

// makeSLO executes the logic to create a Static Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeSLO(cliConnection plugin.CliConnection, args []string) error {
	fmt.Println("making slo")
	return nil
}

// GetMetadata returns a PluginMetadata struct with information
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
				Name:     getAuthInfoCommand,
				HelpText: "Display an Object Storage service's authentication url and x-auth token",
				UsageDetails: plugin.Usage{
					Usage: "cf " + getAuthInfoCommand + " SERVICE_NAME [-url] [-x]",
					Options: map[string]string{
						"url": "Display auth url in quiet mode",
						"x":   "Display x-auth token in quiet mode",
					},
				},
			},
			{
				Name:     makeDLOCommand,
				HelpText: "LargeObjects plugin command's help text",
				UsageDetails: plugin.Usage{
					Usage: "cf " + makeDLOCommand + " [args]",
				},
			},
			{
				Name:     makeSLOCommand,
				HelpText: "LargeObjects plugin command's help text",
				UsageDetails: plugin.Usage{
					Usage: "cf " + makeSLOCommand + " [args]",
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
