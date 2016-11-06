package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	"github.ibm.com/ckwaldon/cf-large-objects/container"
	"github.ibm.com/ckwaldon/cf-large-objects/dlo"
	"github.ibm.com/ckwaldon/cf-large-objects/object"
	"github.ibm.com/ckwaldon/cf-large-objects/slo"
	"github.ibm.com/ckwaldon/cf-large-objects/x_auth"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

const (
	// Name of this plugin for use installing and uninstalling it
	pluginName string = "cf-object-storage"

	// Namespace for the plugin's subcommands
	namespace string = "os"

	// Name of the subcommand that provides help info for other subcommands
	helpCommand string = "help"

	// Name of the subcommand that fetches X-Auth Tokens
	getAuthInfoCommand string = "get-auth-info"

	// Names of the container subcommands
	showContainersCommand  string = "containers"
	containerInfoCommand   string = "container-info"
	makeContainerCommand   string = "put-container"
	deleteContainerCommand string = "rm-container"

	// Names of the single object subcommands
	showObjectsCommand  string = "objects"
	objectInfoCommand   string = "object-info"
	putObjectCommand    string = "put-object"
	getObjectCommand    string = "get-object"
	deleteObjectCommand string = "rm-object"

	// Names of the subcommands that create large objects in object storage
	makeDLOCommand string = "make-dlo"
	makeSLOCommand string = "make-slo"
)

// LargeObjectsPlugin is the struct implementing the plugin interface.
// It has no public members.
type LargeObjectsPlugin struct {
	subcommands map[string](func(plugin.CliConnection, []string) error)
	writer      *cw.ConsoleWriter
}

// Run handles each invocation of the CLI plugin.
func (c *LargeObjectsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Associate each subcommand with a handler function
	c.subcommands = map[string](func(plugin.CliConnection, []string) error){
		helpCommand: c.help,

		getAuthInfoCommand: c.getAuthInfo,

		showContainersCommand:  c.containers,
		containerInfoCommand:   c.containers,
		makeContainerCommand:   c.containers,
		deleteContainerCommand: c.containers,

		showObjectsCommand:  c.objects,
		objectInfoCommand:   c.objects,
		putObjectCommand:    c.objects,
		getObjectCommand:    c.objects,
		deleteObjectCommand: c.objects,

		makeDLOCommand: c.makeDLO,
		makeSLOCommand: c.makeSLO,
	}

	// Create writer to provide output
	c.writer = cw.NewConsoleWriter()

	// Dispatch the subcommand that the user wanted, if it exists
	var err error
	if len(args) < 2 {
		err = fmt.Errorf("Please provide a valid subcommand\nA list of subcommands can be found with the command 'cf help os'")
	} else {
		subcommandFunc, found := c.subcommands[args[1]]
		if !found {
			err = fmt.Errorf("%s is not a valid subcommand", args[1])
		} else {
			err = subcommandFunc(cliConnection, args[1:])
		}
	}

	// Report any fatal errors returned by the subcommand
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

// getAuthInfo fetches the x-auth token and auth url for an Object Storage instance.
func (c *LargeObjectsPlugin) getAuthInfo(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 2 {
		return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", getAuthInfoCommand)
	}

	// Parse arguments
	serviceName := args[1]
	flags, err := x_auth.ParseFlags(args[2:])
	if err != nil {
		return err
	}

	quiet := flags.Url_flag || flags.X_auth_flag

	if !quiet {
		// Start console writer if not in quiet mode
		task := "Fetching auth info from"

		err := displayUserInfo(cliConnection, task)
		if err != nil {
			return err
		}

		go c.writer.Write()
	} else {
		// Clear any output that other processes generate
		go c.writer.ClearStatus()
	}

	// Get authorization info
	destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
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
		c.writer.Quit()

		fmt.Printf("\r%s%s\n\n%s\n%s%s\n%s%s\n", cw.ClearLine, cw.Green("OK"), cw.Cyan(serviceName),
			cw.White("auth url: "), authUrl, cw.White("x-auth:   "), xAuth)
	}

	return nil
}

