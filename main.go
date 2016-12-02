package main

import (
	"fmt"
	"os"
	"strings"

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
	execute         func(auth.Destination, []string) (string, error)
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
		// Help command
		//helpCommand: c.help,

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
		/*
			makeSLOCommand: c.makeSLO,
		*/
	}

	// Create writer to provide output
	c.writer = cw.NewConsoleWriter()

	// Dispatch the subcommand that the user wanted, if it exists
	var err error
	if len(args) < 2 {
		err = fmt.Errorf("Please provide a valid subcommand\nA list of subcommands can be found with the command 'cf help os'")
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

	splitTask := strings.Split(cmd.task, " ")
	curStage := strings.Join(splitTask[:len(splitTask)-1], " ")
	c.writer.SetCurrentStage(curStage)

	result, err := cmd.execute(destination, args)
	if err != nil {
		return err
	}

	c.writer.Quit()

	fmt.Print(result)

	return nil
}

// makeSLO creates a Static Large Object in an Object Storage instance.
func (c *LargeObjectsPlugin) makeSLO(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 5 {
		help, _ := getSubcommandHelp(makeSLOCommand)
		return fmt.Errorf("Missing required arguments\n%s", help)
	}

	// Parse arguments
	serviceName := args[1]
	argVals, err := slo.ParseArgs(args[2:])
	if err != nil {
		return fmt.Errorf("Failed to parse arguments: %s", err)
	}

	// Display startup info
	task := "Creating SLO in"
	err = displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	go c.writer.Write()

	// Authenticate with Object Storage
	destination, err := x_auth.Authenticate(cliConnection, c.writer, serviceName)
	if err != nil {
		return fmt.Errorf("Failed to authenticate: %s", err)
	}

	// Create SLO
	err = slo.MakeSlo(cliConnection, c.writer, destination, argVals)
	if err != nil {
		return fmt.Errorf("Failed to create SLO: %s", err)
	}

	// Kill console writer and display completion info
	c.writer.Quit()
	fmt.Printf("\r%s%s\n%s\nSuccessfully created SLO %s in container %s\n", cw.ClearLine, cw.Green("OK"), cw.ClearLine, cw.Cyan(argVals.SloName), cw.Cyan(argVals.SloContainer))

	return nil
}

// help prints help info for a given subcommand
func (c *LargeObjectsPlugin) help(cliConnection plugin.CliConnection, args []string) error {
	if len(args) < 2 {
		help := "Please provide a subcommand to fetch its help info\n" +
			"Available subcommands:\n" +
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
			"      " + makeSLOCommand + "\n"

		fmt.Print(help)

		return nil
	}

	help, err := getSubcommandHelp(args[1])
	if err != nil {
		return err
	}
	fmt.Print(help)

	return nil
}

