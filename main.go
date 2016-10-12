package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	"github.ibm.com/ckwaldon/cf-large-objects/dlo"
	"github.ibm.com/ckwaldon/cf-large-objects/object"
	"github.ibm.com/ckwaldon/cf-large-objects/slo"
	"github.ibm.com/ckwaldon/cf-large-objects/x_auth"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

// pluginName defines the name of this plugin for use installing and
// uninstalling it.
const pluginName string = "ObjectStorageLargeObjects"

// getXAuthCommand defines the name of the command that fetches X-Auth Tokens.
const getAuthInfoCommand string = "get-auth-info"

// putObjectCommand defines the name of the command that uploads objects to
// object storage.
const putObjectCommand string = "put-object"

// makeDLOCommand defines the name of the command that creates DLOs in
// object storage.
const makeDLOCommand string = "make-dlo"

// makeDLOCommand defines the name of the command that creates SLOs in
// object storage.
const makeSLOCommand string = "make-slo"

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
		putObjectCommand:   c.putObject,
		makeDLOCommand:     c.makeDLO,
		makeSLOCommand:     c.makeSLO,
	}

	// Dispatch the subcommand that the user wanted, if it exists
	subcommandFunc := c.subcommands[args[0]]
	err := subcommandFunc(cliConnection, args)

	// Check for an error
	if err != nil {
		fmt.Printf("\r%s\n%s\n%s\n", cw.ClearLine, cw.Red("FAILED"), err)
		os.Exit(1)
	}
}

// displayUserInfo shows the username, org and space corresponding to the requested service.
func displayUserInfo(cliConnection plugin.CliConnection, task string) error {
	// Find username
	username, err := cliConnection.Username()
	if err != nil {
		return fmt.Errorf("Failed to get username: %s", err)
	}

	// Find org
	org, err := cliConnection.GetCurrentOrg()
	if err != nil {
		return fmt.Errorf("Failed to get organization: %s", err)
	}

	// Find space
	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return fmt.Errorf("Failed to get space: %s", err)
	}

	fmt.Printf("%s org %s / space %s as %s...\n", task, cw.Cyan(org.Name), cw.Cyan(space.Name), cw.Cyan(username))

	return nil
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
	writer := cw.NewConsoleWriter()

	// Start console writer if not in quiet mode
	if !quiet {
		task := "Fetching auth info from"

		err := displayUserInfo(cliConnection, task)
		if err != nil {
			return err
		}

		go writer.Write()
	}

	// Get authorization info
	destination, err := x_auth.GetAuthInfo(cliConnection, writer, args[1])
	if err != nil {
		return err
	}

	authUrl := destination.(*auth.SwiftDestination).SwiftConnection.StorageUrl
	xAuth := destination.(*auth.SwiftDestination).SwiftConnection.AuthToken

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

		fmt.Printf("\r%s%s\n\n%s\n%s%s\n%s%s\n", cw.ClearLine, cw.Green("OK"), cw.Cyan(args[1]), cw.White("auth url: "), authUrl, cw.White("x-auth:   "), xAuth)
	}

	return nil
}

// putObject executes the logic to upload an object to an object storage instance.
func (c *LargeObjectsPlugin) putObject(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 4 {
		return fmt.Errorf("Missing required arguments\nUsage: %s", c.GetMetadata().Commands[1].UsageDetails.Usage)
	}

	// Display startup info
	task := "Uploading object in"
	err := displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	writer := cw.NewConsoleWriter()
	go writer.Write()

	// Authenticate with Object Storage
	destination, err := x_auth.GetAuthInfo(cliConnection, writer, args[1])
	if err != nil {
		return fmt.Errorf("Failed to authenticate: %s", err)
	}

	// Upload object
	name, err := object.PutObject(cliConnection, writer, destination, args[2:])
	if err != nil {
		return fmt.Errorf("Failed to upload object: %s", err)
	}

	// Kill console writer and display completion info
	writer.Quit()
	fmt.Printf("\r%s%s\n\nUploaded %s to container %s\n", cw.ClearLine, cw.Green("OK"), cw.Cyan(name), cw.Cyan(args[2]))

	return nil
}