// container executes the container commands
func (c *LargeObjectsPlugin) containers(cliConnection plugin.CliConnection, args []string) error {
	command := args[0]

	// Display startup info
	task := "Working with containers in"
	err := displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	go c.writer.Write()

	switch command {
	case showContainersCommand:
		if len(args) < 2 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", showContainersCommand)
		}

		serviceName := args[1]

		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containers, err := container.ShowContainers(destination)
		if err != nil {
			return fmt.Errorf("Failed to get containers: %s", err)
		}

		fmt.Printf("\r%s%s\n\nContainers in OS %s: %v\n", cw.ClearLine, cw.Green("OK"), serviceName, containers)
	case containerInfoCommand:
		if len(args) < 3 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", containerInfoCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerArg := args[2]
		containerInfo, headers, err := container.GetContainerInfo(destination, containerArg)
		if err != nil {
			return fmt.Errorf("Failed to get container %s: %s", containerArg, err)
		}

		fmt.Printf("\r%s%s\n\nName: %s\nNumber of objects: %d\nSize: %d bytes\nHeaders:", cw.ClearLine, cw.Green("OK"), containerInfo.Name, containerInfo.Count, containerInfo.Bytes)
		for k, h := range headers {
			fmt.Printf("\n\tName: %s Value: %s", k, h)
		}
		fmt.Printf("\n")
	case makeContainerCommand:
		if len(args) < 3 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", makeContainerCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerArg := args[2]
		headersArg := args[3:]
		err = container.MakeContainer(destination, containerArg, headersArg...)
		if err != nil {
			return fmt.Errorf("Failed to make container: %s", err)
		}

		fmt.Printf("\r%s%s\n\nCreated container %s in OS %s\n", cw.ClearLine, cw.Green("OK"), containerArg, serviceName)
	case deleteContainerCommand:
		if len(args) < 3 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", deleteContainerCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerArg := args[2]
		err = container.DeleteContainer(destination, containerArg)
		if err != nil {
			return fmt.Errorf("Failed to delete container: %s", err)
		}

		fmt.Printf("\r%s%s\n\nDeleted container %s from OS %s\n", cw.ClearLine, cw.Green("OK"), containerArg, serviceName)
	}

	// Kill console writer
	c.writer.Quit()

	return nil
}

