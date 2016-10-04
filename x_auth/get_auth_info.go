package x_auth

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	verbex "github.com/VerbalExpressions/GoVerbalExpressions"
	"github.com/cloudfoundry/cli/plugin"
	"github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	"github.ibm.com/ckwaldon/swiftly-go/slo"
)

// flagVal holds the flag values
type flagVal struct {
	Url_flag    bool
	X_auth_flag bool
}

// credentials holds the info returned with a new cliConnection.
type credentials struct {
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

// isServiceFound returns true if the target service is present in the current space.
func isServiceFound(cliConnection plugin.CliConnection, targetService string) bool {
	// Get the services in the current space
	services, _ := cliConnection.GetServices()
	// checkErr(err)

	if len(services) < 1 {
		// panic(errors.New("No services found in current space. (Check your internet connection)"))
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
	stdout, _ := cliConnection.CliCommandWithoutTerminalOutput("service-keys", targetService)
	// checkErr(err)

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
		// panic(errors.New("Could not find credentials for target service."))
	}

	return serviceCredentialsName
}

// getJSONCredentials returns the target service's credentials
func getJSONCredentials(cliConnection plugin.CliConnection, targetService, serviceCredentialsName string) string {
	// Get the service key for the target service's credentials
	stdout, _ := cliConnection.CliCommandWithoutTerminalOutput("service-key", targetService, serviceCredentialsName)
	// checkErr(err)

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
		// panic(errors.New("Could not fetch JSON credentials for target service."))
	}

	return serviceCredentialsJSON
}

// extractFromJSON unmarshalls the JSON returned by a new cliConnection.
func extractFromJSON(serviceCredentialsJSON string) credentials {
	var creds credentials
	_ = json.Unmarshal([]byte(serviceCredentialsJSON), &creds)
	// checkErr(err)

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

	creds.Auth_URL = unescape(creds.Auth_URL)
	creds.DomainID = unescape(creds.DomainID)
	creds.DomainName = unescape(creds.DomainName)
	creds.Password = unescape(creds.Password)
	creds.Project = unescape(creds.Project)
	creds.ProjectID = unescape(creds.ProjectID)
	creds.Region = unescape(creds.Region)
	creds.Role = unescape(creds.Role)
	creds.UserID = unescape(creds.UserID)
	creds.Username = unescape(creds.Username)

	return creds
}

// ParseArgs reads the flags provided.
func ParseArgs(args []string) flagVal {
	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	url := flagSet.Bool("url", false, "output only the url")
	x_auth := flagSet.Bool("x", false, "output only the x-auth token")

	_ = flagSet.Parse(args[2:])

	flagVals := flagVal{
		Url_flag:    bool(*url),
		X_auth_flag: bool(*x_auth),
	}

	return flagVals
}

// DisplayUserInfo shows the username, org and space corresponding to the requested service.
func DisplayUserInfo(cliConnection plugin.CliConnection) {
	// Find username
	username, _ := cliConnection.Username()
	// checkErr(err)

	// Find org
	org, _ := cliConnection.GetCurrentOrg()
	// checkErr(err)

	// Find space
	space, _ := cliConnection.GetCurrentSpace()
	// checkErr(err)

	fmt.Printf("Fetching X-Auth info from org %s / space %s as %s...\n", console_writer.Cyan(org.Name), console_writer.Cyan(space.Name), console_writer.Cyan(username))
}

// GetAuthInfo executes the logic to fetch the auth URL and X-Auth token for an object storage instance.
func GetAuthInfo(cliConnection plugin.CliConnection, writer *console_writer.ConsoleWriter, targetService string) (string, string) {
	// Ensure that user is logged in
	if loggedIn, _ := cliConnection.IsLoggedIn(); !loggedIn {
		// panic(errors.New("You are not logged in. Run `cf login` and then rerun this command."))
	} else {
		// checkErr(err)
	}

	// Find and display services. Ensure target service is within current space
	writer.SetCurrentStage("Searching for target service      ")
	found := isServiceFound(cliConnection, targetService)
	if !found {
		// panic(errors.New("Service " + targetService + " not found in current space!"))
	}

	// Get service keys for target service
	writer.SetCurrentStage("Getting target service keys       ")
	serviceCredentialsName := getCredentialsName(cliConnection, targetService)

	// Fetch the JSON credentials
	writer.SetCurrentStage("Getting target service credentials")
	serviceCredentialsJSON := getJSONCredentials(cliConnection, targetService, serviceCredentialsName)

	// Parse the JSON credentials
	writer.SetCurrentStage("Parsing credentials               ")
	credentials := extractFromJSON(serviceCredentialsJSON)

	// Authenticate using service credentials
	writer.SetCurrentStage("Authenticating                    ")
	connection, _ := slo.Authenticate(credentials.Username, credentials.Password, credentials.Auth_URL+"/v3", credentials.DomainName, "")
	// checkErr(err)

	return connection.AuthUrl(), connection.AuthToken()
}
