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
	cw "github.com/ibmjstart/cf-object-storage/console_writer"
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

// authInfo holds the info used to restablish an existing connection.
type authInfo struct {
	AuthToken  string
	Service    string
	StorageUrl string
}

// authenticator holds the info required to authenticate with Object Storage.
type authenticator struct {
	cliConnection plugin.CliConnection
	authInfo      authInfo
	creds         credentials

	writer        *cw.ConsoleWriter
	flagVals      flagVal
	targetService string
	doSave        bool

	logFile     *os.File
	logFileSize int64
}

// findService returns true if the target service is present in the current space.
func (a *authenticator) findService() error {
	// Get the services in the current space
	services, err := a.cliConnection.GetServices()
	if err != nil {
		return fmt.Errorf("Failed to get services in requested org: %s", err)
	}

	if len(services) < 1 {
		return fmt.Errorf("No services found in current space (check your internet connection)")
	}

	// Check for target service in the list of present services
	found := false
	for _, service := range services {
		if service.Name == a.targetService {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("Service '%s' not found in current space", a.targetService)
	}

	return nil
}

// getCredentialsName returns the name of the target service's credentials.
func (a *authenticator) getCredentialsName() (string, error) {
	// Get the service keys for the target service
	stdout, err := a.cliConnection.CliCommandWithoutTerminalOutput("service-keys", a.targetService)
	if err != nil {
		return "", fmt.Errorf("Failed to find credentials for service '%s': %s", a.targetService, err)
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
func (a *authenticator) getJSONCredentials(serviceCredentialsName string) (string, error) {
	// Get the service key for the target service's credentials
	stdout, err := a.cliConnection.CliCommandWithoutTerminalOutput("service-key", a.targetService, serviceCredentialsName)
	if err != nil {
		return "", fmt.Errorf("Failed to get credentials '%s' for service '%s': %s", serviceCredentialsName, a.targetService, err)
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
		return "", fmt.Errorf("Failed to fetch JSON credentials for service '%s'", a.targetService)
	}

	return serviceCredentialsJSON, nil
}

// unescape handles escaped unicode characters in JSON
// see: https://github.com/cloudfoundry/cli/issues/794
func unescape(escaped string) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				escaped,
				"\u003c", "<", -1),
			"\u003e", ">", -1),
		"\u0026", "&", -1)
}

// extractAuthFromJSON unmarshalls the JSON saved with a successful connection.
func (a *authenticator) extractAuthFromJSON(authInfoJSON string) error {
	var authInfo authInfo
	err := json.Unmarshal([]byte(authInfoJSON), &authInfo)
	if err != nil {
		return fmt.Errorf("Failed to unmarshall authentication info: %s", err)
	}

	authInfo.AuthToken = unescape(authInfo.AuthToken)
	authInfo.StorageUrl = unescape(authInfo.StorageUrl)

	a.authInfo = authInfo

	return nil
}

// extractCredsFromJSON unmarshalls the JSON returned by a new cliConnection.
func (a *authenticator) extractCredsFromJSON(serviceCredentialsJSON string) error {
	var creds credentials
	err := json.Unmarshal([]byte(serviceCredentialsJSON), &creds)
	if err != nil {
		return fmt.Errorf("Failed to unmarshall JSON credentials: %s", err)
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

	a.creds = creds

	return nil
}

// getNewCredentials fetches a new set of authentication credentials from Object Storage.
func (a *authenticator) getNewCredentials() error {
	// Ensure that user is logged in
	if loggedIn, err := a.cliConnection.IsLoggedIn(); !loggedIn {
		return fmt.Errorf("You are not logged in, please run `cf login` and rerun this command")
	} else if err != nil {
		return fmt.Errorf("Failed to log in to Cloud Foundry: %s", err)
	}

	// Find and display services. Ensure target service is within current space
	a.writer.SetCurrentStage("Searching for target service")
	err := a.findService()
	if err != nil {
		return fmt.Errorf("Failed to fetch services: %s", err)
	}

	// Get service keys for target service
	a.writer.SetCurrentStage("Locating target service's credentials")
	serviceCredentialsName, err := a.getCredentialsName()
	if err != nil {
		return fmt.Errorf("Failed to locate target service's credentials: %s", err)
	}

	// Fetch the JSON credentials
	a.writer.SetCurrentStage("Fetching credentials")
	serviceCredentialsJSON, err := a.getJSONCredentials(serviceCredentialsName)
	if err != nil {
		return fmt.Errorf("Failed to fetch target service's credentials: %s", err)
	}

	// Parse the JSON credentials
	a.writer.SetCurrentStage("Parsing credentials")
	err = a.extractCredsFromJSON(serviceCredentialsJSON)
	if err != nil {
		return fmt.Errorf("Failed to extract JSON credentials: %s", err)
	}

	// Ensure new credentails are saved
	a.doSave = true

	return nil
}

// getSavedCredentials loads the locally saved credentails.
func (a *authenticator) getSavedCredentials() error {
	a.writer.SetCurrentStage("Locating service credentials")

	// Get current user
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("Failed to get current user: %s", err)
	}

	// Find current user's home directory and construct path to credential file
	homeDir := currentUser.HomeDir
	logLocation := filepath.Join(homeDir, ".cf", "os_creds.json")

	// Create directory structure if necessary
	err = os.MkdirAll(filepath.Dir(logLocation), 0700)
	if err != nil {
		return fmt.Errorf("Failed to create directory %s: %s", filepath.Dir(logLocation), err)
	}

	// Open or create credential file
	logFile, err := os.OpenFile(logLocation, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return fmt.Errorf("Failed to open/create %s: %s", logLocation, err)
	}

	// Get credential file's size
	logFileStat, err := logFile.Stat()
	if err != nil {
		return fmt.Errorf("Failed to get file size: %s", err)
	}

	a.logFile = logFile
	a.logFileSize = logFileStat.Size()

	return nil
}

