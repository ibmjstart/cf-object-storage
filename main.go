package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	verbex "github.com/VerbalExpressions/GoVerbalExpressions"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"github.com/ncw/swift"
	"github.ibm.com/ckwaldon/cf-large-objects/console_writer"
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

// isServiceFound returns true if the target service is present in the current space.
func isServiceFound(cliConnection plugin.CliConnection, targetService string) bool {
	// Get the services in the current space
	services, err := cliConnection.GetServices()
	checkErr(err)

	if len(services) < 1 {
		panic(errors.New("No services found in current space. (Check your internet connection)"))
	}

	// Check for target service in the list of present services
	found := false
	for _, service := range services {
		if service.Name == targetService {
			found = true
		}
	}

	return found
}

// getCredentialsName returns the name of the target service's credentials.
func getCredentialsName(cliConnection plugin.CliConnection, targetService string) string {
	// Get the service keys for the target service
	stdout, err := cliConnection.CliCommandWithoutTerminalOutput("service-keys", targetService)
	checkErr(err)

	// Construct regex to extract credentials name
	v := verbex.New().
		Find("\nname\n").
		BeginCapture().
		AnythingBut("\n").
		EndCapture().
		Captures(strings.Join(stdout, ""))

	// Get name of target service's credentials
	var serviceCredentialsName string
	if len(v) > 0 && len(v[0]) > 1 {
		serviceCredentialsName = v[0][1]
	} else {
		panic(errors.New("Could not find credentials for target service."))
	}

	return serviceCredentialsName
}

// getJSONCredentials returns the target service's credentials
func getJSONCredentials(cliConnection plugin.CliConnection, targetService, serviceCredentialsName string) string {
	// Get the service key for the target service's credentials
	stdout, err := cliConnection.CliCommandWithoutTerminalOutput("service-key", targetService, serviceCredentialsName)
	checkErr(err)

	// Construct regex to extract JSON
	v := verbex.New().
		AnythingBut("{").
		BeginCapture().
		Then("{").
		AnythingBut("}").
		Then("}").
		EndCapture().
		Captures(strings.Join(stdout, ""))

	// Get target service's credentials
	var serviceCredentialsJSON string
	if len(v) > 0 && len(v[0]) > 1 {
		serviceCredentialsJSON = v[0][1]
	} else {
		panic(errors.New("Could not fetch JSON credentials for target service."))
	}

	return serviceCredentialsJSON
}

// authenticate creates a connection to the target service
func authenticate(username, apiKey, authURL, domain, tenant string) swift.Connection {
	connection := swift.Connection{
		UserName:    username,
		ApiKey:      apiKey,
		AuthUrl:     authURL,
		Domain:      domain,
		Tenant:      tenant,
		AuthVersion: 3,
	}
	err := connection.Authenticate()
	checkErr(err)

	return connection
}

// getXAuthToken executes the logic to fetch the X-Auth token for an object storage instance.
func (c *LargeObjectsPlugin) getXAuthToken(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 2 {
		panic(errors.New("Incorrect Usage: " + c.GetMetadata().Commands[0].UsageDetails.Usage))
	}

	// Find and style username
	username, err := cliConnection.Username()
	checkErr(err)
	username = ansi.Color(username, "cyan+b")

	// Find and style org
	org, err := cliConnection.GetCurrentOrg()
	checkErr(err)
	orgStr := ansi.Color(org.OrganizationFields.Name, "cyan+b")

	// Find and style space
	space, err := cliConnection.GetCurrentSpace()
	checkErr(err)
	spaceStr := ansi.Color(space.SpaceFields.Name, "cyan+b")

	fmt.Printf("Fetching X-Auth token in org %s / space %s as %s...\n", orgStr, spaceStr, username)

	// begin console writer
	writer := console_writer.NewConsoleWriter()
	go writer.Write()

	// Handle Command Line arguments
	targetService := args[1]

	// Ensure that user is logged in
	if loggedIn, err := cliConnection.IsLoggedIn(); !loggedIn {
		panic(errors.New("You are not logged in. Run `cf login` and then rerun this command."))
	} else {
		checkErr(err)
	}

	// Find and display services. Ensure target service is within current space
	writer.SetCurrentStage("Searching for target service      ")
	found := isServiceFound(cliConnection, targetService)
	if !found {
		panic(errors.New("Service " + targetService + " not found in current space!"))
	}

	// Get service keys for target service
	writer.SetCurrentStage("Getting target service keys       ")
	serviceCredentialsName := getCredentialsName(cliConnection, targetService)

	// Fetch the JSON credentials
	writer.SetCurrentStage("Getting target service credentials")
	serviceCredentialsJSON := getJSONCredentials(cliConnection, targetService, serviceCredentialsName)

	// Parse the JSON credentials
	writer.SetCurrentStage("Parsing credentials               ")
	var credentials struct {
		Auth_URL   string
		DomainID   string
		DomainName string
		Password   string
		Project    string
		ProjectID  string
		Region     string
		Role       string
		UserID     string
		Username   string
	}
	err = json.Unmarshal([]byte(serviceCredentialsJSON), &credentials)
	checkErr(err)

	// Handle escaped unicode characters in JSON
	// see: https://github.com/cloudfoundry/cli/issues/794
	unescape := func(escaped string) string {
		return strings.Replace(
			strings.Replace(
				strings.Replace(
					escaped,
					"\u003c", "<", -1),
				"\u003e", ">", -1),
			"\u0026", "&", -1)
	}

	credentials.Auth_URL = unescape(credentials.Auth_URL)
	credentials.DomainID = unescape(credentials.DomainID)
	credentials.DomainName = unescape(credentials.DomainName)
	credentials.Password = unescape(credentials.Password)
	credentials.Project = unescape(credentials.Project)
	credentials.ProjectID = unescape(credentials.ProjectID)
	credentials.Region = unescape(credentials.Region)
	credentials.Role = unescape(credentials.Role)
	credentials.UserID = unescape(credentials.UserID)
	credentials.Username = unescape(credentials.Username)

	// Authenticate using service credentials
	writer.SetCurrentStage("Authenticating                    ")
	connection := authenticate(credentials.Username, credentials.Password, credentials.Auth_URL+"/v3", credentials.DomainName, "")

	// Print completion info
	writer.Quit()
	service := ansi.Color(targetService, "cyan+b")
	xAuth := ansi.Color("X-Auth:", "white+bh")
	fmt.Printf("%s\t%s %s\n", service, xAuth, connection.AuthToken)
}

// makeDLO executes the logic to create a Dynamic Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeDLO(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("making dlo")
}

// makeSLO executes the logic to create a Static Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeSLO(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("making slo")
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
