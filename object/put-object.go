package object

import (
	"flag"
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/cloudfoundry/cli/plugin"
	cw "github.ibm.com/ckwaldon/cf-large-objects/console_writer"
	sg "github.ibm.com/ckwaldon/swiftlygo"
	"github.ibm.com/ckwaldon/swiftlygo/auth"
)

// flagVal holds the flag values
type flagVal struct {
	Rename_flag string
}

// parseArgs parses the arguments provided to put-object
func parseArgs(args []string) (*flagVal, error) {
	flagSet := flag.NewFlagSet("flagSet", flag.ContinueOnError)

	// Define flag to default to original filename
	rename := flagSet.String("n", fp.Base(args[1]), "Rename object before uploading")

	// Parse optional flags if they have been provided
	if len(args) > 2 {
		err := flagSet.Parse(args[2:])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse flags: %s", err)
		}
	}

	flagVals := flagVal{
		Rename_flag: string(*rename),
	}

	return &flagVals, nil
}

// PutObject uploads a file to Object Storage
func PutObject(cliConnection plugin.CliConnection, writer *cw.ConsoleWriter, dest auth.Destination, args []string) (string, error) {
	writer.SetCurrentStage("Uploading object")
	flags, err := parseArgs(args)
	if err != nil {
		return "", fmt.Errorf("Failed to parse arguments: %s", err)
	}

	// Verify that the source file exists
	file, err := os.Open(args[1])
	if err != nil {
		return "", fmt.Errorf("Failed to open source file: %s", err)
	}

	// Create uploader to upload object
	uploader := sg.NewObjectUploader(dest, file, args[0], flags.Rename_flag)
	err = uploader.Upload()
	if err != nil {
		return "", fmt.Errorf("Failed to upload object: %s", err)
	}

	return flags.Rename_flag, nil
}