// saveCredentials writes the credentails to a local file.
func (a *authenticator) saveCredentials() error {
	// Encode JSON credentials
	marshalledCredentials, err := json.Marshal(a.authInfo)
	if err != nil {
		return fmt.Errorf("Failed to JSON encode authentication info: %s", err)
	}

	// Write credentails to file
	err = a.logFile.Truncate(0)
	if err != nil {
		return fmt.Errorf("Failed to truncate file: %s", err)
	}
	_, err = a.logFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("Failed to set file offset to 0: %s", err)
	}
	_, err = a.logFile.Write(marshalledCredentials)
	if err != nil {
		return fmt.Errorf("Failed to write authentication info to file: %s", err)
	}

	return nil
}

// parseFlags reads the flags provided.
func parseFlags(args []string) (*flagVal, error) {
	flags := args[3:]
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

// Authenticate authenticates the current session with Object Storage and saves the credentails.
func Authenticate(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, targetService string) (auth.Destination, error) {
	var (
		destination auth.Destination
		a           = authenticator{
			cliConnection: cliConnection,
			writer:        writer,
			targetService: targetService,
			doSave:        false,
		}
	)

	// Check for and get saved service credentials
	err := a.getSavedCredentials()
	if err != nil {
		return nil, fmt.Errorf("Failed to get saved credentials: %s", err)
	}
	defer a.logFile.Close()

	// Read contents of credential file
	logContents := make([]byte, a.logFileSize)
	bytesRead, err := a.logFile.Read(logContents)
	if err != nil {
		return nil, fmt.Errorf("Failed to read credentials: %s", err)
	}

	// Extract authentication info from file, if not empty
	if bytesRead > 0 {
		err = a.extractAuthFromJSON(string(logContents))
		if err != nil {
			return nil, fmt.Errorf("Failed to parse JSON: %s", err)
		}
	}

	// Determine if the found authentication corresponds to the target server
	isTargetService := a.authInfo.Service == a.targetService

	// Authenticate using service credentials
	a.writer.SetCurrentStage("Authenticating")

	if isTargetService {
		destination, err = auth.AuthenticateWithToken(a.authInfo.AuthToken, a.authInfo.StorageUrl)
	}

	if !isTargetService || err != nil {
		err = a.getNewCredentials()
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch a new set of credentials (Try running `cf login`): %s", err)
		}

		destination, err = auth.Authenticate(a.creds.Username, a.creds.Password, a.creds.Auth_URL+"/v3", a.creds.DomainName, "")
		if err != nil {
			return nil, fmt.Errorf("Failed to authenticate: %s", err)
		}

		a.authInfo.AuthToken = destination.(*auth.SwiftDestination).SwiftConnection.AuthToken
		a.authInfo.StorageUrl = destination.(*auth.SwiftDestination).SwiftConnection.StorageUrl
		a.authInfo.Service = a.targetService
	}

	// Save the credentials, if necessary
	if a.doSave {
		a.writer.SetCurrentStage("Saving credentials")
		err = a.saveCredentials()
		if err != nil {
			return nil, fmt.Errorf("Failed to save credentials: %s", err)
		}
	}

	return destination, nil
}

// DisplayAuthInfo prints the requested values.
func DisplayAuthInfo(destination auth.Destination, writer *cw.ConsoleWriter, args []string) (string, error) {
	writer.SetCurrentStage("Fetching authentication info")

	serviceName := args[2]
	flagVals, err := parseFlags(args)
	if err != nil {
		return "", fmt.Errorf("Failed to parse flags: %s", err)
	}

	result := fmt.Sprintf("\r%s%s\n\nAuthenticated with %s\n", cw.ClearLine, cw.Green("OK"), cw.Cyan(serviceName))

	// Print requested attributes
	if flagVals.Url_flag {
		authUrl := destination.(*auth.SwiftDestination).SwiftConnection.StorageUrl
		result += fmt.Sprintf("%s%s\n", cw.White("auth url: "), authUrl)
	}
	if flagVals.X_auth_flag {
		xAuth := destination.(*auth.SwiftDestination).SwiftConnection.AuthToken
		result += fmt.Sprintf("%s%s\n", cw.White("x-auth: "), xAuth)
	}

	return result, nil
}