// makeDLO executes the logic to create a Dynamic Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeDLO(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 4 {
		return fmt.Errorf("Missing required arguments\nUsage: %s", c.GetMetadata().Commands[2].UsageDetails.Usage)
	}

	// Display startup info
	task := "Creating DLO in"
	err := displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	writer := cw.NewConsoleWriter()
	go writer.Write()

	// Authenticate with Object Storage
	destination, err := x_auth.GetAuthInfo(cliConnection, writer, args[1])
	if err != nil {
		return fmt.Errorf("Failed to authenticate: %s", err)
	}

	// Create DLO
	prefix, container, err := dlo.MakeDlo(cliConnection, writer, destination, args[2:])
	if err != nil {
		return fmt.Errorf("Failed to create DLO manifest: %s", err)
	}

	// Kill console writer and display completion info
	writer.Quit()
	fmt.Printf("\r%s%s\n\nCreated manifest for %s, upload segments to container %s prefixed with %s\n", cw.ClearLine, cw.Green("OK"), cw.Cyan(args[3]), cw.Cyan(container), cw.Cyan(prefix))

	return nil
}

// makeSLO executes the logic to create a Static Large Object in an object storage instance.
func (c *LargeObjectsPlugin) makeSLO(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 5 {
		return fmt.Errorf("Missing required arguments\nUsage: %s", c.GetMetadata().Commands[3].UsageDetails.Usage)
	}

	// Display startup info
	task := "Creating SLO in"
	err := displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	writer := cw.NewConsoleWriter()
	go writer.Write()

	// Authenticate with Object Storage
	destination, err := x_auth.GetAuthInfo(cliConnection, writer, args[1])
	if err != nil {
		return fmt.Errorf("Failed to authenticate: %s", err)
	}

	// Create SLO
	err = slo.MakeSlo(cliConnection, writer, destination, args[2:])
	if err != nil {
		return fmt.Errorf("Failed to create SLO: %s", err)
	}

	// Kill console writer and display completion info
	writer.Quit()
	fmt.Printf("\r%s%s\n%s\nSuccessfully created SLO %s in container %s\n", cw.ClearLine, cw.Green("OK"), cw.ClearLine, cw.Cyan(args[3]), cw.Cyan(args[2]))

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
					Usage: "cf " + getAuthInfoCommand + " service_name [--url] [-x]",
					Options: map[string]string{
						"url": "Display auth url in quiet mode",
						"x":   "Display x-auth token in quiet mode",
					},
				},
			},
			{
				Name:     putObjectCommand,
				HelpText: "Upload a file as an object to Object Storage",
				UsageDetails: plugin.Usage{
					Usage: "cf " + putObjectCommand + " service_name container_name path_to_source [-n object_name]",
					Options: map[string]string{
						"n": "Rename object before uploading",
					},
				},
			},
			{
				Name:     makeDLOCommand,
				HelpText: "Create a Dynamic Large Object in Object Storage",
				UsageDetails: plugin.Usage{
					Usage: "cf " + makeDLOCommand + " service_name dlo_container dlo_name [-c object_container] [-p dlo_prefix]",
					Options: map[string]string{
						"c": "Destination container for DLO segments (defaults to dlo_container)",
						"p": "Prefix to be used for DLO segments (default to dlo_name)",
					},
				},
			},
			{
				Name:     makeSLOCommand,
				HelpText: "Create a Static Large Object in Object Storage",
				UsageDetails: plugin.Usage{
					Usage: "cf " + makeSLOCommand + " service_name slo_container slo_name source_file [-m] [-o output_file] [-s chunk_size] [-j num_threads]",
					Options: map[string]string{
						"m": "Only upload missing chunks",
						"o": "Destination for log data, if desired",
						"s": "Chunk size, in bytes (defaults to create 1000 chunks)",
						"j": "Maximum number of uploader threads (defaults to the available number of CPUs)",
					},
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
