package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.com/ibmjstart/cf-object-storage/console_writer"
	"github.com/ibmjstart/cf-object-storage/container"
	"github.com/ibmjstart/cf-object-storage/dlo"
	"github.com/ibmjstart/cf-object-storage/object"
	"github.com/ibmjstart/cf-object-storage/slo"
	"github.com/ibmjstart/cf-object-storage/x_auth"
	"github.com/ibmjstart/swiftlygo/auth"
)

const (
	// Name of this plugin for use installing and uninstalling it
	pluginName string = "cf-object-storage"

	// Namespace for the plugin's subcommands
	namespace string = "os"

	// Name of the subcommand that provides help info for other subcommands
	helpCommand string = "help"

	// Name of the subcommand that fetches X-Auth Tokens
	getAuthInfoCommand string = "auth"

	// Names of the container subcommands
	showContainersCommand  string = "containers"
	containerInfoCommand   string = "container"
	makeContainerCommand   string = "create-container"
	updateContainerCommand string = "update-container"
	renameContainerCommand string = "rename-container"
	deleteContainerCommand string = "delete-container"

	// Names of the single object subcommands
	showObjectsCommand  string = "objects"
	objectInfoCommand   string = "object"
	putObjectCommand    string = "put-object"
	getObjectCommand    string = "get-object"
	renameObjectCommand string = "rename-object"
	copyObjectCommand   string = "copy-object"
	deleteObjectCommand string = "delete-object"

	// Names of the subcommands that create large objects in object storage
	makeDLOCommand string = "create-dynamic-object"
	makeSLOCommand string = "put-large-object"
)

// LargeObjectsPlugin is the struct implementing the plugin interface.
// It has no public members.
type LargeObjectsPlugin struct {
	subcommands   map[string](command)
	cliConnection plugin.CliConnection
	writer        *cw.ConsoleWriter
}