func getSubcommandHelp(name string) (string, error) {
	var subcommandMap = map[string]int{
		getAuthInfoCommand:     0,
		showContainersCommand:  1,
		containerInfoCommand:   2,
		makeContainerCommand:   3,
		updateContainerCommand: 4,
		renameContainerCommand: 5,
		deleteContainerCommand: 6,
		showObjectsCommand:     7,
		objectInfoCommand:      8,
		putObjectCommand:       9,
		getObjectCommand:       10,
		renameObjectCommand:    11,
		copyObjectCommand:      12,
		deleteObjectCommand:    13,
		makeDLOCommand:         14,
		makeSLOCommand:         15,
	}

	var subcommands = []plugin.Command{
		{
			Name:     getAuthInfoCommand,
			HelpText: "Authenticate with Object Storage and save credentials",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + getAuthInfoCommand +
					" service_name [--url] [-x]",
				Options: map[string]string{
					"url": "Display auth url in quiet mode",
					"x":   "Display x-auth token in quiet mode",
				},
			},
		},
		{
			Name:     showContainersCommand,
			HelpText: "Show all containers in an Object Storage instance",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + showContainersCommand +
					" service_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     containerInfoCommand,
			HelpText: "Show a given container's information",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + containerInfoCommand +
					" service_name container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     makeContainerCommand,
			HelpText: "Create a new container in an Object Storage instance",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + makeContainerCommand +
					" service_name container_name [headers...] [r] [-r]",
				Options: map[string]string{
					"r":  "Short name for global read header",
					"-r": "Short name for remove read restrictions header",
				},
			},
		},
		{
			Name:     updateContainerCommand,
			HelpText: "Update a container's metadata",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + updateContainerCommand +
					" service_name container_name headers... [r] [-r]",
				Options: map[string]string{
					"r":  "Short name for global read header",
					"-r": "Short name for remove read restrictions header",
				},
			},
		},
		{
			Name:     renameContainerCommand,
			HelpText: "Rename a container",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + renameContainerCommand +
					" service_name container_name new_container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     deleteContainerCommand,
			HelpText: "Remove a container from an Object Storage instance",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + deleteContainerCommand +
					" service_name container_name [-f]",
				Options: map[string]string{
					"f": "Force delete even if not empty",
				},
			},
		},
		{
			Name:     showObjectsCommand,
			HelpText: "Show all objects in a container",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + showObjectsCommand +
					" service_name container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     objectInfoCommand,
			HelpText: "Show a given object's information",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + objectInfoCommand +
					" service_name container_name object_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     putObjectCommand,
			HelpText: "Upload a file as an object to Object Storage",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + putObjectCommand +
					" service_name container_name path_to_local_file [-n object_name]",
				Options: map[string]string{},
			},
		},
		{
			Name:     getObjectCommand,
			HelpText: "Download an object from Object Storage",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + getObjectCommand +
					" service_name container_name object_name path_to_dl_location",
				Options: map[string]string{},
			},
		},
		{
			Name:     renameObjectCommand,
			HelpText: "Rename an object",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + renameObjectCommand +
					" service_name container_name object_name new_object_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     copyObjectCommand,
			HelpText: "Copy an object to another container",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + copyObjectCommand +
					" service_name container_name object_name new_container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     deleteObjectCommand,
			HelpText: "Remove an object from a container",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + deleteObjectCommand +
					" service_name container_name object_name [-l]",
				Options: map[string]string{
					"l": "Delete all files associated with large object manifest object_name",
				},
			},
		},
		{
			Name:     makeDLOCommand,
			HelpText: "Create a Dynamic Large Object in Object Storage",
			UsageDetails: plugin.Usage{
				Usage: "cf " + namespace + " " + makeDLOCommand +
					" service_name dlo_container dlo_name [-c object_container] [-p dlo_prefix]",
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
				Usage: "cf " + namespace + " " + makeSLOCommand +
					" service_name slo_container slo_name source_file [-m] [-o output_file] [-s chunk_size] [-j num_threads]",
				Options: map[string]string{
					"m": "Only upload missing chunks",
					"o": "Destination for log data, if desired",
					"s": "Chunk size, in bytes (defaults to create 1GB chunks)",
					"j": "Maximum number of uploader threads (defaults to the available number of CPUs)",
				},
			},
		},
	}

	idx, found := subcommandMap[name]
	if !found {
		return "", fmt.Errorf("%s is not a valid subcommand", name)
	}

	help := ""
	help += fmt.Sprintf("NAME:\n\t%s - %s\n\nUSAGE:\n\t%s\n", subcommands[idx].Name, subcommands[idx].HelpText, subcommands[idx].UsageDetails.Usage)
	if len(subcommands[idx].UsageDetails.Options) > 0 {
		help += fmt.Sprintf("\nOPTIONS:\n")
		for k, v := range subcommands[idx].UsageDetails.Options {
			if name != makeContainerCommand && name != updateContainerCommand {
				help += fmt.Sprintf("\t-%s\t\t%v\n", k, v)
			} else {
				help += fmt.Sprintf("\t%s\t\t%v\n", k, v)
			}
		}
	}

	return help, nil
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