// object executes single object commands
func (c *LargeObjectsPlugin) objects(cliConnection plugin.CliConnection, args []string) error {
	command := args[0]

	// Display startup info
	task := "Working with objects in"
	err := displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	go c.writer.Write()

	switch command {
	case showObjectsCommand:
		if len(args) < 3 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", showObjectsCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerName := args[2]
		objects, err := object.ShowObjects(destination, containerName)
		if err != nil {
			return fmt.Errorf("Failed to get objects: %s", err)
		}

		fmt.Printf("\r%s%s\n\nObjects in container %s: %v\n", cw.ClearLine, cw.Green("OK"), containerName, objects)
	case objectInfoCommand:
		if len(args) < 4 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", objectInfoCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerName := args[2]
		objectArg := args[3]
		objectInfo, headers, err := object.GetObjectInfo(destination, containerName, objectArg)
		if err != nil {
			return fmt.Errorf("Failed to get object %s: %s", objectArg, err)
		}

		fmt.Printf("\r%s%s\n\nName: %s\nContent type: %s\nSize: %d bytes\nLast modified: %s\nHash: %s\nIs pseudo dir: %t\nSubdirectory: \n%sHeaders:", cw.ClearLine, cw.Green("OK"), objectInfo.Name, objectInfo.ContentType, objectInfo.Bytes, objectInfo.ServerLastModified, objectInfo.Hash, objectInfo.PseudoDirectory, objectInfo.SubDir)
		for k, h := range headers {
			fmt.Printf("\n\tName: %s Value: %s", k, h)
		}
		fmt.Printf("\n")
	case putObjectCommand:
		if len(args) < 5 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", putObjectCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerName := args[2]
		objectArg := args[3]
		path := args[4]
		err = object.PutObject(destination, containerName, objectArg, path)
		if err != nil {
			return fmt.Errorf("Failed to upload object: %s", err)
		}

		fmt.Printf("\r%s%s\n\nUploaded object %s to container %s\n", cw.ClearLine, cw.Green("OK"), objectArg, containerName)
	case getObjectCommand:
		if len(args) < 5 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", getObjectCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerName := args[2]
		objectArg := args[3]
		destinationPath := args[4]
		err = object.GetObject(destination, containerName, objectArg, destinationPath)
		if err != nil {
			return fmt.Errorf("Failed to upload object: %s", err)
		}

		fmt.Printf("\r%s%s\n\nDownloaded object %s to %s\n", cw.ClearLine, cw.Green("OK"), objectArg, destinationPath)
	case deleteObjectCommand:
		if len(args) < 4 {
			return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", deleteObjectCommand)
		}

		serviceName := args[1]
		destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
		if err != nil {
			return fmt.Errorf("Failed to authenticate: %s", err)
		}

		containerName := args[2]
		objectArg := args[3]
		err = object.DeleteObject(destination, containerName, objectArg)
		if err != nil {
			return fmt.Errorf("Failed to delete object %s: %s", objectArg, err)
		}

		fmt.Printf("\r%s%s\n\nDeleted object %s from container %s\n", cw.ClearLine, cw.Green("OK"), objectArg, containerName)
	}

	// Kill console writer
	c.writer.Quit()

	return nil
}

// makeDLO creates a Dynamic Large Object manifest in an Object Storage instance.
func (c *LargeObjectsPlugin) makeDLO(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 4 {
		return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", makeDLOCommand)
	}

	// Parse arguments
	serviceName := args[1]
	argVals, err := dlo.ParseArgs(args[2:])
	if err != nil {
		return fmt.Errorf("Failed to parse arguments: %s", err)
	}

	// Display startup info
	task := "Creating DLO in"
	err = displayUserInfo(cliConnection, task)
	if err != nil {
		return fmt.Errorf("Failed to display user info: %s", err)
	}

	// Start console writer
	go c.writer.Write()

	// Authenticate with Object Storage
	destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
	if err != nil {
		return fmt.Errorf("Failed to authenticate: %s", err)
	}

	// Create DLO
	err = dlo.MakeDlo(cliConnection, c.writer, destination, argVals)
	if err != nil {
		return fmt.Errorf("Failed to create DLO manifest: %s", err)
	}

	// Kill console writer and display completion info
	c.writer.Quit()
	fmt.Printf("\r%s%s\n\nCreated manifest for %s, upload segments to container %s prefixed with %s\n", cw.ClearLine, cw.Green("OK"), cw.Cyan(argVals.DloName), cw.Cyan(argVals.FlagVals.Container_flag), cw.Cyan(argVals.FlagVals.Prefix_flag))

	return nil
}