// command contains the info needed to execute a subcommand.
type command struct {
	name            string
	task            string
	numExpectedArgs int
	execute         func(auth.Destination, *cw.ConsoleWriter, []string) (string, error)
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

// Run handles each invocation of the CLI plugin.
func (c *LargeObjectsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Attach connection object to plugin struct
	c.cliConnection = cliConnection

	// Associate each subcommand with a handler function
	c.subcommands = map[string](command){
		// Authenticate command
		getAuthInfoCommand: command{
			name:            getAuthInfoCommand,
			task:            "Authenticating with",
			numExpectedArgs: 3,
			execute:         x_auth.DisplayAuthInfo,
		},

		// Container commands
		showContainersCommand: command{
			name:            showContainersCommand,
			task:            "Displaying containers in",
			numExpectedArgs: 3,
			execute:         container.ShowContainers,
		},
		containerInfoCommand: command{
			name:            containerInfoCommand,
			task:            "Fetching container info from",
			numExpectedArgs: 4,
			execute:         container.GetContainerInfo,
		},
		makeContainerCommand: command{
			name:            makeContainerCommand,
			task:            "Creating container in",
			numExpectedArgs: 4,
			execute:         container.MakeContainer,
		},
		updateContainerCommand: command{
			name:            updateContainerCommand,
			task:            "Updating container in",
			numExpectedArgs: 4,
			execute:         container.UpdateContainer,
		},
		renameContainerCommand: command{
			name:            renameContainerCommand,
			task:            "Renaming container in",
			numExpectedArgs: 5,
			execute:         container.RenameContainer,
		},
		deleteContainerCommand: command{
			name:            deleteContainerCommand,
			task:            "Removing container from",
			numExpectedArgs: 4,
			execute:         container.DeleteContainer,
		},

		// Object commands
		showObjectsCommand: command{
			name:            showObjectsCommand,
			task:            "Displaying objects in",
			numExpectedArgs: 4,
			execute:         object.ShowObjects,
		},
		objectInfoCommand: command{
			name:            objectInfoCommand,
			task:            "Fetching object info from",
			numExpectedArgs: 5,
			execute:         object.GetObjectInfo,
		},
		putObjectCommand: command{
			name:            putObjectCommand,
			task:            "Uploading object to",
			numExpectedArgs: 5,
			execute:         object.PutObject,
		},
		getObjectCommand: command{
			name:            getObjectCommand,
			task:            "Downloading object from",
			numExpectedArgs: 6,
			execute:         object.GetObject,
		},
		renameObjectCommand: command{
			name:            renameObjectCommand,
			task:            "Renaming object in",
			numExpectedArgs: 6,
			execute:         object.RenameObject,
		},
		copyObjectCommand: command{
			name:            copyObjectCommand,
			task:            "Copying object in",
			numExpectedArgs: 6,
			execute:         object.CopyObject,
		},
		deleteObjectCommand: command{
			name:            deleteObjectCommand,
			task:            "Removing object from",
			numExpectedArgs: 5,
			execute:         object.DeleteObject,
		},

		// Large object commands
		makeDLOCommand: command{
			name:            makeDLOCommand,
			task:            "Creating DLO in",
			numExpectedArgs: 5,
			execute:         dlo.MakeDlo,
		},
		makeSLOCommand: command{
			name:            makeSLOCommand,
			task:            "Creating SLO in",
			numExpectedArgs: 6,
			execute:         slo.MakeSlo,
		},
	}

	// Create writer to provide output
	c.writer = cw.NewConsoleWriter()

	// Dispatch the subcommand that the user wanted, if it exists
	var err error
	if len(args) < 2 || args[1] == helpCommand {
		err = c.help(args)
	} else {
		subcommand, found := c.subcommands[args[1]]
		if !found {
			err = fmt.Errorf("%s is not a valid subcommand", args[1])
		} else {
			err = c.executeCommand(subcommand, args)
		}
	}

	// Report any fatal errors returned by the subcommand
	if err != nil {
		fmt.Printf("\r%s\n%s\n%s\n", cw.ClearLine, cw.Red("FAILED"), err)
		os.Exit(1)
	}
}

func (c *LargeObjectsPlugin) executeCommand(cmd command, args []string) error {
	if len(args) < cmd.numExpectedArgs {
		help, _ := getSubcommandHelp(cmd.name)
		return fmt.Errorf("Missing required arguments\n%s", help)
	}

	err := displayUserInfo(c.cliConnection, cmd.task)
	if err != nil {
		return err
	}

	go c.writer.Write()

	serviceName := args[2]
	destination, err := x_auth.Authenticate(c.cliConnection, c.writer, serviceName)
	if err != nil {
		return err
	}

	result, err := cmd.execute(destination, c.writer, args)
	if err != nil {
		return err
	}

	c.writer.Quit()

	fmt.Print(result)

	return nil
}

// GetMetadata returns a PluginMetadata struct with information
// about the current version of this plugin and how to use it. This
// information is used to build the CF CLI helptext for this plugin's
// commands.
func (c *LargeObjectsPlugin) GetMetadata() plugin.PluginMetadata {
	var usageContent = "cf " + namespace + " COMMAND [ARGS...] \n" +
		"\n   Object Storage commands:\n" +
		"      " + getAuthInfoCommand + "\n" +
		"      " + showContainersCommand + "\n" +
		"      " + containerInfoCommand + "\n" +
		"      " + makeContainerCommand + "\n" +
		"      " + updateContainerCommand + "\n" +
		"      " + renameContainerCommand + "\n" +
		"      " + deleteContainerCommand + "\n" +
		"      " + showObjectsCommand + "\n" +
		"      " + objectInfoCommand + "\n" +
		"      " + putObjectCommand + "\n" +
		"      " + getObjectCommand + "\n" +
		"      " + renameObjectCommand + "\n" +
		"      " + copyObjectCommand + "\n" +
		"      " + deleteObjectCommand + "\n" +
		"      " + makeDLOCommand + "\n" +
		"      " + makeSLOCommand + "\n" +
		"   For more detailed information on subcommands use 'cf os help subcommand'"

	return plugin.PluginMetadata{
		Name: pluginName,
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 21,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     namespace,
				HelpText: "Work with SoftLayer Object Storage",
				UsageDetails: plugin.Usage{
					Usage: usageContent,
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
