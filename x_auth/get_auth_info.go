package x_auth

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	verbex "github.com/VerbalExpressions/GoVerbalExpressions"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/ibmjstart/cf-object-storage/console_writer"
	"github.com/ibmjstart/swiftlygo/auth"
)

// flagVal holds the flag values.
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

// findService returns true if the target service is present in the current space.
func findService(cliConnection plugin.CliConnection, targetService string) error {
	// Get the services in the current space
	services, err := cliConnection.GetServices()
	if err != nil {
		return fmt.Errorf("Failed to get services in requested org: %s", err)
	}

	if len(services) < 1 {
		return fmt.Errorf("No services found in current space (check your internet connection)")
	}

	// Check for target service in the list of present services
	found := false
	for _, service := range services {
		if service.Name == targetService {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("Service '%s' not found in current space", targetService)
	}

	return nil
}

// getCredentialsName returns the name of the target service's credentials.
func getCredentialsName(cliConnection plugin.CliConnection, targetService string) (string, error) {
	// Get the service keys for the target service
	stdout, err := cliConnection.CliCommandWithoutTerminalOutput("service-keys", targetService)
	if err != nil {
		return "", fmt.Errorf("Failed to find credentials for service '%s': %s", targetService, err)
	}

	// Construct regex to extract credentials name
	v := verbex.New().
		Find("\nname\n").
		BeginCapture().
		AnythingBut("\n").
		EndCapture().
		Captures(strings.Join(stdout, "\n"))

	// Get name of target service's credentials
	var serviceCredentialsName string
	if len(v) > 0 && len(v[0]) > 1 {
		serviceCredentialsName = v[0][1]
	} else {
		return "", fmt.Errorf("Could not find credentials for target service")
	}

	return serviceCredentialsName, nil
}

// getJSONCredentials returns the target service's credentials.
func getJSONCredentials(cliConnection plugin.CliConnection, targetService, serviceCredentialsName string) (string, error) {
	// Get the service key for the target service's credentials
	stdout, err := cliConnection.CliCommandWithoutTerminalOutput("service-key", targetService, serviceCredentialsName)
	if err != nil {
		return "", fmt.Errorf("Failed to get credentials '%s' for service '%s': %s", serviceCredentialsName, targetService, err)
	}

	// Construct regex to extract JSON
	v := verbex.New().
		AnythingBut("{").
		BeginCapture().
		Then("{").
		Anything().
		StartOfLine().
		Then("}").
		EndCapture().
		Captures(strings.Join(stdout, ""))

	// Get target service's credentials
	var serviceCredentialsJSON string
	if len(v) > 0 && len(v[0]) > 1 {
		serviceCredentialsJSON = v[0][1]
	} else {
		return "", fmt.Errorf("Failed to fetch JSON credentials for service '%s'", targetService)
	}

	return serviceCredentialsJSON, nil
}

// extractFromJSON unmarshalls the JSON returned by a new cliConnection.
func extractFromJSON(serviceCredentialsJSON string) (*credentials, error) {
	var creds credentials
	err := json.Unmarshal([]byte(serviceCredentialsJSON), &creds)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshall JSON credentials: %s", err)
	}

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

	return &creds, nil
}

// ParseFlags reads the flags provided.
func ParseFlags(flags []string) (*flagVal, error) {
	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	url := flagSet.Bool("url", false, "Display auth url in quiet mode")
	x_auth := flagSet.Bool("x", false, "Display x-auth token in quiet mode")

	err := flagSet.Parse(flags)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse arguments: %s")
	}

	flagVals := flagVal{
		Url_flag:    bool(*url),
		X_auth_flag: bool(*x_auth),
	}

	return &flagVals, nil
}

// GetAuthInfo executes the logic to fetch the auth URL and X-Auth token for an object storage instance.
func GetAuthInfo(cliConnection plugin.CliConnection, writer *console_writer.ConsoleWriter, targetService string) (auth.Destination, error) {
	// Ensure that user is logged in
	if loggedIn, err := cliConnection.IsLoggedIn(); !loggedIn {
		return nil, fmt.Errorf("You are not logged in, please run `cf login` and rerun this command")
	} else if err != nil {
		return nil, fmt.Errorf("Failed to log in to Cloud Foundry: %s", err)
	}

	// Find and display services. Ensure target service is within current space
	writer.SetCurrentStage("Searching for target service")
	err := findService(cliConnection, targetService)
	if err != nil {
		return nil, err
	}

	// Get service keys for target service
	writer.SetCurrentStage("Locating target service's credentials")
	serviceCredentialsName, err := getCredentialsName(cliConnection, targetService)
	if err != nil {
		return nil, err
	}

	// Fetch the JSON credentials
	writer.SetCurrentStage("Fetching credentials")
	serviceCredentialsJSON, err := getJSONCredentials(cliConnection, targetService, serviceCredentialsName)
	if err != nil {
		return nil, err
	}

	// Parse the JSON credentials
	writer.SetCurrentStage("Parsing credentials")
	credentials, err := extractFromJSON(serviceCredentialsJSON)
	if err != nil {
		return nil, err
	}

	// Authenticate using service credentials
	writer.SetCurrentStage("Authenticating")
	destination, err := auth.Authenticate(credentials.Username, credentials.Password, credentials.Auth_URL+"/v3", credentials.DomainName, "")
	if err != nil {
		return nil, err
	}

	//return connection.AuthUrl(), connection.AuthToken(), nil
	return destination, nil
}