// makeSLO creates a Static Large Object in an Object Storage instance.
func (c *LargeObjectsPlugin) makeSLO(cliConnection plugin.CliConnection, args []string) error {
	// Check that the minimum number of arguments are present
	if len(args) < 5 {
		return fmt.Errorf("Missing required arguments\nSee 'cf os help %s' for details", makeSLOCommand)
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
	destination, err := x_auth.GetAuthInfo(cliConnection, c.writer, serviceName)
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
		return fmt.Errorf("Must provide subcommand to fetch help info for")
	}

	var subcommands = []plugin.Command{
		{
			Name:     getAuthInfoCommand,
			HelpText: "Display an Object Storage service's authentication url and x-auth token",
			UsageDetails: plugin.Usage{
				Usage: "cf " + getAuthInfoCommand +
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
				Usage: "cf " + showContainersCommand +
					" service_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     containerInfoCommand,
			HelpText: "Show a given container's information",
			UsageDetails: plugin.Usage{
				Usage: "cf " + containerInfoCommand +
					" service_name container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     makeContainerCommand,
			HelpText: "Create a new container in an Object Storage instance",
			UsageDetails: plugin.Usage{
				Usage: "cf " + makeContainerCommand +
					" service_name container_name [headers...] [r] [-r]",
				Options: map[string]string{},
			},
		},
		{
			Name:     deleteContainerCommand,
			HelpText: "Remove a container from an Object Storage instance",
			UsageDetails: plugin.Usage{
				Usage: "cf " + deleteContainerCommand +
					" service_name container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     showObjectsCommand,
			HelpText: "Show all objects in a container",
			UsageDetails: plugin.Usage{
				Usage: "cf " + showObjectsCommand +
					" service_name container_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     objectInfoCommand,
			HelpText: "Show a given object's information",
			UsageDetails: plugin.Usage{
				Usage: "cf " + objectInfoCommand +
					" service_name container_name object_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     putObjectCommand,
			HelpText: "Upload a file as an object to Object Storage",
			UsageDetails: plugin.Usage{
				Usage: "cf " + putObjectCommand +
					" service_name container_name object_name path_to_local_file",
				Options: map[string]string{},
			},
		},
		{
			Name:     getObjectCommand,
			HelpText: "Download an object from Object Storage",
			UsageDetails: plugin.Usage{
				Usage: "cf " + getObjectCommand +
					" service_name container_name object_name path_to_dl_location",
				Options: map[string]string{},
			},
		},
		{
			Name:     deleteObjectCommand,
			HelpText: "Remove an object from a container",
			UsageDetails: plugin.Usage{
				Usage: "cf " + deleteObjectCommand +
					" service_name container_name object_name",
				Options: map[string]string{},
			},
		},
		{
			Name:     makeDLOCommand,
			HelpText: "Create a Dynamic Large Object in Object Storage",
			UsageDetails: plugin.Usage{
				Usage: "cf " + makeDLOCommand +
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
				Usage: "cf " + makeSLOCommand +
					" service_name slo_container slo_name source_file [-m] [-o output_file] [-s chunk_size] [-j num_threads]",
				Options: map[string]string{
					"m": "Only upload missing chunks",
					"o": "Destination for log data, if desired",
					"s": "Chunk size, in bytes (defaults to create 1000 chunks)",
					"j": "Maximum number of uploader threads (defaults to the available number of CPUs)",
				},
			},
		},
	}

	var subcommandMap = map[string]int{
		getAuthInfoCommand:     0,
		showContainersCommand:  1,
		containerInfoCommand:   2,
		makeContainerCommand:   3,
		deleteContainerCommand: 4,
		showObjectsCommand:     5,
		objectInfoCommand:      6,
		putObjectCommand:       7,
		getObjectCommand:       8,
		deleteObjectCommand:    9,
		makeDLOCommand:         10,
		makeSLOCommand:         11,
	}

	idx, found := subcommandMap[args[1]]
	if !found {
		return fmt.Errorf("%s is not a valid subcommand", args[1])
	}

	fmt.Printf("NAME:\n\t%s - %s\n\nUSAGE:\n\t%s\n", subcommands[idx].Name, subcommands[idx].HelpText, subcommands[idx].UsageDetails.Usage)
	if len(subcommands[idx].UsageDetails.Options) > 0 {
		fmt.Printf("\nOPTIONS:\n")
		for k, v := range subcommands[idx].UsageDetails.Options {
			fmt.Printf("\t-%s\t\t%v\n", k, v)
		}
	}

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
		"      " + deleteContainerCommand + "\n" +
		"      " + showObjectsCommand + "\n" +
		"      " + objectInfoCommand + "\n" +
		"      " + putObjectCommand + "\n" +
		"      " + getObjectCommand + "\n" +
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
