package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

var (
	subcommands = []plugin.Command{
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

	subcommandMap = map[string]plugin.Command{
		getAuthInfoCommand:     subcommands[0],
		showContainersCommand:  subcommands[1],
		containerInfoCommand:   subcommands[2],
		makeContainerCommand:   subcommands[3],
		updateContainerCommand: subcommands[4],
		renameContainerCommand: subcommands[5],
		deleteContainerCommand: subcommands[6],
		showObjectsCommand:     subcommands[7],
		objectInfoCommand:      subcommands[8],
		putObjectCommand:       subcommands[9],
		getObjectCommand:       subcommands[10],
		renameObjectCommand:    subcommands[11],
		copyObjectCommand:      subcommands[12],
		deleteObjectCommand:    subcommands[13],
		makeDLOCommand:         subcommands[14],
		makeSLOCommand:         subcommands[15],
	}
)

func toString(s plugin.Command) string {
	help := ""

	help += fmt.Sprintf("NAME:\n\t%s - %s\n\nUSAGE:\n\t%s\n", s.Name, s.HelpText, s.UsageDetails.Usage)
	if len(s.UsageDetails.Options) > 0 {
		help += fmt.Sprintf("\nOPTIONS:\n")
		for k, v := range s.UsageDetails.Options {
			if s.Name != makeContainerCommand && s.Name != updateContainerCommand {
				help += fmt.Sprintf("\t-%s\t\t%v\n", k, v)
			} else {
				help += fmt.Sprintf("\t%s\t\t%v\n", k, v)
			}
		}
	}

	return help
}

func getSubcommandHelp(name string) (string, error) {
	subcommand, found := subcommandMap[name]
	if !found {
		return "", fmt.Errorf("%s is not a valid subcommand", name)
	}

	return toString(subcommand), nil
}

// help prints help info for a given subcommand
func (c *ObjectStoragePlugin) help(args []string) error {
	if len(args) < 3 {
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

	help, err := getSubcommandHelp(args[2])
	if err != nil {
		return err
	}
	fmt.Print(help)

	return nil
}
