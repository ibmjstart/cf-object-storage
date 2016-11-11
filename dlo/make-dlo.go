package dlo

import (
	"flag"
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.com/ibmjstart/cf-large-objects/console_writer"
	sg "github.com/ibmjstart/swiftlygo"
	"github.com/ibmjstart/swiftlygo/auth"
)

// argVal holds the parsed argument values.
type argVal struct {
	dloContainer string
	DloName      string
	FlagVals     *flagVal
}

// flagVal holds the flag values.
type flagVal struct {
	Container_flag string
	Prefix_flag    string
}

// ParseArgs parses the arguments provided to make-dlo.
func ParseArgs(args []string) (*argVal, error) {
	dloContainer := args[0]
	dloName := args[1]

	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flags to default to matching required arguments
	container := flagSet.String("c", dloContainer, "Destination container for DLO segments (defaults to manifest container)")
	prefix := flagSet.String("p", dloName, "Prefix to be used for DLO segments (defaults to DLO name)")

	// Parse optional flags if they have been provided
	if len(args) > 2 {
		err := flagSet.Parse(args[2:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		Container_flag: string(*container),
		Prefix_flag:    string(*prefix),
	}

	argVals := argVal{
		dloContainer: dloContainer,
		DloName:      dloName,
		FlagVals:     &flagVals,
	}

	return &argVals, nil
}

// MakeDlo uploads a DLO manifest to Object Storage.
func MakeDlo(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, dest auth.Destination, argVals *argVal) error {
	writer.SetCurrentStage("Preparing DLO manifest")

	// Create uploader to build manifest
	writer.SetCurrentStage("Uploading DLO manifest")
	uploader := sg.NewDloManifestUploader(dest, argVals.dloContainer, argVals.DloName, argVals.FlagVals.Container_flag, argVals.FlagVals.Prefix_flag)
	err := uploader.Upload()
	if err != nil {
		return fmt.Errorf("Failed to upload DLO manifest: %s", err)
	}

	return nil
}