func Authenticate(cliConnection plugin.CliConnection, writer *console_writer.ConsoleWriter, targetService string) (auth.Destination, error) {
	// Get current user
	writer.SetCurrentStage("Locating service credentials")
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("Failed to get current user: %s", err)
	}

	// Find current user's home directory and construct path to credential file
	homeDir := currentUser.HomeDir
	logLocation := filepath.Join(homeDir, ".cf", "os_creds.json")

	// Create directory structure if necessary
	err = os.MkdirAll(filepath.Dir(logLocation), 0700)
	if err != nil {
		return nil, fmt.Errorf("Failed to create directory %s: %s", filepath.Dir(logLocation), err)
	}

	// Open or create credential file
	logFile, err := os.OpenFile(logLocation, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return nil, fmt.Errorf("Failed to open/create %s: %s", logLocation, err)
	}
	defer logFile.Close()

	// Get credential file's size
	logFileStat, err := logFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("Failed to get file size: %s", err)
	}

	var creds *credentials
	if logFileStat.Size() > 0 {
		// Read contents of credential file
		logContents := make([]byte, logFileStat.Size())
		_, err = logFile.Read(logContents)
		if err != nil {
			return nil, fmt.Errorf("Failed to read credentials: %s", err)
		}

		// Parse the JSON credentials
		writer.SetCurrentStage("Parsing credentials")
		creds, err = extractFromJSON(string(logContents))
		if err != nil {
			return nil, fmt.Errorf("Failed to extract JSON credentials: %s", err)
		}
	} else {
		creds, err = getNewCredentials(cliConnection, writer, targetService)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch a new set of credentials: %s", err)
		}
	}

	// Authenticate using service credentials
	writer.SetCurrentStage("Authenticating")
	destination, err := auth.Authenticate(creds.Username, creds.Password, creds.Auth_URL+"/v3", creds.DomainName, "")
	if err != nil {
		creds, err = getNewCredentials(cliConnection, writer, targetService)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch a new set of credentials: %s", err)
		}

		destination, err = auth.Authenticate(creds.Username, creds.Password, creds.Auth_URL+"/v3", creds.DomainName, "")
		if err != nil {
			return nil, fmt.Errorf("Failed to authenticate: %s", err)
		}
	}

	// Encode JSON credentials
	writer.SetCurrentStage("Saving credentials")
	marshalledCredentials, err := json.Marshal(creds)
	if err != nil {
		return nil, fmt.Errorf("Failed to JSON encode credentials: %s", err)
	}

	// Write credentails to file
	_, err = logFile.Write(marshalledCredentials)
	if err != nil {
		return nil, fmt.Errorf("Failed to write credentials to file: %s", err)
	}

	return destination, nil
}

func getNewCredentials(cliConnection plugin.CliConnection, writer *console_writer.ConsoleWriter, targetService string) (*credentials, error) {
	// Ensure that user is logged in
	if loggedIn, err := cliConnection.IsLoggedIn(); !loggedIn {
		return nil, fmt.Errorf("You are not logged in, please run `cf login` and rerun this command")
	} else if err != nil {
		return nil, fmt.Errorf("Failed to log in to Cloud Foundry: %s", err)
	}

	// Find and display services. Ensure target service is within current space
	writer.SetCurrentStage("Searching for target service")
	err := findService(cliConnection, targetService)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch services: %s", err)
	}

	// Get service keys for target service
	writer.SetCurrentStage("Locating target service's credentials")
	serviceCredentialsName, err := getCredentialsName(cliConnection, targetService)
	if err != nil {
		return nil, fmt.Errorf("Failed to locate target service's credentials: %s", err)
	}

	// Fetch the JSON credentials
	writer.SetCurrentStage("Fetching credentials")
	serviceCredentialsJSON, err := getJSONCredentials(cliConnection, targetService, serviceCredentialsName)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch target service's credentials: %s", err)
	}

	// Parse the JSON credentials
	writer.SetCurrentStage("Parsing credentials")
	credentials, err := extractFromJSON(serviceCredentialsJSON)
	if err != nil {
		return nil, fmt.Errorf("Failed to extract JSON credentials: %s", err)
	}

	return credentials, nil
}
