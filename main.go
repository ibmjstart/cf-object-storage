package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	verbex "github.com/VerbalExpressions/GoVerbalExpressions"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"github.com/ncw/swift"
)

var (
	curStep string
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

func getServiceKeys(cliConnection plugin.CliConnection, targetService string) string {
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
		// fmt.Println("Service Creds Name: " + serviceCredentialsName)
	} else {
		panic(errors.New("Could not find credentials for target service."))
	}

	return serviceCredentialsName
}

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
	// fmt.Println(connection)
	checkErr(err)
	// fmt.Println("Authenticated!")
	return connection
}

// getXAuthToken executes the logic to fetch the X-Auth token for an object storage instance.
func (c *LargeObjectsPlugin) getXAuthToken(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 2 {
		panic(errors.New("Incorrect Usage: " + c.GetMetadata().Commands[0].UsageDetails.Usage))
	}

	// Find and display username
	username, err := cliConnection.Username()
	checkErr(err)
	username = ansi.Color(username, "cyan+b")
	// fmt.Println("Username: ", username)

	// Find and display org
	org, err := cliConnection.GetCurrentOrg()
	checkErr(err)
	orgStr := ansi.Color(org.OrganizationFields.Name, "cyan+b")
	// fmt.Println("Current Org: " + org.OrganizationFields.Name)

	// Find and display space
	space, err := cliConnection.GetCurrentSpace()
	checkErr(err)
	spaceStr := ansi.Color(space.SpaceFields.Name, "cyan+b")
	// fmt.Println("Current Space: " + space.SpaceFields.Name)

	fmt.Printf("Fetching X-Auth Token in org %s / space %s as %s...\n", orgStr, spaceStr, username)

	curStep = "Starting                          "

	// begin console writer
	quit := make(chan int)
	go consoleWriter(quit)

	// Handle Command Line arguments
	targetService := args[1]

	// Ensure that user is logged in
	if loggedIn, err := cliConnection.IsLoggedIn(); !loggedIn {
		panic(errors.New("You are not logged in. Run `cf login` and then rerun this command."))
	} else {
		checkErr(err)
	}

	curStep = "Searching for target service     "
	// Find and display services. Ensure target service is within current space
	services, err := cliConnection.GetServices()
	checkErr(err)

	if len(services) < 1 {
		panic(errors.New("No services found in current space. (Check your internet connection)"))
	}

	// fmt.Println("Services:")
	found := false
	for _, service := range services {
		// fmt.Println("\t" + service.Name)
		if service.Name == targetService {
			found = true
		}
	}
	if !found {
		panic(errors.New("Service " + targetService + " not found in current space!"))
	}

	curStep = "Getting target service credentials"
	serviceCredentialsName := getServiceKeys(cliConnection, targetService)

	// Fetch the JSON credentials
	stdout, err := cliConnection.CliCommandWithoutTerminalOutput("service-key", targetService, serviceCredentialsName)
	checkErr(err)
	v := verbex.New().
		AnythingBut("{").
		BeginCapture().
		Then("{").
		AnythingBut("}").
		Then("}").
		EndCapture().
		Captures(strings.Join(stdout, ""))
	var serviceCredentialsJSON string
	if len(v) > 0 && len(v[0]) > 1 {
		serviceCredentialsJSON = v[0][1]
		// fmt.Println("Service Creds JSON: " + serviceCredentialsJSON)
	} else {
		panic(errors.New("Could not fetch JSON credentials for target service."))
	}

	curStep = "Parsing credentials               "

	// Parse the JSON credentials
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
	// fmt.Println(credentials)

	curStep = "Authenticating                    "
	connection := authenticate(credentials.Username, credentials.Password, credentials.Auth_URL+"/v3", credentials.DomainName, "")

	quit <- 0
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

func printCompletionInfo() {

}

func consoleWriter(quit chan int) {
	count := 0
	for {
		currentStep := curStep
		loading := [4]string{"*   ", " *  ", "  * ", "   *"}
		select {
		case <-quit:
			ok := ansi.Color("OK", "green+b")
			fmt.Printf("\r%s                                     ", ok)
			fmt.Println("\n")
			return
		default:
			switch count = (count + 1) % 6; count {
			case 0:
				fmt.Printf("\r %s %s", loading[0], currentStep)
			case 1, 5:
				fmt.Printf("\r %s %s", loading[1], currentStep)
			case 2, 4:
				fmt.Printf("\r %s %s", loading[2], currentStep)
			case 3:
				fmt.Printf("\r %s %s", loading[3], currentStep)
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
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
