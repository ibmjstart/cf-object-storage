package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	verbex "github.com/VerbalExpressions/GoVerbalExpressions"
	"github.com/cloudfoundry/cli/plugin"
)

// getXAuthCommand defines the name of the command that fetches X-Auth Tokens.
const getXAuthCommand string = "get-x-auth"

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
	subcommands map[string]func(plugin.CliConnection, []string)
}

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

	// Associate each subcommand with a handler function
	c.subcommands = map[string]func(plugin.CliConnection, []string){
		getXAuthCommand: c.getXAuthToken,
		makeDLOCommand:  c.makeDLO,
		makeSLOCommand:  c.makeSLO,
	}

	// Dispatch the subcommand that the user wanted, if it exists
	if subcommandFunc, keyExists := c.subcommands[args[0]]; !keyExists {
		panic(errors.New("Invocation error!\n" + pluginName + " called with args: " + strings.Join(args, " ")))
	} else {
		subcommandFunc(cliConnection, args)
	}
}

// getXAuthToken executes the logic to fetch the X-Auth token for an object storage instance.
func (c *LargeObjectsPlugin) getXAuthToken(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 2 {
		panic(errors.New("Incorrect Usage: " + c.GetMetadata().Commands[0].UsageDetails.Usage))
	}
	fmt.Println("Fetching X-Auth Token")

	// Handle Command Line arguments
	targetService := args[1]

	// Ensure that user is logged in
	if logged_in, err := cliConnection.IsLoggedIn(); !logged_in {
		panic(errors.New("You are not logged in. Run `cf login` and then rerun this command."))
	} else {
		checkErr(err)
	}

	fmt.Println("Discovering User Information...")
	// Find and display username
	username, err := cliConnection.Username()
	checkErr(err)
	fmt.Println("Username: ", username)

	// Find and display org
	org, err := cliConnection.GetCurrentOrg()
	checkErr(err)
	fmt.Println("Current Org: " + org.OrganizationFields.Name)

	// Find and display space
	space, err := cliConnection.GetCurrentSpace()
	checkErr(err)
	fmt.Println("Current Space: " + space.SpaceFields.Name)

	fmt.Println("Searching for target service...")
	// Find and display services. Ensure target service is within current space
	services, err := cliConnection.GetServices()
	checkErr(err)

	if len(services) < 1 {
		panic(errors.New("No services found in current space. (Check your internet connection)"))
	}

	fmt.Println("Services:")
	found := false
	for _, service := range services {
		fmt.Println("\t" + service.Name)
		if service.Name == targetService {
			found = true
		}
	}
	if !found {
		panic(errors.New("Service " + targetService + " not found in current space!"))
	}

	fmt.Println("Getting target service credentials...")
	// Get service keys for target service
	stdout, err := cliConnection.CliCommandWithoutTerminalOutput("service-keys", targetService)
	checkErr(err)
	v := verbex.New().
		Find("\nname\n").
		BeginCapture().
		AnythingBut("\n").
		EndCapture().
		Captures(strings.Join(stdout, ""))
	var serviceCredentialsName string
	if len(v) > 0 && len(v[0]) > 1 {
		serviceCredentialsName = v[0][1]
		fmt.Println("Service Creds Name: " + serviceCredentialsName)
	} else {
		panic(errors.New("Could not find credentials for target service."))
	}
}

// makeDLO executes the logic to create a Dynamic Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeDLO(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("making dlo")
}

// makeSLO executes the logic to create a Static Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeSLO(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("making slo")
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
				Name:     getXAuthCommand,
				HelpText: "LargeObjects plugin command's help text",
				UsageDetails: plugin.Usage{
					Usage: "command\n   cf " + getXAuthCommand + " [args]",
				},
			},
			{
				Name:     makeDLOCommand,
				HelpText: "LargeObjects plugin command's help text",
				UsageDetails: plugin.Usage{
					Usage: "command\n   cf " + makeDLOCommand + " [args]",
				},
			},
			{
				Name:     makeSLOCommand,
				HelpText: "LargeObjects plugin command's help text",
				UsageDetails: plugin.Usage{
					Usage: "command\n   cf " + makeSLOCommand + " [args]",
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
